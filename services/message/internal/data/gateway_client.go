package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	gwpb "github.com/murphy-hc/h-im/gen/go/him/gateway/v1"
)

// GatewayClient proxies calls to the Gateway service.
type GatewayClient struct {
	client gwpb.GatewayServiceClient
}

// NewGatewayClient creates a Kratos gRPC client for the Gateway service.
func NewGatewayClient() (*GatewayClient, func(), error) {
	conn, err := grpc.DialInsecure(context.Background(),
		grpc.WithEndpoint("discovery:///gateway.default.svc.cluster.local:9200"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &GatewayClient{client: gwpb.NewGatewayServiceClient(conn)}, func() { conn.Close() }, nil
}

// SendToUser pushes a message to a specific user via the gateway.
func (c *GatewayClient) SendToUser(ctx context.Context, userID string, frameType int32, payload []byte) error {
	_, err := c.client.SendToUser(ctx, &gwpb.SendToUserRequest{
		UserId:    userID,
		FrameType: frameType,
		Payload:   payload,
	})
	return err
}
