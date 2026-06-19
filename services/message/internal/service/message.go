package service

import (
	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
)

// MessageService implements the MessageService gRPC server.
type MessageService struct {
	pb.UnimplementedMessageServiceServer
	uc *biz.MessageUseCase
}

// NewMessageService creates a MessageService.
func NewMessageService(uc *biz.MessageUseCase) *MessageService {
	return &MessageService{uc: uc}
}
