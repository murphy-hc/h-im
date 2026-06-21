package kafka

import (
	"context"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Producer sends messages to Kafka topics.
type Producer struct {
	w       *kafka.Writer
	tracing bool
}

// ProducerOption configures a Producer.
type ProducerOption func(*Producer)

// WithProducerTracing enables automatic injection of OpenTelemetry trace
// context into Kafka message headers.
func WithProducerTracing() ProducerOption {
	return func(p *Producer) { p.tracing = true }
}

// Legacy alias kept for compatibility.
func WithTracing() ProducerOption { return WithProducerTracing() }

// NewProducer creates a new Kafka Producer.
func NewProducer(brokers []string, opts ...ProducerOption) *Producer {
	p := &Producer{
		w: &kafka.Writer{
			Addr:     kafka.TCP(brokers...),
			Balancer: &kafka.LeastBytes{},
		},
	}
	for _, o := range opts {
		o(p)
	}
	return p
}

// Send sends messages to the given topic. If tracing is enabled, the
// current trace context is injected into the message headers.
func (p *Producer) Send(ctx context.Context, topic string, msgs ...Message) error {
	if p.tracing {
		for i := range msgs {
			carrier := propagation.MapCarrier{}
			otel.GetTextMapPropagator().Inject(ctx, carrier)
			for k, v := range carrier {
				msgs[i].Headers = append(msgs[i].Headers, kafka.Header{Key: k, Value: []byte(v)})
			}
		}
	}
	msg := kafka.Message{Topic: topic}
	if len(msgs) == 1 {
		msg = msgs[0]
		msg.Topic = topic
		return p.w.WriteMessages(ctx, msg)
	}
	kmsgs := make([]kafka.Message, len(msgs))
	for i, m := range msgs {
		kmsgs[i] = m
		kmsgs[i].Topic = topic
	}
	return p.w.WriteMessages(ctx, kmsgs...)
}

// Close shuts down the producer.
func (p *Producer) Close() error { return p.w.Close() }
