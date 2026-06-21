package data

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/pkg/kafka"
	"google.golang.org/protobuf/proto"
)

const kafkaTopicPrivate = "him.private.message"

// grpcMessageClient handles gRPC calls to the message service (used for Ack).
type grpcMessageClient struct {
	client pb.MessageServiceClient
}

func NewGrpcMessageClient() (*grpcMessageClient, func(), error) {
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///message.default.svc.cluster.local:9102"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &grpcMessageClient{client: pb.NewMessageServiceClient(conn)}, func() { conn.Close() }, nil
}

func (c *grpcMessageClient) ackMessage(ctx context.Context, serverID int64, userID string) error {
	_, err := c.client.AckMessage(ctx, &pb.AckMessageReq{MessageServerId: serverID, UserId: userID})
	return err
}

// KafkaMessageClient implements biz.MessageClient, sending messages via Kafka
// and delegating Ack calls to the gRPC client.
type KafkaMessageClient struct {
	producer *kafka.Producer
	grpc     *grpcMessageClient
}

// NewKafkaMessageClient creates a Kafka-backed MessageClient.
func NewKafkaMessageClient(grpc *grpcMessageClient) (*KafkaMessageClient, func(), error) {
	brokers := kafkaBrokers()
	p := kafka.NewProducer(brokers, kafka.WithProducerTracing())
	return &KafkaMessageClient{producer: p, grpc: grpc}, func() { p.Close() }, nil
}

func kafkaBrokers() []string {
	if v := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")); v != "" {
		return strings.Split(v, ",")
	}
	return []string{"localhost:9092"}
}

// SendMessage produces the message to Kafka and returns immediately.
func (c *KafkaMessageClient) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	payload, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = c.producer.Send(ctx, kafkaTopicPrivate, kafka.Message{
		Key:   []byte(req.MessageClientId),
		Value: payload,
	})
	if err != nil {
		log.Context(ctx).Errorf("kafka send: %v", err)
		return nil, err
	}
	return &pb.SendMessageResp{ServerTime: time.Now().UnixMilli()}, nil
}

// AckMessage delegates to the gRPC client.
func (c *KafkaMessageClient) AckMessage(ctx context.Context, serverID int64, userID string) error {
	return c.grpc.ackMessage(ctx, serverID, userID)
}
