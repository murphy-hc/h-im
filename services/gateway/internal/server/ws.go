package server

import (
	"context"
	"net/http"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

type WSServer struct{ *http.Server }

func (s *WSServer) Start(ctx context.Context) error { return s.ListenAndServe() }
func (s *WSServer) Stop(ctx context.Context) error  { return s.Shutdown(ctx) }

func NewWSServer(c *conf.Server, svc *service.GatewayService) *WSServer {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", svc.HandleWebSocket)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	return &WSServer{
		Server: &http.Server{
			Addr:         c.GetWs().GetAddr(),
			Handler:      mux,
			ReadTimeout:  c.GetWs().GetReadTimeout().AsDuration(),
			WriteTimeout: c.GetWs().GetWriteTimeout().AsDuration(),
		},
	}
}
