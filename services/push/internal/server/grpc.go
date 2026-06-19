package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
	"github.com/murphy-hc/h-im/services/push/internal/conf"
	"github.com/murphy-hc/h-im/services/push/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the push service registered.
func NewGRPCServer(c *conf.Server, svc *service.PushService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterPushServiceServer(s, svc)
	reflection.Register(s)
	return s
}
