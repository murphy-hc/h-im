package server

import (
	"context"

	"github.com/murphy-hc/h-im/pkg/kafka"
	"github.com/murphy-hc/h-im/services/message/internal/conf"
	"github.com/murphy-hc/h-im/services/message/internal/service"
)

// KafkaServers wraps multiple ConsumerGroups as a single kratos transport.Server.
type KafkaServers struct {
	groups []*kafka.ConsumerGroup
}

func (s *KafkaServers) Start(ctx context.Context) error {
	for _, g := range s.groups {
		go g.Start(ctx)
	}
	<-ctx.Done()
	return nil
}

func (s *KafkaServers) Stop(ctx context.Context) error {
	for _, g := range s.groups {
		g.Stop(ctx)
	}
	return nil
}

// NewAllConsumers creates all Kafka consumer groups (private + chatroom + DLQs).
func NewAllConsumers(bc *conf.Bootstrap, svc *service.KafkaService) *KafkaServers {
	kc := bc.GetData().GetKafka()
	brokers := kc.GetBrokers()
	mainTopic := kc.GetTopic()

	return &KafkaServers{groups: []*kafka.ConsumerGroup{
		// Private messages
		kafka.NewConsumerGroup(brokers, kc.GetGroupId(), mainTopic, svc.Handle,
			&kafka.DLQConfig{GroupID: kc.GetGroupId() + "-dlq"},
			kafka.WithConsumerTracing(),
		),
		// Chatroom messages
		kafka.NewConsumerGroup(brokers, kc.GetGroupId()+"-chatroom", "him.chatroom.message", svc.Handle,
			&kafka.DLQConfig{GroupID: kc.GetGroupId() + "-chatroom-dlq"},
			kafka.WithConsumerTracing(),
		),
	}}
}
