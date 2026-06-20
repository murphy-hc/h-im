package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	pb "github.com/murphy-hc/h-im/gen/go/him/message/v1"
)

// MessageClient proxies calls to the Message service.
type MessageClient struct {
	client pb.MessageServiceClient
}

// NewMessageClient creates a Kratos gRPC client for the Message service.
func NewMessageClient() (*MessageClient, func(), error) {
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///message.default.svc.cluster.local:9102"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &MessageClient{client: pb.NewMessageServiceClient(conn)}, func() { conn.Close() }, nil
}

// SendMessage sends a private message.
func (c *MessageClient) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageResp, error) {
	resp, err := c.client.SendMessage(ctx, req)
	if err != nil {
		log.Errorf("message client send: %v", err)
		return nil, err
	}
	return resp, nil
}

// AckMessage acknowledges receipt of a message.
func (c *MessageClient) AckMessage(ctx context.Context, serverID int64, userID string) error {
	_, err := c.client.AckMessage(ctx, &pb.AckMessageReq{MessageServerId: serverID, UserId: userID})
	return err
}
