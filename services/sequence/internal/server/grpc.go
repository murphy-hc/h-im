package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"github.com/murphy-hc/h-im/services/sequence/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the sequence service registered.
func NewGRPCServer(c *conf.Server, svc *service.SequenceService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterSequenceServiceServer(s, svc)
	reflection.Register(s)
	return s
}
