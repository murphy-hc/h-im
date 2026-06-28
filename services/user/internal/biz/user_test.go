package biz

import (
	"context"
	"testing"
	"time"

	"github.com/murphy-hc/h-im/pkg/jwt"
)

// mockUserRepo implements UserRepo for testing.
type mockUserRepo struct {
	users    map[string]*userRecord // username -> record
	byID     map[string]*userRecord // userID -> record
	findErr  error
	regErr   error
}

type userRecord struct {
	userID       string
	username     string
	passwordHash string
	nickname     string
	avatar       string
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users: make(map[string]*userRecord),
		byID:  make(map[string]*userRecord),
	}
}

func (m *mockUserRepo) SetOnline(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error {
	return nil
}
func (m *mockUserRepo) SetOffline(ctx context.Context, userID, deviceID string) error {
	return nil
}
func (m *mockUserRepo) GetUserOnline(ctx context.Context, userID string) ([]OnlineDevice, error) {
	return nil, nil
}
func (m *mockUserRepo) SweepOffline(ctx context.Context, timeoutSeconds int64) ([]OfflinePair, error) {
	return nil, nil
}
func (m *mockUserRepo) FindAppByID(ctx context.Context, appID string) (*App, error) {
	return &App{AppID: "app1", AppSecret: "secret"}, nil
}
func (m *mockUserRepo) FindByUserID(ctx context.Context, userID string) (*User, error) {
	r, ok := m.byID[userID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return &User{UserID: r.userID, Username: r.username, Nickname: r.nickname, Avatar: r.avatar}, nil
}
func (m *mockUserRepo) BatchGetUsers(ctx context.Context, userIDs []string) ([]*User, error) {
	var out []*User
	for _, id := range userIDs {
		if r, ok := m.byID[id]; ok {
			out = append(out, &User{UserID: r.userID, Username: r.username, Nickname: r.nickname, Avatar: r.avatar})
		}
	}
	return out, nil
}
func (m *mockUserRepo) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error {
	r, ok := m.byID[userID]
	if !ok {
		return context.DeadlineExceeded
	}
	if nickname != "" {
		r.nickname = nickname
	}
	if avatar != "" {
		r.avatar = avatar
	}
	return nil
}
func (m *mockUserRepo) Register(ctx context.Context, userID, username, passwordHash string) error {
	if m.regErr != nil {
		return m.regErr
	}
	r := &userRecord{userID: userID, username: username, passwordHash: passwordHash}
	m.users[username] = r
	m.byID[userID] = r
	return nil
}
func (m *mockUserRepo) FindByUsername(ctx context.Context, username string) (string, string, error) {
	r, ok := m.users[username]
	if !ok {
		return "", "", context.DeadlineExceeded
	}
	return r.userID, r.passwordHash, nil
}

func testJWTManager() *jwt.Manager {
	return jwt.NewManager("test-secret", 1*time.Hour, 24*time.Hour)
}

func TestRegisterSuccess(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	userID, err := uc.Register(context.Background(), "alice", "password123")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if userID == "" {
		t.Fatal("expected non-empty userID")
	}
}

func TestRegisterEmptyUsername(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	_, err := uc.Register(context.Background(), "", "password")
	if err == nil {
		t.Fatal("expected error for empty username")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	uc.Register(context.Background(), "bob", "password123")
	_, err := uc.Register(context.Background(), "bob", "another")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	uc.Register(context.Background(), "charlie", "mypassword")

	access, refresh, expiresAt, err := uc.Login(context.Background(), "charlie", "mypassword")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("empty tokens")
	}
	if expiresAt <= time.Now().Unix() {
		t.Fatal("expiresAt should be in the future")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	uc.Register(context.Background(), "dave", "correct")

	_, _, _, err := uc.Login(context.Background(), "dave", "wrong")
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	_, _, _, err := uc.Login(context.Background(), "nobody", "password")
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestGetProfile(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	id, _ := uc.Register(context.Background(), "eve", "password")

	user, err := uc.GetProfile(context.Background(), id)
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if user.Username != "eve" {
		t.Fatalf("expected username 'eve', got %q", user.Username)
	}
}

func TestBatchGetUsers(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	id1, _ := uc.Register(context.Background(), "user1", "pass")
	id2, _ := uc.Register(context.Background(), "user2", "pass")

	users, err := uc.BatchGetUsers(context.Background(), []string{id1, id2})
	if err != nil {
		t.Fatalf("BatchGetUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestUpdateProfile(t *testing.T) {
	repo := newMockUserRepo()
	uc := NewUserUseCase(repo, HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}, testJWTManager())

	id, _ := uc.Register(context.Background(), "frank", "pass")

	err := uc.UpdateProfile(context.Background(), id, "Frankie", "https://img/frank.jpg")
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	user, _ := uc.GetProfile(context.Background(), id)
	if user.Nickname != "Frankie" {
		t.Fatalf("expected nickname Frankie, got %q", user.Nickname)
	}
	if user.Avatar != "https://img/frank.jpg" {
		t.Fatalf("expected avatar URL, got %q", user.Avatar)
	}
}
