package service

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/data"
)

// GatewayService handles WebSocket connections.
type GatewayService struct {
	uc       *biz.GatewayUseCase
	cm       biz.ConnManager
	upgrader websocket.Upgrader
	cfg      *conf.User
	appRepo  *data.AppRepo
}

// NewGatewayService creates a GatewayService.
func NewGatewayService(uc *biz.GatewayUseCase, cm biz.ConnManager, upgrader websocket.Upgrader, cfg *conf.User, appRepo *data.AppRepo) *GatewayService {
	return &GatewayService{uc: uc, cm: cm, upgrader: upgrader, cfg: cfg, appRepo: appRepo}
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

	// Upgrade first — avoid wasted DB work on failed upgrades.
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade failed: %v", err)
		return
	}

	// Verify app token after successful upgrade.
	app, err := s.appRepo.FindByAppID(r.Context(), appID)
	if err != nil {
		conn.Close()
		return
	}
	if err := biz.VerifyAppToken(app.AppSecret, userID, token); err != nil {
		conn.Close()
		return
	}

	if !s.cfg.GetMultiDevice() {
		for _, old := range s.cm.KickUser(userID) {
			old.Close()
		}
	}

	s.cm.Add(userID, deviceID, conn)
	s.uc.HandleConnection(r.Context(), conn, userID, deviceID)
}
