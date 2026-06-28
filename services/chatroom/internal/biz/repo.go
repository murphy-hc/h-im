package biz

import "context"

// Room is the domain entity for a chatroom.
type Room struct {
	RoomID      string
	Name        string
	OwnerID     string
	MemberCount int32
	CreatedAt   int64
}

// ChatroomRepo defines the chatroom repository interface.
type ChatroomRepo interface {
	CreateRoom(ctx context.Context, roomID, name, ownerID string) error
	FindByID(ctx context.Context, roomID string) (*Room, error)
	JoinRoom(ctx context.Context, roomID, userID string) error
	LeaveRoom(ctx context.Context, roomID, userID string) error
	GetMembers(ctx context.Context, roomID string) ([]string, error)
}
