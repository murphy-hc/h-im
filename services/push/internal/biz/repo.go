package biz

import "context"

// DeviceInfo holds push device registration info.
type DeviceInfo struct {
	DeviceID    string
	DeviceToken string
	Platform    int32
}

// PushRepo defines the push repository interface.
type PushRepo interface {
	RegisterDevice(ctx context.Context, userID string, info *DeviceInfo) error
	UnregisterDevice(ctx context.Context, deviceID string) error
	FindDevicesByUser(ctx context.Context, userID string) ([]DeviceInfo, error)
}

// Pusher sends push notifications to devices.
type Pusher interface {
	Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error
}
