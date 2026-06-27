package server

import (
	"github.com/murphy-hc/h-im/pkg/kafka"
	"github.com/murphy-hc/h-im/services/message/internal/conf"
	"github.com/murphy-hc/h-im/services/message/internal/service"
)

// NewConsumerGroup creates a Kafka consumer group (main + DLQ) that
// implements kratos transport.Server.
func NewConsumerGroup(bc *conf.Bootstrap, svc *service.KafkaService) *kafka.ConsumerGroup {
	kc := bc.GetData().GetKafka()
	return kafka.NewConsumerGroup(
		kc.GetBrokers(),
		kc.GetGroupId(),
		kc.GetTopic(),
		svc.Handle,
		&kafka.DLQConfig{
			GroupID: kc.GetGroupId() + "-dlq",
		},
		kafka.WithConsumerTracing(),
	)
}
