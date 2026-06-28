package biz

import (
	"context"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/coder/websocket"
	"github.com/murphy-hc/h-im/pkg/gp"
	"google.golang.org/protobuf/proto"
)

const (
	readTimeout  = 60 * time.Second
	writeTimeout = 5 * time.Second
	pingTimeout  = 5 * time.Second
)

type HeartbeatConfig struct {
	IntervalSeconds int32
	TimeoutSeconds  int32
	SweepInterval   int32
}

func (c HeartbeatConfig) Timeout() time.Duration {
	if c.TimeoutSeconds <= 0 {
		return 180 * time.Second
	}
	return time.Duration(c.TimeoutSeconds) * time.Second
}

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
	chatroomSvc ChatroomClient
	groupSvc    GroupClient
	hbCfg       HeartbeatConfig
	gatewayAddr string
}

func NewGatewayUseCase(cm ConnManager, msgClient MessageClient, userStatus UserStatusClient, chatroomSvc ChatroomClient, groupSvc GroupClient, hbCfg HeartbeatConfig, gatewayAddr string) *GatewayUseCase {
	uc := &GatewayUseCase{
		cm: cm, msgClient: msgClient, userStatus: userStatus, chatroomSvc: chatroomSvc, groupSvc: groupSvc,
		hbCfg: hbCfg, gatewayAddr: gatewayAddr,
	}
	gp.SafeGo(context.Background(), func(_ context.Context) { uc.sweepLoop() })
	return uc
}

func writeCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), writeTimeout)
}

func (uc *GatewayUseCase) ValidateToken(ctx context.Context, appID, userID, token string) (bool, error) {
	return uc.userStatus.ValidateAppToken(ctx, appID, userID, token)
}

func (uc *GatewayUseCase) sweepLoop() {
	ticker := time.NewTicker(uc.hbCfg.SweepDuration())
	defer ticker.Stop()
	timeout := uc.hbCfg.Timeout()
	for range ticker.C {
		offline := uc.cm.SweepOffline(timeout)
		for _, dev := range offline {
			dev.Conn.Close(websocket.StatusNormalClosure, CloseReasonHeartbeatTimeout)
			gp.SafeGo(context.Background(), func(_ context.Context) {
				uc.userStatus.ReportDisconnect(context.Background(), dev.UserID, dev.DeviceID)
			})
		}
	}
}

