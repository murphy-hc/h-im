package server

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

func NewHTTPServer(bc *conf.Bootstrap, meter metric.Meter) *khttp.Server {
	_ = meter
	srv := khttp.NewServer(
		khttp.Address(bc.GetServer().GetHttp().GetAddr()),
		khttp.Middleware(
			recovery.Recovery(),
			tracing.Server(),
		),
	)
	srv.Handle("/metrics", promhttp.Handler())
	srv.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	return srv
}
