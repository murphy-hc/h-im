package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/contact/v1"
	"github.com/murphy-hc/h-im/services/contact/internal/conf"
	"github.com/murphy-hc/h-im/services/contact/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the contact service registered.
func NewGRPCServer(c *conf.Server, svc *service.ContactService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterContactServiceServer(s, svc)
	reflection.Register(s)
	return s
}
