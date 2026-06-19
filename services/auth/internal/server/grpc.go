package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/auth/v1"
	"github.com/murphy-hc/h-im/services/auth/internal/conf"
	"github.com/murphy-hc/h-im/services/auth/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the auth service registered.
func NewGRPCServer(c *conf.Server, svc *service.AuthService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, svc)
	reflection.Register(s)
	return s
}
