package server

import (
	"context"

	"golang.org/x/time/rate"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"

	gatewayv1 "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

// rateLimitMiddleware returns a Kratos middleware that rate-limits requests.
// Default: 1000 req/s with burst of 2000.
func rateLimitMiddleware() middleware.Middleware {
	limiter := rate.NewLimiter(1000, 2000)
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if !limiter.Allow() {
				return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded")
			}
			return handler(ctx, req)
		}
	}
}

func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.GatewayGrpcService) *kgrpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	srv := kgrpc.NewServer(
		kgrpc.Address(bc.GetServer().GetGrpc().GetAddr()),
		kgrpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			rateLimitMiddleware(),
			metadata.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram)),
		),
	)
	gatewayv1.RegisterGatewayServiceServer(srv, svc)
	return srv
}
