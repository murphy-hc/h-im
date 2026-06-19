package service

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// GatewayService handles WebSocket connections.
type GatewayService struct {
	uc *biz.GatewayUseCase
}

// NewGatewayService creates a GatewayService.
func NewGatewayService(uc *biz.GatewayUseCase) *GatewayService {
	return &GatewayService{uc: uc}
}

// HandleWebSocket handles a WebSocket upgrade request.
func (s *GatewayService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// TODO: authenticate via JWT, route messages to backend services
	s.uc.HandleConnection(r.Context(), conn)
}
