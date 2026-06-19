package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/murphy-hc/h-im/pkg/logger"
	"github.com/murphy-hc/h-im/services/auth/internal/conf"
)

func main() {
	log := logger.WithService(logger.New(os.Getenv("LOG_LEVEL")), "auth")

	cfg := &conf.Server{
		GRPC: conf.GRPCServer{Addr: envOrDefault("Auth_GRPC_ADDR", ":9100")},
	}
	dataCfg := &conf.Data{}

	grpcServer, cleanup, err := wireApp(cfg, dataCfg)
	if err != nil {
		log.Error("wireApp failed", "error", err)
		os.Exit(1)
	}
	defer cleanup()

	addr := cfg.GRPC.Addr
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Error("listen failed", "error", err)
		os.Exit(1)
	}

	go func() {
		log.Info("auth service starting", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("serve failed", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down...")
	grpcServer.GracefulStop()
	log.Info("auth service stopped")
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
