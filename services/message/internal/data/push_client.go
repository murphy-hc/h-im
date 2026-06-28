package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	pushpb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
)

// PushClient proxies push calls to the Push service.
type PushClient struct {
	client pushpb.PushServiceClient
}

// NewPushClient creates a gRPC client for the Push service.
func NewPushClient() (*PushClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///push.default.svc.cluster.local:9106"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &PushClient{client: pushpb.NewPushServiceClient(conn)}, func() { conn.Close() }, nil
}

func (c *PushClient) PushToUser(ctx context.Context, userID, title, body string, payload []byte) error {
	_, err := c.client.PushToUser(ctx, &pushpb.PushToUserRequest{
		UserId: userID, Title: title, Body: body, Payload: string(payload),
	})
	return err
}
