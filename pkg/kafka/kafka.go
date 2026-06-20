// Package rocketmq provides a RocketMQ producer/consumer wrapper.
package kafka

import (
	"context"
	"fmt"
)

// Config holds RocketMQ connection parameters.
type Config struct {
	NameServer string
	AccessKey  string
	SecretKey  string
}

// Producer wraps a RocketMQ producer.
type Producer struct {
	cfg Config
	// TODO: replace with real RocketMQ Go client when available
}

// Consumer wraps a RocketMQ consumer.
type Consumer struct {
	cfg     Config
	GroupID string
	// TODO: replace with real RocketMQ Go client when available
}

// NewProducer creates a new Producer.
func NewProducer(cfg Config) (*Producer, error) {
	if cfg.NameServer == "" {
		return nil, fmt.Errorf("rocketmq: nameserver is required")
	}
	return &Producer{cfg: cfg}, nil
}

// NewConsumer creates a new Consumer for the given consumer group.
func NewConsumer(cfg Config, groupID string) (*Consumer, error) {
	if cfg.NameServer == "" {
		return nil, fmt.Errorf("rocketmq: nameserver is required")
	}
	return &Consumer{cfg: cfg, GroupID: groupID}, nil
}

// Send sends a message to the given topic. The shardingKey is used for
// MessageGroup-based ordering (same key → same queue → ordered delivery).
func (p *Producer) Send(ctx context.Context, topic string, body []byte, shardingKey string) (string, error) {
	// TODO: integrate with RocketMQ Go SDK
	_ = ctx
	_ = topic
	_ = body
	_ = shardingKey
	return "", fmt.Errorf("rocketmq: not yet implemented")
}

// Subscribe starts consuming messages from the given topic.
func (c *Consumer) Subscribe(ctx context.Context, topic string, handler func(ctx context.Context, body []byte) error) error {
	// TODO: integrate with RocketMQ Go SDK
	_ = ctx
	_ = topic
	_ = handler
	return fmt.Errorf("rocketmq: not yet implemented")
}

// Close shuts down the producer.
func (p *Producer) Close() error { return nil }

// Close shuts down the consumer.
func (c *Consumer) Close() error { return nil }
