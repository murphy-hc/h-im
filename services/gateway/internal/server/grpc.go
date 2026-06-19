package server

import (
	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.GatewayGrpcService) *grpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	srv := grpc.NewServer(
		grpc.Address(":9200"),
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram)),
		),
	)
	gatewayv1.RegisterGatewayServiceServer(srv, svc)
	return srv
}
