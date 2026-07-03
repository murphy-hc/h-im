package kafka

import (
	"context"
	"time"

	"github.com/murphy-hc/h-im/pkg/gp"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

const (
	defaultMaxRetries = 3
	defaultRetryBase  = 500 * time.Millisecond
)

// MessageHandler is the callback for consumed messages.
type MessageHandler func(ctx context.Context, msg Message) error

// AlertHandler is called when a message exhausts all retry attempts.
type AlertHandler func(ctx context.Context, msg Message, err error)

// DeadLetterConfig holds the dead-letter configuration.
type DeadLetterConfig struct {
	Producer *Producer
	Topic    string
}

// Consumer reads messages from a single Kafka topic. It implements
// kratos.transport.Server (Start/Stop).
type Consumer struct {
	r           *kafka.Reader
	handler     MessageHandler
	onAlert     AlertHandler
	tracing     bool
	maxRetries  int
	retryBase   time.Duration
	dlq         *DeadLetterConfig
	startOffset int64
}

// ConsumerOption configures a Consumer.
type ConsumerOption func(*Consumer)

// WithConsumerTracing enables extraction of OpenTelemetry trace context.
func WithConsumerTracing() ConsumerOption {
	return func(c *Consumer) { c.tracing = true }
}

// WithConsumerRetry sets the retry policy.
func WithConsumerRetry(maxRetries int, baseDelay time.Duration) ConsumerOption {
	return func(c *Consumer) { c.maxRetries = maxRetries; c.retryBase = baseDelay }
}

// WithDeadLetter configures a dead-letter queue.
func WithDeadLetter(producer *Producer, topic string) ConsumerOption {
	return func(c *Consumer) { c.dlq = &DeadLetterConfig{Producer: producer, Topic: topic} }
}

// WithAlertHandler sets the callback for messages that fail all retries.
func WithAlertHandler(h AlertHandler) ConsumerOption {
	return func(c *Consumer) { c.onAlert = h }
}

// WithStartOffset sets the initial offset for new consumer groups.
// Default: kafka.FirstOffset (replay history). Use kafka.LastOffset for
// groups that should only process new messages (e.g. broadcast consumers).
func WithStartOffset(offset int64) ConsumerOption {
	return func(c *Consumer) { c.startOffset = offset }
}

// NewConsumer creates a single Kafka Consumer (kratos transport.Server).
func NewConsumer(brokers []string, groupID string, topic string, handler MessageHandler, opts ...ConsumerOption) *Consumer {
	c := &Consumer{
		handler:    handler,
		maxRetries: defaultMaxRetries,
		retryBase:  defaultRetryBase,
		startOffset: kafka.FirstOffset,
	}
	for _, o := range opts {
		o(c)
	}
	c.r = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     brokers,
		GroupID:     groupID,
		Topic:       topic,
		StartOffset: c.startOffset,
		MinBytes:    10e3,
		MaxBytes:    10e6,
	})
	return c
}

// Start begins consuming messages. Blocks until context is cancelled.
func (c *Consumer) Start(ctx context.Context) error {
	for {
		msg, err := c.r.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			continue
		}

		hCtx := context.Background()
		if c.tracing {
			carrier := propagation.MapCarrier{}
			for _, h := range msg.Headers {
				carrier[h.Key] = string(h.Value)
			}
			hCtx = otel.GetTextMapPropagator().Extract(hCtx, carrier)
		}

		if err := c.process(ctx, hCtx, msg); err != nil {
			if c.dlq != nil {
				c.dlq.Producer.Send(ctx, c.dlq.Topic, msg)
			}
			c.r.CommitMessages(ctx, msg)
			if c.onAlert != nil {
				c.onAlert(hCtx, msg, err)
			}
		}
	}
}

func (c *Consumer) process(ctx context.Context, handlerCtx context.Context, raw kafka.Message) error {
	var lastErr error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		lastErr = c.handler(handlerCtx, Message(raw))
		if lastErr == nil {
			return c.r.CommitMessages(ctx, raw)
		}
		if attempt < c.maxRetries-1 {
			delay := c.retryBase * time.Duration(1<<attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return lastErr
}

// Stop gracefully shuts down the consumer.
func (c *Consumer) Stop(ctx context.Context) error { return c.r.Close() }

// ── ConsumerGroup ──────────────────────────────────────────────────────────

// DLQConfig configures the dead-letter consumer.
type DLQConfig struct {
	GroupID      string
	AlertHandler AlertHandler
}

// ConsumerGroup wraps a main consumer and an optional dead-letter consumer
// as a single kratos transport.Server. When DLQ is configured, the DLQ
// consumer runs in a background goroutine and processes failed messages.
type ConsumerGroup struct {
	main *Consumer
	dlq  *Consumer
}

// NewConsumerGroup creates a consumer group (main + DLQ consumer).
// DLQ is enabled by passing a non-nil DLQConfig. The DLQ topic is derived
// from the main topic (".dlq" suffix). The DLQ consumer has no DLQ of its
// own — it alerts on failure.
func NewConsumerGroup(brokers []string, groupID string, topic string, handler MessageHandler, dlqCfg *DLQConfig, opts ...ConsumerOption) *ConsumerGroup {
	var dlqConsumer *Consumer
	if dlqCfg != nil {
		dlqProducer := NewProducer(brokers, WithProducerTracing())
		dlqConsumer = NewConsumer(brokers, dlqCfg.GroupID, topic+".dlq", handler, opts...)
		dlqConsumer.onAlert = dlqCfg.AlertHandler
		// main consumer's DLQ publishes to the DLQ topic
		opts = append(opts, WithDeadLetter(dlqProducer, topic+".dlq"))
	}
	main := NewConsumer(brokers, groupID, topic, handler, opts...)
	return &ConsumerGroup{main: main, dlq: dlqConsumer}
}

// Start starts both consumers. The DLQ consumer runs in a background goroutine
// with panic recovery. The main consumer blocks until the context is cancelled.
func (g *ConsumerGroup) Start(ctx context.Context) error {
	if g.dlq != nil {
		gp.SafeGo(ctx, func(ctx context.Context) { _ = g.dlq.Start(ctx) })
	}
	return g.main.Start(ctx)
}

// Stop gracefully shuts down both consumers.
func (g *ConsumerGroup) Stop(ctx context.Context) error {
	if g.dlq != nil {
		g.dlq.Stop(ctx)
	}
	return g.main.Stop(ctx)
}
