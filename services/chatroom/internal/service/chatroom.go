package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/chatroom/v1"
	"github.com/murphy-hc/h-im/services/chatroom/internal/biz"
)

// ChatroomService implements the ChatroomService gRPC server.
type ChatroomService struct {
	pb.UnimplementedChatroomServiceServer
	uc *biz.ChatroomUseCase
}

// NewChatroomService creates a ChatroomService.
func NewChatroomService(uc *biz.ChatroomUseCase) *ChatroomService {
	return &ChatroomService{uc: uc}
}
