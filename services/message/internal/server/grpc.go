package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/services/message/internal/conf"
	"github.com/murphy-hc/h-im/services/message/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the message service registered.
func NewGRPCServer(c *conf.Server, svc *service.MessageService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterMessageServiceServer(s, svc)
	reflection.Register(s)
	return s
}
