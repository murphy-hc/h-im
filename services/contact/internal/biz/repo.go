package biz

import "context"

// FriendInfo holds friend information for list queries.
type FriendInfo struct {
	UserID   string
	Nickname string
	Avatar   string
	Status   int32
}

// FriendRequest holds request details.
type FriendRequest struct {
	RequestID string
	FromUser  string
	ToUser    string
	Message   string
	Status    int32
}

// ContactRepo defines the contact repository interface.
type ContactRepo interface {
	SendRequest(ctx context.Context, reqID, fromUser, toUser, msg string) error
	RespondRequest(ctx context.Context, reqID string, accept bool) error
	FindPendingRequest(ctx context.Context, fromUser, toUser string) (*FriendRequest, error)
	RemoveFriend(ctx context.Context, userID, friendID string) error
	BlockUser(ctx context.Context, userID, blockID string) error
	UnblockUser(ctx context.Context, userID, blockID string) error
	GetFriends(ctx context.Context, userID string, offset, limit int32) ([]FriendInfo, error)
	GetRequests(ctx context.Context, userID string) ([]FriendRequest, error)
}
