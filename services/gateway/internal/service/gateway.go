package service

import (
	"net/http"

	"github.com/coder/websocket"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

// GatewayService handles WebSocket connections.
type GatewayService struct {
	uc      *biz.GatewayUseCase
	cm      biz.ConnManager
	cfg     *conf.User
	acceptOpts *websocket.AcceptOptions
}

// NewGatewayService creates a GatewayService.
func NewGatewayService(uc *biz.GatewayUseCase, cm biz.ConnManager, cfg *conf.User, wsCfg *conf.Server_WS) *GatewayService {
	opts := &websocket.AcceptOptions{}
	if wsCfg.GetEnableCompression() {
		opts.CompressionMode = websocket.CompressionNoContextTakeover
	}
	return &GatewayService{uc: uc, cm: cm, cfg: cfg, acceptOpts: opts}
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

	valid, err := s.uc.ValidateToken(r.Context(), appID, userID, token)
	if err != nil || !valid {
		log.Errorf("token validation failed: app=%s user=%s err=%v", appID, userID, err)
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	conn, err := websocket.Accept(w, r, s.acceptOpts)
	if err != nil {
		log.Errorf("websocket upgrade failed: %v", err)
		return
	}

	if !s.cfg.GetMultiDevice() {
		if conns, err := s.cm.KickUser(userID); err == nil {
			for _, old := range conns {
				old.Close(websocket.StatusNormalClosure, biz.CloseReasonKicked)
			}
		}
	}

	s.cm.Add(userID, deviceID, conn)
	s.uc.HandleConnection(r.Context(), conn, userID, deviceID)
}
