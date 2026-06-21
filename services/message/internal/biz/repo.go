package biz

import "context"

// MessageRepo defines the message repository interface.
type MessageRepo interface {
	Insert(ctx context.Context, m *Message) error
	MarkRead(ctx context.Context, serverID int64) error
	PullSince(ctx context.Context, userID string, sinceID int64, limit int32) ([]Message, error)
}

// MessageGateway sends messages to users via the gateway service.
type MessageGateway interface {
	SendToDevice(ctx context.Context, gatewayAddr, userID string, frameType int32, payload []byte) error
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
