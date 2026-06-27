package biz

import (
	"context"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/pkg/gp"
	"google.golang.org/protobuf/proto"
)

const readTimeout = 60 * time.Second

// HeartbeatConfig holds heartbeat parameters.
type HeartbeatConfig struct {
	IntervalSeconds int32
	TimeoutSeconds  int32
	SweepInterval   int32
}

// Timeout returns the heartbeat timeout duration.
func (c HeartbeatConfig) Timeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return 180 * time.Second
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

// SweepDuration returns the sweep interval duration.
func (c HeartbeatConfig) SweepDuration() time.Duration {
	if c.SweepInterval <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.SweepInterval) * time.Second
}

type GatewayUseCase struct {
	cm          ConnManager
	msgClient   MessageClient
	userStatus  UserStatusClient
	hbCfg       HeartbeatConfig
	gatewayAddr string
}

func NewGatewayUseCase(cm ConnManager, msgClient MessageClient, userStatus UserStatusClient, hbCfg HeartbeatConfig, gatewayAddr string) *GatewayUseCase {
	uc := &GatewayUseCase{
		cm:          cm,
		msgClient:   msgClient,
		userStatus:  userStatus,
		hbCfg:       hbCfg,
		gatewayAddr: gatewayAddr,
	}
	gp.SafeGo(context.Background(), func(_ context.Context) { uc.sweepLoop() })
	return uc
}

// ValidateToken validates a client token via the user service.
func (uc *GatewayUseCase) ValidateToken(ctx context.Context, appID, userID, token string) (bool, error) {
	return uc.userStatus.ValidateAppToken(ctx, appID, userID, token)
}

// sweepLoop periodically scans connections and closes those that have timed out.
// It also reports disconnects to the user service.
func (uc *GatewayUseCase) sweepLoop() {
	ticker := time.NewTicker(uc.hbCfg.SweepDuration())
	defer ticker.Stop()
	timeout := uc.hbCfg.Timeout()
	for range ticker.C {
		offline := uc.cm.SweepOffline(timeout)
		for _, dev := range offline {
			dev.Conn.Close()
			// Report disconnect to user service (best-effort, don't block)
			gp.SafeGo(context.Background(), func(_ context.Context) {
				uc.userStatus.ReportDisconnect(context.Background(), dev.UserID, dev.DeviceID)
			})
		}
	}
}

func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID, deviceID string) {
	defer func() {
		uc.cm.Remove(userID, deviceID)
		// Report disconnect to user service
		gp.SafeGo(context.Background(), func(_ context.Context) {
			uc.userStatus.ReportDisconnect(context.Background(), userID, deviceID)
		})
	}()

	done := make(chan struct{})
	defer close(done)
	gp.SafeGo(context.Background(), func(_ context.Context) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			}
		}
	})

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			break
		}
		version, ft, payload, err := Decode(raw)
		if err != nil {
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
				&gatewayv1.ErrorMessage{Code: 1, Message: err.Error()})
			conn.WriteMessage(websocket.BinaryMessage, frame)
			continue
		}
		_ = version
		switch ft {
		case gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT:
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
			if err := conn.WriteMessage(websocket.BinaryMessage, frame); err != nil {
				uc.cm.MarkHeartbeatFail(userID, deviceID)
			} else {
				uc.cm.MarkHeartbeatSuccess(userID, deviceID)
				// Forward successful heartbeat to user service
				gp.SafeGo(context.Background(), func(_ context.Context) {
					uc.userStatus.ReportHeartbeat(context.Background(), userID, deviceID, uc.gatewayAddr, time.Now().Unix())
				})
			}
		case gatewayv1.FrameType_FRAME_TYPE_PRIVATE_CHAT:
			uc.handlePrivateChat(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK:
			var ack msgpb.PrivateAck
			if err := proto.Unmarshal(payload, &ack); err == nil {
				gp.SafeGo(ctx, func(_ context.Context) { uc.msgClient.AckMessage(ctx, ack.MsgServerId, userID) })
			}
		case gatewayv1.FrameType_FRAME_TYPE_PRIVATE_RECALL:
			uc.handleRecallMsg(ctx, conn, userID, payload)
		}
	}
}

func (uc *GatewayUseCase) handleRecallMsg(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var req msgpb.RecallMessageReq
	if err := proto.Unmarshal(payload, &req); err != nil {
		frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
			&gatewayv1.ErrorMessage{Code: 1, Message: "invalid recall request"})
		conn.WriteMessage(websocket.BinaryMessage, frame)
		return
	}
	req.SenderId = senderID

	if err := uc.msgClient.RecallMessage(ctx, &req); err != nil {
		frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
			&gatewayv1.ErrorMessage{Code: 2, Message: err.Error()})
		conn.WriteMessage(websocket.BinaryMessage, frame)
		return
	}

	ack := &msgpb.PrivateAck{
		MsgServerId: req.MessageServerId,
		Status:      msgpb.AckStatus_ACK_RECALLED,
	}
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK, ack)
	conn.WriteMessage(websocket.BinaryMessage, frame)
}

func (uc *GatewayUseCase) handlePrivateChat(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var msg msgpb.Message
	if err := proto.Unmarshal(payload, &msg); err != nil {
		return
	}
	resp, err := uc.msgClient.SendMessage(ctx, &msgpb.SendMessageReq{
		SenderId:        senderID,
		ReceiverId:      msg.ReceiverId,
		ConvType:        msgpb.ConversationType_CONVERSATION_PRIVATE,
		MsgType:         msg.MsgType,
		Text:            msg.Text,
		MessageClientId: msg.MessageClientId,
		Attachment:      msg.Attachment,
	})
	if err != nil {
		return
	}
	ack := &msgpb.PrivateAck{
		MsgServerId: resp.MessageServerId,
		MsgClientId: msg.MessageClientId,
		Status:      msgpb.AckStatus_ACK_SENT,
	}
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK, ack)
	conn.WriteMessage(websocket.BinaryMessage, frame)
}
