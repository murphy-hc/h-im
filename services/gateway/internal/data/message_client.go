package data

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/log"
	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
	"github.com/murphy-hc/h-im/pkg/kafka"
	"google.golang.org/protobuf/proto"
)

const (
	kafkaTopicPrivate  = "him.private.message"
	kafkaTopicChatroom = "him.chatroom.message"
	kafkaTopicGroup    = "him.group.message"
)

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

// KafkaMessageClient implements biz.MessageClient, sending messages via Kafka.
type KafkaMessageClient struct {
	producer *kafka.Producer
	grpc     *grpcMessageClient
}

// NewKafkaMessageClient creates a Kafka-backed MessageClient.
func NewKafkaMessageClient(grpc *grpcMessageClient) (*KafkaMessageClient, func(), error) {
	brokers := kafkaBrokers()
	producer := kafka.NewProducer(brokers, kafka.WithProducerTracing())
	return &KafkaMessageClient{
		producer: producer,
		grpc:     grpc,
	}, func() { producer.Close() }, nil
}

func kafkaBrokers() []string {
	if v := strings.TrimSpace(os.Getenv("KAFKA_BROKERS")); v != "" {
		return strings.Split(v, ",")
	}
	return []string{"localhost:9092"}
}

func (c *KafkaMessageClient) sendEnvelope(ctx context.Context, topic, key string, envelope *pb.MessageEnvelope) error {
	data, err := proto.Marshal(envelope)
	if err != nil {
		return err
	}
	msg := kafka.Message{Value: data}
	if key != "" {
		msg.Key = []byte(key)
	}
	if err := c.producer.Send(ctx, topic, msg); err != nil {
		log.Context(ctx).Errorf( "kafka send: %v", err)
		return err
	}
	return nil
}

func (c *KafkaMessageClient) topicFor(convType pb.ConversationType) string {
	switch convType {
	case pb.ConversationType_CONVERSATION_ROOM:
		return kafkaTopicChatroom
	case pb.ConversationType_CONVERSATION_GROUP:
		return kafkaTopicGroup
	default:
		return kafkaTopicPrivate
	}
}

// SendMessage wraps the request in a MessageEnvelope and produces to Kafka.
func (c *KafkaMessageClient) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	if req.ConvType == pb.ConversationType_CONVERSATION_ROOM {
		return c.sendChatroom(ctx, req)
	}
	return c.sendPrivate(ctx, req)
}

func (c *KafkaMessageClient) sendPrivate(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	err := c.sendEnvelope(ctx, c.topicFor(req.ConvType), req.MessageClientId, &pb.MessageEnvelope{
		Type:    pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_SEND,
		Payload: &pb.MessageEnvelope_Send{Send: req},
	})
	if err != nil {
		return nil, err
	}
	return &pb.SendMessageResp{ServerTime: time.Now().UnixMilli()}, nil
}

func (c *KafkaMessageClient) sendChatroom(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	err := c.sendEnvelope(ctx, c.topicFor(req.ConvType), req.MessageClientId, &pb.MessageEnvelope{
		Type:    pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_CHATROOM_SEND,
		Payload: &pb.MessageEnvelope_ChatroomSend{ChatroomSend: req},
	})
	if err != nil {
		return nil, err
	}
	return &pb.SendMessageResp{ServerTime: time.Now().UnixMilli()}, nil
}

// AckMessage delegates to the gRPC client.
func (c *KafkaMessageClient) AckMessage(ctx context.Context, serverID int64, userID string) error {
	return c.grpc.ackMessage(ctx, serverID, userID)
}

// RecallMessage wraps the request in a MessageEnvelope and produces to Kafka.
func (c *KafkaMessageClient) RecallMessage(ctx context.Context, req *pb.RecallMessageReq) error {
	return c.sendEnvelope(ctx, kafkaTopicPrivate, "", &pb.MessageEnvelope{
		Type:    pb.MessagePayloadType_MESSAGE_PAYLOAD_TYPE_RECALL,
		Payload: &pb.MessageEnvelope_Recall{Recall: req},
	})
}
