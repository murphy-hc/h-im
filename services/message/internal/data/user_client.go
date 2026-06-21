package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	userpb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/services/message/internal/biz"
)

var _ biz.UserStatusClient = (*UserClient)(nil)

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
func (c *UserClient) GetUserOnline(ctx context.Context, userID string) ([]biz.OnlineDevice, error) {
	resp, err := c.client.GetUserOnline(ctx, &userpb.GetUserOnlineRequest{UserId: userID})
	if err != nil {
		return nil, err
	}
	devices := make([]biz.OnlineDevice, 0, len(resp.GetDevices()))
	for _, d := range resp.GetDevices() {
		devices = append(devices, biz.OnlineDevice{
			DeviceID:      d.GetDeviceId(),
			GatewayAddr:   d.GetGatewayAddr(),
			LastHeartbeat: d.GetLastHeartbeat(),
		})
	}
	return devices, nil
}
