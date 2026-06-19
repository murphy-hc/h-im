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
	cm biz.ConnManager
}

// NewGatewayService creates a GatewayService.
func NewGatewayService(uc *biz.GatewayUseCase, cm biz.ConnManager) *GatewayService {
	return &GatewayService{uc: uc, cm: cm}
}

// HandleWebSocket handles a WebSocket upgrade request.
func (s *GatewayService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// For now, generate anonymous ID (auth to be added later)
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = "anon-" + r.RemoteAddr
	}

	s.cm.Add(userID, conn)
	s.uc.HandleConnection(r.Context(), conn, userID)
}
