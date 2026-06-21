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

func (s *MessageService) PullMessages(ctx context.Context, req *pb.PullMessagesReq) (*pb.PullMessagesResp, error) {
	limit := req.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	msgs, err := s.sendUC.PullMessagesSince(ctx, req.UserId, req.SinceMessageId, limit)
	if err != nil {
		return nil, err
	}
	pbMsgs := make([]*pb.Message, 0, len(msgs))
	for _, m := range msgs {
		pbMsgs = append(pbMsgs, &pb.Message{
			MessageServerId: m.MessageServerID,
			MessageClientId: m.MessageClientID,
			SenderId:        m.SenderID,
			ReceiverId:      m.ReceiverID,
			ConvType:        pb.ConversationType(m.ConvType),
			MsgType:         pb.MessageType(m.MsgType),
			Text:            m.Text,
			ServerTime:      m.ServerTime,
			CreateTime:      m.CreateTime,
			IsDeleted:       m.IsDeleted,
			IsRemoteRead:    m.IsRemoteRead,
		})
	}
	return &pb.PullMessagesResp{Messages: pbMsgs}, nil
}
