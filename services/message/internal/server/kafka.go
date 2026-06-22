package server

import (
	"context"
	"strings"

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

// NewConsumerGroup creates all Kafka consumer groups (main + DLQ + recall + recall-DLQ)
// as a single kratos transport.Server.
func NewConsumerGroup(bc *conf.Bootstrap, svc *service.KafkaService) *KafkaServers {
	kc := bc.GetData().GetKafka()
	brokers := kc.GetBrokers()
	mainTopic := kc.GetTopic()
	recallTopic := strings.Replace(mainTopic, ".message", ".recall", 1)

	return &KafkaServers{groups: []*kafka.ConsumerGroup{
		kafka.NewConsumerGroup(brokers, kc.GetGroupId(), mainTopic, svc.Handle,
			&kafka.DLQConfig{GroupID: kc.GetGroupId() + "-dlq"},
			kafka.WithConsumerTracing(),
		),
		kafka.NewConsumerGroup(brokers, kc.GetGroupId()+"-recall", recallTopic, svc.Handle,
			&kafka.DLQConfig{GroupID: kc.GetGroupId() + "-recall-dlq"},
			kafka.WithConsumerTracing(),
		),
	}}
}
