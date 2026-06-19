package biz

import (
	"context"

	"github.com/gorilla/websocket"
)

// GatewayUseCase handles gateway business logic.
type GatewayUseCase struct{}

// NewGatewayUseCase creates a GatewayUseCase.
func NewGatewayUseCase() *GatewayUseCase {
	return &GatewayUseCase{}
}

// HandleConnection processes a WebSocket connection.
func (uc *GatewayUseCase) HandleConnection(ctx context.Context, conn *websocket.Conn) {
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		_ = msg // TODO: route to backend services
	}
}
