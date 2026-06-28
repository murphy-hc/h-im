package data

import (
	"context"

	kgrpc "github.com/go-kratos/kratos/v2/transport/grpc"
	chatroompb "github.com/murphy-hc/h-im/gen/go/him/chatroom/v1"
)

// ChatroomClient proxies calls to the Chatroom service.
type ChatroomClient struct {
	client chatroompb.ChatroomServiceClient
}

// NewChatroomClient creates a gRPC client for the Chatroom service.
func NewChatroomClient() (*ChatroomClient, func(), error) {
	conn, err := kgrpc.DialInsecure(context.Background(),
		kgrpc.WithEndpoint("discovery:///chatroom.default.svc.cluster.local:9105"),
	)
	if err != nil {
		return nil, nil, err
	}
	return &ChatroomClient{client: chatroompb.NewChatroomServiceClient(conn)}, func() { conn.Close() }, nil
}

func (c *ChatroomClient) JoinRoom(ctx context.Context, roomID, userID string) error {
	_, err := c.client.JoinRoom(ctx, &chatroompb.JoinRoomRequest{RoomId: roomID, UserId: userID})
	return err
}

func (c *ChatroomClient) LeaveRoom(ctx context.Context, roomID, userID string) error {
	_, err := c.client.LeaveRoom(ctx, &chatroompb.LeaveRoomRequest{RoomId: roomID, UserId: userID})
	return err
}
