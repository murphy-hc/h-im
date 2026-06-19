package server

import (
	"net/http"
	"time"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

// NewWSServer creates an HTTP server with WebSocket support.
func NewWSServer(c *conf.Server, svc *service.GatewayService) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return &http.Server{
		Addr:         c.WS.Addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}
