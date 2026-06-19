package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/group/v1"
	"github.com/murphy-hc/h-im/services/group/internal/conf"
	"github.com/murphy-hc/h-im/services/group/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the group service registered.
func NewGRPCServer(c *conf.Server, svc *service.GroupService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterGroupServiceServer(s, svc)
	reflection.Register(s)
	return s
}
