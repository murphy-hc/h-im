package biz

import (
	"context"
	"testing"
)

type mockPushRepo struct {
	devices map[string][]DeviceInfo
}

func newMockPushRepo() *mockPushRepo {
	return &mockPushRepo{devices: make(map[string][]DeviceInfo)}
}

func (m *mockPushRepo) RegisterDevice(ctx context.Context, userID string, info *DeviceInfo) error {
	m.devices[userID] = append(m.devices[userID], *info)
	return nil
}
func (m *mockPushRepo) UnregisterDevice(ctx context.Context, deviceID string) error {
	return nil
}
func (m *mockPushRepo) FindDevicesByUser(ctx context.Context, userID string) ([]DeviceInfo, error) {
	return m.devices[userID], nil
}

type mockPusher struct {
	sent         int
	topicSends   int
	lastPlatform int32
}

func (p *mockPusher) Send(ctx context.Context, tokens []string, platform int32, title, body string, data map[string]string) error {
	p.sent += len(tokens)
	p.lastPlatform = platform
	return nil
}
func (p *mockPusher) SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error {
	p.topicSends++
	return nil
}

func TestRegisterDevice(t *testing.T) {
	repo := newMockPushRepo()
	pusher := &mockPusher{}
	uc := NewPushUseCase(repo, pusher)

	err := uc.RegisterDevice(context.Background(), "user-1", &DeviceInfo{
		DeviceID: "dev-1", DeviceToken: "token-abc", Platform: PlatformIOS,
	})
	if err != nil {
		t.Fatalf("RegisterDevice: %v", err)
	}

	devices, _ := repo.FindDevicesByUser(context.Background(), "user-1")
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
}

func TestPushToUser(t *testing.T) {
	repo := newMockPushRepo()
	pusher := &mockPusher{}
	uc := NewPushUseCase(repo, pusher)

	uc.RegisterDevice(context.Background(), "user-1", &DeviceInfo{
		DeviceID: "dev-1", DeviceToken: "ios-token", Platform: PlatformIOS,
	})
	uc.RegisterDevice(context.Background(), "user-1", &DeviceInfo{
		DeviceID: "dev-2", DeviceToken: "android-token", Platform: PlatformAndroid,
	})

	err := uc.PushToUser(context.Background(), "user-1", "Title", "Body", []byte(`{"key":"val"}`))
	if err != nil {
		t.Fatalf("PushToUser: %v", err)
	}
	if pusher.sent != 2 {
		t.Fatalf("expected 2 pushes, got %d", pusher.sent)
	}
}

func TestPushToUserNoDevices(t *testing.T) {
	repo := newMockPushRepo()
	pusher := &mockPusher{}
	uc := NewPushUseCase(repo, pusher)

	err := uc.PushToUser(context.Background(), "no-devices", "Title", "Body", nil)
	if err == nil {
		t.Fatal("expected error for user with no devices")
	}
}

func TestPushToTopic(t *testing.T) {
	repo := newMockPushRepo()
	pusher := &mockPusher{}
	uc := NewPushUseCase(repo, pusher)

	err := uc.PushToTopic(context.Background(), "news", "Breaking", "Something happened", nil)
	if err != nil {
		t.Fatalf("PushToTopic: %v", err)
	}
	if pusher.topicSends != 1 {
		t.Fatal("expected 1 topic send")
	}
}
