package server

import (
	"net/http"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/metric"
)

func NewHTTPServer(bc *conf.Bootstrap, meter metric.Meter, h *MediaHTTPHandler) *khttp.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	srv := khttp.NewServer(
		khttp.Address(bc.GetServer().GetHttp().GetAddr()),
		khttp.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram)),
		),
	)
	srv.Handle("/metrics", promhttp.Handler())
	srv.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	srv.HandleFunc("/media/v1/upload", h.Upload)
	srv.HandleFunc("/media/v1/token", h.Token)
	srv.HandleFunc("/media/v1/confirm", h.Confirm)
	return srv
}
