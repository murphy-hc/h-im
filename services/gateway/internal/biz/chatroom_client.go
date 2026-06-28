package biz

import "context"

// ChatroomClient is the interface for calling the chatroom service.
type ChatroomClient interface {
	JoinRoom(ctx context.Context, roomID, userID string) error
	LeaveRoom(ctx context.Context, roomID, userID string) error
}
