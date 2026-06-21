package biz

import "context"

// OnlineDevice holds online status for a single device.
type OnlineDevice struct {
	DeviceID      string
	GatewayAddr   string
	LastHeartbeat int64
}

// UserRepo defines the user repository interface.
type UserRepo interface {
	SetOnline(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error
	SetOffline(ctx context.Context, userID, deviceID string) error
	GetUserOnline(ctx context.Context, userID string) ([]OnlineDevice, error)
	SweepOffline(ctx context.Context, timeoutSeconds int64) ([]OfflinePair, error)
}

type OfflinePair struct {
	UserID   string
	DeviceID string
}
