package biz

import (
	"context"
	"time"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/gorilla/websocket"
)

const readTimeout = 60 * time.Second

// GatewayUseCase handles gateway business logic.
type GatewayUseCase struct {
	cm ConnManager
}

// NewGatewayUseCase creates a GatewayUseCase.
func NewGatewayUseCase(cm ConnManager) *GatewayUseCase {
	return &GatewayUseCase{cm: cm}
}

// HandleConnection processes a WebSocket connection for the given user and device.
// Auth is already verified by the caller — the connection is immediately ready for messaging.
func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn, userID, deviceID string) {
	defer uc.cm.Remove(userID, deviceID)

	// Send periodic pings.
	done := make(chan struct{})
	defer close(done)
	go func() {
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
	}()

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
			conn.WriteMessage(websocket.BinaryMessage, frame)
		default:
			_ = payload
			_ = ft
		}
	}
}
