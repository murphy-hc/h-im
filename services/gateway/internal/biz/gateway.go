package biz

import (
	"context"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/gorilla/websocket"
)

// GatewayUseCase handles gateway business logic.
type GatewayUseCase struct {
	cm ConnManager
}

// NewGatewayUseCase creates a GatewayUseCase.
func NewGatewayUseCase(cm ConnManager) *GatewayUseCase {
	return &GatewayUseCase{cm: cm}
}

// HandleConnection processes a WebSocket connection.
func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID string) {
	defer uc.cm.Remove(userID)

	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
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
			// respond with heartbeat echo
			frame, _ := Encode(CurrentVersion, gatewayv1.FrameType_FRAME_TYPE_HEARTBEAT, nil)
			conn.WriteMessage(websocket.BinaryMessage, frame)
		default:
			// For client-to-server messages, route to backend via gRPC (future tasks)
			_ = payload
			_ = ft
		}
	}
}
