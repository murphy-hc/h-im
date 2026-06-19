package server

import (
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"

	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"github.com/murphy-hc/h-im/services/sequence/internal/service"
)

// NewGRPCServer creates a kratos gRPC server with observability middleware.
func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.SequenceService) *grpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)

	opts := []grpc.ServerOption{
		grpc.Address(bc.GetServer().GetGrpc().GetAddr()),
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			metadata.Server(),
			metrics.Server(
				metrics.WithRequests(counter),
				metrics.WithSeconds(histogram),
			),
		),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterSequenceServiceServer(srv, svc)
	return srv
}
