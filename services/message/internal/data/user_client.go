package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	userpb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
)

// UserClient queries the User service for online status.
type UserClient struct {
	client userpb.UserServiceClient
}

// NewUserClient creates a gRPC client for the User service.
func NewUserClient() (*UserClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///user.default.svc.cluster.local:9101"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &UserClient{client: userpb.NewUserServiceClient(conn)}, func() { conn.Close() }, nil
}

// GetUserOnline returns all online devices for a user.
func (c *UserClient) GetUserOnline(ctx context.Context, userID string) (*userpb.GetUserOnlineResponse, error) {
	return c.client.GetUserOnline(ctx, &userpb.GetUserOnlineRequest{UserId: userID})
}
