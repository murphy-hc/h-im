package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	grouppb "github.com/murphy-hc/h-im/gen/go/him/group/v1"
)

// GroupClient proxies calls to the Group service.
type GroupClient struct {
	client grouppb.GroupServiceClient
}

// NewGroupClient creates a gRPC client for the Group service.
func NewGroupClient() (*GroupClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///group.default.svc.cluster.local:9104"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &GroupClient{client: grouppb.NewGroupServiceClient(conn)}, func() { conn.Close() }, nil
}

func (c *GroupClient) JoinGroup(ctx context.Context, groupID, userID string) error {
	_, err := c.client.JoinGroup(ctx, &grouppb.JoinGroupRequest{GroupId: groupID, UserId: userID})
	return err
}

func (c *GroupClient) LeaveGroup(ctx context.Context, groupID, userID string) error {
	_, err := c.client.LeaveGroup(ctx, &grouppb.LeaveGroupRequest{GroupId: groupID, UserId: userID})
	return err
}
