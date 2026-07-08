package server

import (
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"go.opentelemetry.io/otel/metric"

	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/pkg/jwt"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
	"github.com/murphy-hc/h-im/services/user/internal/service"
)

// Public RPC methods that do not require JWT authentication.
var publicMethods = []string{
	"/him.user.v1.UserService/Login",
	"/him.user.v1.UserService/Register",
}

func NewGRPCServer(bc *conf.Bootstrap, meter metric.Meter, svc *service.UserService, jwtMgr *jwt.Manager) *grpc.Server {
	counter, _ := metrics.DefaultRequestsCounter(meter, metrics.DefaultServerRequestsCounterName)
	histogram, _ := metrics.DefaultSecondsHistogram(meter, metrics.DefaultServerSecondsHistogramName)
	opts := []grpc.ServerOption{
		grpc.Address(bc.GetServer().GetGrpc().GetAddr()),
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			jwt.Server(jwtMgr, publicMethods...),
			metadata.Server(),
			metrics.Server(metrics.WithRequests(counter), metrics.WithSeconds(histogram)),
		),
	}
	srv := grpc.NewServer(opts...)
	pb.RegisterUserServiceServer(srv, svc)
	return srv
}