func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID, deviceID string) {
	defer func() {
		uc.cm.Remove(userID, deviceID)
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
				pingCtx, cancel := context.WithTimeout(context.Background(), pingTimeout)
				conn.Ping(pingCtx)
				cancel()
			}
		}
	})

	for {
		readCtx, cancel := context.WithTimeout(ctx, readTimeout)
		_, raw, err := conn.Read(readCtx)
		cancel()
		if err != nil {
			break
		}
		version, ft, payload, err := Decode(raw)
		if err != nil {
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
				&gatewayv1.ErrorMessage{Code: 1, Message: err.Error()})
			wc, wcCancel := writeCtx()
			conn.Write(wc, websocket.MessageBinary, frame)
			wcCancel()
			continue
		}
		_ = version
		switch ft {
		case gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT:
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
			wc, wcCancel := writeCtx()
			err := conn.Write(wc, websocket.MessageBinary, frame)
			wcCancel()
			if err != nil {
				uc.cm.MarkHeartbeatFail(userID, deviceID)
			} else {
				uc.cm.MarkHeartbeatSuccess(userID, deviceID)
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
		case gatewayv1.FrameType_FRAME_TYPE_CHATROOM_MSG:
			uc.handleChatroomMsg(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_CHATROOM_ACK:
			var ack msgpb.PrivateAck
			if err := proto.Unmarshal(payload, &ack); err == nil {
				gp.SafeGo(ctx, func(_ context.Context) { uc.msgClient.AckMessage(ctx, ack.MsgServerId, userID) })
			}
		case gatewayv1.FrameType_FRAME_TYPE_CHATROOM_JOIN:
			uc.handleChatroomJoin(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_CHATROOM_LEAVE:
			uc.handleChatroomLeave(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_GROUP_CHAT:
			uc.handleGroupChat(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_GROUP_ACK:
			var ack msgpb.PrivateAck
			if err := proto.Unmarshal(payload, &ack); err == nil {
				gp.SafeGo(ctx, func(_ context.Context) { uc.msgClient.AckMessage(ctx, ack.MsgServerId, userID) })
			}
		case gatewayv1.FrameType_FRAME_TYPE_GROUP_JOIN:
			uc.handleGroupJoin(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_GROUP_LEAVE:
			uc.handleGroupLeave(ctx, conn, userID, payload)
		}
	}
}

func (uc *GatewayUseCase) handleChatroomMsg(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var msg msgpb.Message
	if err := proto.Unmarshal(payload, &msg); err != nil {
		return
	}
	resp, err := uc.msgClient.SendMessage(ctx, &msgpb.SendMessageReq{
		SenderId:        senderID,
		ReceiverId:      msg.ReceiverId,
		ConvType:        msgpb.ConversationType_CONVERSATION_ROOM,
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
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_CHATROOM_ACK, ack)
	wc, wcCancel := writeCtx()
	conn.Write(wc, websocket.MessageBinary, frame)
	wcCancel()
}

func (uc *GatewayUseCase) handleRecallMsg(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var req msgpb.RecallMessageReq
	if err := proto.Unmarshal(payload, &req); err != nil {
		frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
			&gatewayv1.ErrorMessage{Code: 1, Message: "invalid recall request"})
		wc, wcCancel := writeCtx()
		conn.Write(wc, websocket.MessageBinary, frame)
		wcCancel()
		return
	}
	req.SenderId = senderID

	if err := uc.msgClient.RecallMessage(ctx, &req); err != nil {
		frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_ERROR,
			&gatewayv1.ErrorMessage{Code: 2, Message: err.Error()})
		wc, wcCancel := writeCtx()
		conn.Write(wc, websocket.MessageBinary, frame)
		wcCancel()
		return
	}

	ack := &msgpb.PrivateAck{
		MsgServerId: req.MessageServerId,
		Status:      msgpb.AckStatus_ACK_RECALLED,
	}
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK, ack)
	wc, wcCancel := writeCtx()
	conn.Write(wc, websocket.MessageBinary, frame)
	wcCancel()
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
	wc, wcCancel := writeCtx()
	conn.Write(wc, websocket.MessageBinary, frame)
	wcCancel()
}

func (uc *GatewayUseCase) handleChatroomJoin(ctx context.Context, _ *websocket.Conn, userID string, payload []byte) {
	roomID := string(payload)
	if roomID == "" {
		return
	}
	if err := uc.chatroomSvc.JoinRoom(ctx, roomID, userID); err != nil {
		return
	}
	uc.cm.JoinRoom(roomID, userID)
}

func (uc *GatewayUseCase) handleChatroomLeave(ctx context.Context, _ *websocket.Conn, userID string, payload []byte) {
	roomID := string(payload)
	if roomID == "" {
		return
	}
	if err := uc.chatroomSvc.LeaveRoom(ctx, roomID, userID); err != nil {
		return
	}
	uc.cm.LeaveRoom(roomID, userID)
}

func (uc *GatewayUseCase) handleGroupChat(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var msg msgpb.Message
	if err := proto.Unmarshal(payload, &msg); err != nil {
		return
	}
	resp, err := uc.msgClient.SendMessage(ctx, &msgpb.SendMessageReq{
		SenderId: senderID, ReceiverId: msg.ReceiverId,
		ConvType: msgpb.ConversationType_CONVERSATION_GROUP,
		MsgType: msg.MsgType, Text: msg.Text,
		MessageClientId: msg.MessageClientId,
		Attachment: msg.Attachment,
	})
	if err != nil {
		return
	}
	ack := &msgpb.PrivateAck{
		MsgServerId: resp.MessageServerId,
		MsgClientId: msg.MessageClientId,
		Status:      msgpb.AckStatus_ACK_SENT,
	}
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_GROUP_ACK, ack)
	wc, wcCancel := writeCtx()
	conn.Write(wc, websocket.MessageBinary, frame)
	wcCancel()
}

func (uc *GatewayUseCase) handleGroupJoin(ctx context.Context, _ *websocket.Conn, userID string, payload []byte) {
	groupID := string(payload)
	if groupID == "" { return }
	if err := uc.groupSvc.JoinGroup(ctx, groupID, userID); err != nil { return }
	uc.cm.JoinGroup(groupID, userID)
}

func (uc *GatewayUseCase) handleGroupLeave(ctx context.Context, _ *websocket.Conn, userID string, payload []byte) {
	groupID := string(payload)
	if groupID == "" { return }
	if err := uc.groupSvc.LeaveGroup(ctx, groupID, userID); err != nil { return }
	uc.cm.LeaveGroup(groupID, userID)
}
