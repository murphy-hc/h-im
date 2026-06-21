package service

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

// GatewayService handles WebSocket connections.
type GatewayService struct {
	uc       *biz.GatewayUseCase
	cm       biz.ConnManager
	upgrader websocket.Upgrader
	cfg      *conf.User
}

// NewGatewayService creates a GatewayService.
func NewGatewayService(uc *biz.GatewayUseCase, cm biz.ConnManager, upgrader websocket.Upgrader, cfg *conf.User) *GatewayService {
	return &GatewayService{uc: uc, cm: cm, upgrader: upgrader, cfg: cfg}
}

// HandleWebSocket handles a WebSocket upgrade request.
func (s *GatewayService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	appID := q.Get("app_id")
	userID := q.Get("user_id")
	token := q.Get("token")
	deviceID := q.Get("device_id")
	if deviceID == "" {
		deviceID = "default"
	}
	if appID == "" || userID == "" || token == "" {
		http.Error(w, "missing app_id, user_id or token", http.StatusUnauthorized)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	valid, err := s.uc.ValidateToken(r.Context(), appID, userID, token)
	if err != nil || !valid {
		log.Printf("token validation failed: app=%s user=%s err=%v", appID, userID, err)
		conn.Close()
		return
	}

	if !s.cfg.GetMultiDevice() {
		if conns, err := s.cm.KickUser(userID); err == nil {
			for _, old := range conns {
				old.Close()
			}
		}
	}

	s.cm.Add(userID, deviceID, conn)
	s.uc.HandleConnection(r.Context(), conn, userID, deviceID)
}
