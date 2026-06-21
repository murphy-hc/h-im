package service

import (
	"context"

	"github.com/murphy-hc/h-im/pkg/kafka"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
)

// KafkaService handles Kafka messages.
type KafkaService struct {
	uc *biz.SendUseCase
}

// NewKafkaService creates a KafkaService.
func NewKafkaService(uc *biz.SendUseCase) *KafkaService {
	return &KafkaService{uc: uc}
}

// Handle processes a Kafka message.
func (s *KafkaService) Handle(ctx context.Context, msg kafka.Message) error {
	return s.uc.ProcessMessage(ctx, msg.Value)
}
