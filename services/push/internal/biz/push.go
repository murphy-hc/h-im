package biz

import (
	"context"
	"fmt"

	pushpb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
)

// PushUseCase handles push business logic.
type PushUseCase struct {
	repo   PushRepo
	pusher Pusher
}

func NewPushUseCase(repo PushRepo, pusher Pusher) *PushUseCase {
	return &PushUseCase{repo: repo, pusher: pusher}
}

func (uc *PushUseCase) RegisterDevice(ctx context.Context, userID string, info *DeviceInfo) error {
	return uc.repo.RegisterDevice(ctx, userID, info)
}

func (uc *PushUseCase) UnregisterDevice(ctx context.Context, deviceID string) error {
	return uc.repo.UnregisterDevice(ctx, deviceID)
}

func (uc *PushUseCase) PushToUser(ctx context.Context, userID, title, body string, payload []byte) error {
	devices, err := uc.repo.FindDevicesByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("find devices: %w", err)
	}
	if len(devices) == 0 {
		return fmt.Errorf("no devices registered")
	}
	data := map[string]string{}
	if len(payload) > 0 {
		data["payload"] = string(payload)
	}
	// Group tokens by platform for per-platform push delivery
	byPlatform := map[int32][]string{}
	for _, d := range devices {
		byPlatform[d.Platform] = append(byPlatform[d.Platform], d.DeviceToken)
	}
	for platform, tokens := range byPlatform {
		if err := uc.pusher.Send(ctx, tokens, platform, title, body, data); err != nil {
			return err
		}
	}
	return nil
}

func (uc *PushUseCase) PushToTopic(ctx context.Context, topic, title, body string, payload []byte) error {
	data := map[string]string{}
	if len(payload) > 0 {
		data["payload"] = string(payload)
	}
	return uc.pusher.SendToTopic(ctx, topic, title, body, data)
}

// Platform constants (aliased from proto for convenience)
const (
	PlatformIOS     = int32(pushpb.DevicePlatform_DEVICE_PLATFORM_IOS)
	PlatformAndroid = int32(pushpb.DevicePlatform_DEVICE_PLATFORM_ANDROID)
	PlatformWeb     = int32(pushpb.DevicePlatform_DEVICE_PLATFORM_WEB)
)
