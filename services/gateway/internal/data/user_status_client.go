package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	userpb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
)

// UserStatusClient reports heartbeat/disconnect events to the user service.
type UserStatusClient struct {
	client userpb.UserServiceClient
}

// NewUserStatusClient creates a gRPC client for the User service.
func NewUserStatusClient() (*UserStatusClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///user.default.svc.cluster.local:9101"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &UserStatusClient{client: userpb.NewUserServiceClient(conn)}, func() { conn.Close() }, nil
}

func (c *UserStatusClient) ReportHeartbeat(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error {
	_, err := c.client.ReportHeartbeat(ctx, &userpb.ReportHeartbeatRequest{
		UserId:      userID,
		DeviceId:    deviceID,
		GatewayAddr: gatewayAddr,
		Timestamp:   timestamp,
	})
	return err
}

func (c *UserStatusClient) ReportDisconnect(ctx context.Context, userID, deviceID string) error {
	_, err := c.client.ReportDisconnect(ctx, &userpb.ReportDisconnectRequest{
		UserId:   userID,
		DeviceId: deviceID,
	})
	return err
}
