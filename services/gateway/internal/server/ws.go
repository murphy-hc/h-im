package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

// WSServer wraps an http.Server for WebSocket connections.
// Implements github.com/go-kratos/kratos/v2/transport.Server.
type WSServer struct {
	*http.Server
}

// Start begins serving WebSocket connections.
func (s *WSServer) Start(ctx context.Context) error {
	return s.ListenAndServe()
}

// Stop gracefully shuts down the WebSocket server.
func (s *WSServer) Stop(ctx context.Context) error {
	return s.Shutdown(ctx)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// NewWSServer creates a WebSocket server.
func NewWSServer(c *conf.Server, svc *service.GatewayService) *WSServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return &WSServer{
		Server: &http.Server{
			Addr:        c.GetWs().GetAddr(),
			Handler:     mux,
			ReadTimeout: 10 * time.Second,
		},
	}
}
