package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/sequence/v1"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"github.com/murphy-hc/h-im/services/sequence/internal/service"

	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer creates a kratos gRPC server with the sequence service registered.
func NewGRPCServer(c *conf.Server, svc *service.SequenceService) *grpc.Server {
	srv := grpc.NewServer(
		grpc.Address(c.GRPC.Addr),
	)
	pb.RegisterSequenceServiceServer(srv, svc)
	return srv
}
