package server

import (
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"

	pb "github.com/murphy-hc/h-im/gen/go/him/auth/v1"
	"github.com/murphy-hc/h-im/services/auth/internal/conf"
	"github.com/murphy-hc/h-im/services/auth/internal/service"
)

func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.AuthService) *grpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	opts := []grpc.ServerOption{
		grpc.Address(bc.GetServer().GetGrpc().GetAddr()),
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metadata.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram)),
		),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterAuthServiceServer(srv, svc)
	return srv
}
