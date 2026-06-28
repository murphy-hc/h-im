package service

import (
	"context"

	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/pkg/kafka"
	"github.com/murphy-hc/h-im/pkg/logger"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
	"google.golang.org/protobuf/proto"
)

// KafkaService handles Kafka messages.
type KafkaService struct {
	uc *biz.SendUseCase
}

// NewKafkaService creates a KafkaService.
func NewKafkaService(uc *biz.SendUseCase) *KafkaService {
	return &KafkaService{uc: uc}
}

// Handle dispatches by envelope type.
func (s *KafkaService) Handle(ctx context.Context, msg kafka.Message) error {
	var env pb.MessageEnvelope
	if err := proto.Unmarshal(msg.Value, &env); err != nil {
		return err
	}
	switch env.Type {
	case pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_SEND:
		req := env.GetSend()
		_, err := s.uc.SendPrivateMessage(ctx, req.SenderId, req.ReceiverId, int32(req.MsgType), req.Text, req.MessageClientId, attachmentBytes(req.Attachment))
		return err
	case pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_CHATROOM_SEND:
		req := env.GetChatroomSend()
		_, err := s.uc.SendChatroomMessage(ctx, req.SenderId, req.ReceiverId, int32(req.MsgType), req.Text, req.MessageClientId, attachmentBytes(req.Attachment))
		return err
	case pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_RECALL:
		req := env.GetRecall()
		return s.uc.RecallMessage(ctx, req.MessageServerId, req.SenderId)
	default:
		logger.ContextWarnf(ctx, "unknown message payload type: %v", env.Type)
		return nil
	}
}
