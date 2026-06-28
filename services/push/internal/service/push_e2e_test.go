package service

import (
	"context"
	"testing"

	pb "github.com/murphy-hc/h-im/gen/go/him/push/v1"
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

type e2ePushRepo struct {
	devices map[string][]biz.DeviceInfo
}

func newE2EPushRepo() *e2ePushRepo {
	return &e2ePushRepo{devices: make(map[string][]biz.DeviceInfo)}
}

func (r *e2ePushRepo) RegisterDevice(_ context.Context, userID string, info *biz.DeviceInfo) error {
	r.devices[userID] = append(r.devices[userID], *info)
	return nil
}
func (r *e2ePushRepo) UnregisterDevice(_ context.Context, deviceID string) error { return nil }
func (r *e2ePushRepo) FindDevicesByUser(_ context.Context, userID string) ([]biz.DeviceInfo, error) {
	return r.devices[userID], nil
}

type e2ePusher struct{ sent int }

func (p *e2ePusher) Send(_ context.Context, tokens []string, _ int32, _, _ string, _ map[string]string) error {
	p.sent += len(tokens)
	return nil
}
func (p *e2ePusher) SendToTopic(_ context.Context, _, _, _ string, _ map[string]string) error {
	return nil
}

func newE2EPushService() *PushService {
	repo := newE2EPushRepo()
	pusher := &e2ePusher{}
	return NewPushService(biz.NewPushUseCase(repo, pusher))
}

func TestE2E_RegisterAndPush(t *testing.T) {
	svc := newE2EPushService()
	ctx := context.Background()

	// Register devices
	_, err := svc.RegisterDevice(ctx, &pb.RegisterDeviceRequest{
		UserId: "user-1",
		Device: &pb.DeviceInfo{DeviceId: "dev-1", DeviceToken: "tok-1", Platform: pb.DevicePlatform_DEVICE_PLATFORM_IOS},
	})
	if err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	_, err = svc.RegisterDevice(ctx, &pb.RegisterDeviceRequest{
		UserId: "user-1",
		Device: &pb.DeviceInfo{DeviceId: "dev-2", DeviceToken: "tok-2", Platform: pb.DevicePlatform_DEVICE_PLATFORM_ANDROID},
	})
	if err != nil {
		t.Fatalf("RegisterDevice 2: %v", err)
	}

	// Push to user
	_, err = svc.PushToUser(ctx, &pb.PushToUserRequest{
		UserId: "user-1", Title: "Hello", Body: "World", Payload: "{}",
	})
	if err != nil {
		t.Fatalf("PushToUser: %v", err)
	}

	// Unregister
	_, err = svc.UnregisterDevice(ctx, &pb.UnregisterDeviceRequest{DeviceId: "dev-1"})
	if err != nil {
		t.Fatalf("UnregisterDevice: %v", err)
	}
}

func TestE2E_PushToTopic(t *testing.T) {
	svc := newE2EPushService()
	ctx := context.Background()

	_, err := svc.PushToTopic(ctx, &pb.PushToTopicRequest{
		Topic: "news", Title: "Breaking", Body: "News", Payload: "{}",
	})
	if err != nil {
		t.Fatalf("PushToTopic: %v", err)
	}
}
