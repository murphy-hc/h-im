package biz

import "context"

// MessageRepo defines the message repository interface.
type MessageRepo interface {
	Insert(ctx context.Context, m *Message) error
	InsertChatroom(ctx context.Context, serverID int64, clientID, roomID, senderID string, msgType int32, text, attachment string, serverTime int64) error
	InsertGroup(ctx context.Context, serverID int64, clientID, groupID, senderID string, msgType int32, text, attachment string, serverTime int64) error
	GetReceiverID(ctx context.Context, serverID int64) (string, error)
	MarkDelivered(ctx context.Context, serverID int64) error
	MarkRead(ctx context.Context, serverID int64) error
	MarkRecalled(ctx context.Context, serverID int64, senderID string, serverTime int64) (bool, error)
	PullSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]Message, error)
}

// MessageGateway sends messages to users via the gateway service.
type MessageGateway interface {
	SendToDevice(ctx context.Context, gatewayAddr, userID string, frameType int32, payload []byte) error
	BroadcastToRoom(ctx context.Context, roomID string, frameType int32, payload []byte) (int32, error)
	BroadcastToGroup(ctx context.Context, groupID string, frameType int32, payload []byte) (int32, error)
}

// UserStatusClient queries user online status.
type UserStatusClient interface {
	GetUserOnline(ctx context.Context, userID string) ([]OnlineDevice, error)
}

// OnlineDevice represents a device currently online.
type OnlineDevice struct {
	DeviceID      string
	GatewayAddr   string
	LastHeartbeat int64
}

// PushClient sends push notifications when users are offline.
type PushClient interface {
	PushToUser(ctx context.Context, userID, title, body string, payload []byte) error
}
