package main

import (
	"os"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
)

func main() {
	logger := log.NewHelper(log.NewStdLogger(os.Stdout))

	cfg := &conf.Server{
		GRPC: conf.GRPCServer{Addr: envOrDefault("SEQUENCE_GRPC_ADDR", ":9108")},
	}

	app, cleanup, err := wireApp(cfg, &conf.Data{})
	if err != nil {
		logger.Fatalf("wireApp failed: %v", err)
	}
	defer cleanup()

	if err := app.Run(); err != nil {
		logger.Fatalf("app run failed: %v", err)
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
