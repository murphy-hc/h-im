package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/media/v1"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	"github.com/murphy-hc/h-im/services/media/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the media service registered.
func NewGRPCServer(c *conf.Server, svc *service.MediaService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterMediaServiceServer(s, svc)
	reflection.Register(s)
	return s
}
