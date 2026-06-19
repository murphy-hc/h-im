package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
	"github.com/murphy-hc/h-im/services/user/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the user service registered.
func NewGRPCServer(c *conf.Server, svc *service.UserService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, svc)
	reflection.Register(s)
	return s
}
