package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
)

type MessageService struct {
	pb.UnimplementedMessageServiceServer
	sendUC *biz.SendUseCase
}

func NewMessageService(sendUC *biz.SendUseCase) *MessageService {
	return &MessageService{sendUC: sendUC}
}

func (s *MessageService) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	serverID, err := s.sendUC.SendPrivateMessage(ctx, req.SenderId, req.ReceiverId, int32(req.MsgType), req.Text, req.MessageClientId)
	if err != nil {
		return nil, err
	}
	return &pb.SendMessageResp{MessageServerId: serverID}, nil
}

func (s *MessageService) AckMessage(ctx context.Context, req *pb.AckMessageReq) (*pb.AckMessageResp, error) {
	if err := s.sendUC.AckMessage(ctx, req.MessageServerId, req.UserId); err != nil {
		return &pb.AckMessageResp{Success: false}, nil
	}
	return &pb.AckMessageResp{Success: true}, nil
}
