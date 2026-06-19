package server

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/chatroom/v1"
	"github.com/murphy-hc/h-im/services/chatroom/internal/conf"
	"github.com/murphy-hc/h-im/services/chatroom/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewGRPCServer creates a gRPC server with the chatroom service registered.
func NewGRPCServer(c *conf.Server, svc *service.ChatroomService) *grpc.Server {
	s := grpc.NewServer()
	pb.RegisterChatroomServiceServer(s, svc)
	reflection.Register(s)
	return s
}
