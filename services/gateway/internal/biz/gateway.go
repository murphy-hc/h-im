package biz

import (
	"context"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	msgpb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
)

const readTimeout = 60 * time.Second

type GatewayUseCase struct {
	cm        ConnManager
	msgClient MessageClient
}

func NewGatewayUseCase(cm ConnManager, msgClient MessageClient) *GatewayUseCase {
	return &GatewayUseCase{cm: cm, msgClient: msgClient}
}

func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID, deviceID string) {
	defer uc.cm.Remove(userID, deviceID)

	done := make(chan struct{})
	defer close(done)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done: return
			case <-ticker.C:
				conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(5*time.Second))
			}
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		return nil
	})

	for {
		conn.SetReadDeadline(time.Now().Add(readTimeout))
		_, raw, err := conn.ReadMessage()
		if err != nil { break }
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
			conn.WriteMessage(websocket.BinaryMessage, frame)
		case gatewayv1.FrameType_FRAME_TYPE_PRIVATE_CHAT:
			uc.handlePrivateChat(ctx, conn, userID, payload)
		case gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK:
			var ack msgpb.PrivateAck
			if err := proto.Unmarshal(payload, &ack); err == nil {
				go uc.msgClient.AckMessage(ctx, ack.MsgServerId, userID)
			}
		}
	}
}

func (uc *GatewayUseCase) handlePrivateChat(ctx context.Context, conn *websocket.Conn, senderID string, payload []byte) {
	var msg msgpb.Message
	if err := proto.Unmarshal(payload, &msg); err != nil { return }
	resp, err := uc.msgClient.SendMessage(ctx, &msgpb.SendMessageReq{
		SenderId:        senderID,
		ReceiverId:      msg.ReceiverId,
		ConvType:        msgpb.ConversationType_CONVERSATION_PRIVATE,
		MsgType:         msg.MsgType,
		Text:            msg.Text,
		MessageClientId: msg.MessageClientId,
	})
	if err != nil { return }
	ack := &msgpb.PrivateAck{
		MsgServerId: resp.MessageServerId,
		MsgClientId: msg.MessageClientId,
		Status:      msgpb.AckStatus_ACK_SENT,
	}
	frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_PRIVATE_ACK, ack)
	conn.WriteMessage(websocket.BinaryMessage, frame)
}
