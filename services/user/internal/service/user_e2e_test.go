package service

import (
	"context"
	"testing"
	"time"

	pb "github.com/murphy-hc/h-im/gen/go/him/user/v1"
	"github.com/murphy-hc/h-im/pkg/jwt"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
)

// e2eUserRepo is an in-memory repo for end-to-end service tests.
type e2eUserRepo struct {
	users  map[string]*userRow // username -> row
	byID   map[string]*userRow // userID -> row
}

type userRow struct {
	userID, username, passwordHash, nickname, avatar string
}

func newE2EUserRepo() *e2eUserRepo {
	return &e2eUserRepo{
		users: make(map[string]*userRow),
		byID:  make(map[string]*userRow),
	}
}

func (r *e2eUserRepo) SetOnline(_ context.Context, _, _, _ string, _ int64) error { return nil }
func (r *e2eUserRepo) SetOffline(_ context.Context, _, _ string) error             { return nil }
func (r *e2eUserRepo) GetUserOnline(_ context.Context, _ string) ([]biz.OnlineDevice, error) {
	return nil, nil
}
func (r *e2eUserRepo) SweepOffline(_ context.Context, _ int64) ([]biz.OfflinePair, error) {
	return nil, nil
}
func (r *e2eUserRepo) FindAppByID(_ context.Context, _ string) (*biz.App, error) {
	return &biz.App{AppID: "app1", AppSecret: "secret"}, nil
}
func (r *e2eUserRepo) FindByUsername(_ context.Context, username string) (string, string, error) {
	row, ok := r.users[username]
	if !ok {
		return "", "", context.DeadlineExceeded
	}
	return row.userID, row.passwordHash, nil
}
func (r *e2eUserRepo) FindByUserID(_ context.Context, userID string) (*biz.User, error) {
	row, ok := r.byID[userID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return &biz.User{UserID: row.userID, Username: row.username, Nickname: row.nickname, Avatar: row.avatar}, nil
}
func (r *e2eUserRepo) BatchGetUsers(_ context.Context, userIDs []string) ([]*biz.User, error) {
	var out []*biz.User
	for _, id := range userIDs {
		if row, ok := r.byID[id]; ok {
			out = append(out, &biz.User{UserID: row.userID, Username: row.username, Nickname: row.nickname, Avatar: row.avatar})
		}
	}
	return out, nil
}
func (r *e2eUserRepo) UpdateProfile(_ context.Context, userID, nickname, avatar string) error {
	row := r.byID[userID]
	if row == nil {
		return context.DeadlineExceeded
	}
	if nickname != "" {
		row.nickname = nickname
	}
	if avatar != "" {
		row.avatar = avatar
	}
	return nil
}
func (r *e2eUserRepo) Register(_ context.Context, userID, username, passwordHash string) error {
	row := &userRow{userID: userID, username: username, passwordHash: passwordHash}
	r.users[username] = row
	r.byID[userID] = row
	return nil
}

func newE2EUserService() *UserService {
	repo := newE2EUserRepo()
	jwtMgr := jwt.NewManager("e2e-test-secret", 1*time.Hour, 24*time.Hour)
	hbCfg := biz.HeartbeatConfig{TimeoutSeconds: 180, SweepInterval: 10}
	uc := biz.NewUserUseCase(repo, hbCfg, jwtMgr)
	authUC := biz.NewAuthUseCase(repo)
	return NewUserService(uc, authUC)
}

func TestE2E_RegisterLoginProfile(t *testing.T) {
	svc := newE2EUserService()
	ctx := context.Background()

	// 1. Register
	regResp, err := svc.Register(ctx, &pb.RegisterRequest{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if regResp.UserId == "" {
		t.Fatal("expected user ID")
	}

	// 2. Login
	loginResp, err := svc.Login(ctx, &pb.LoginRequest{Username: "alice", Password: "secret123"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if loginResp.AccessToken == "" || loginResp.RefreshToken == "" {
		t.Fatal("empty tokens")
	}
	if loginResp.ExpiresAt <= time.Now().Unix() {
		t.Fatal("token should not be expired")
	}

	// 3. GetProfile
	profile, err := svc.GetProfile(ctx, &pb.GetProfileRequest{UserId: regResp.UserId})
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if profile.User.UserId != regResp.UserId {
		t.Fatal("user ID mismatch")
	}

	// 4. UpdateProfile
	_, err = svc.UpdateProfile(ctx, &pb.UpdateProfileRequest{
		UserId: regResp.UserId, Nickname: "Alice", Avatar: "https://img/a.jpg",
	})
	if err != nil {
		t.Fatalf("UpdateProfile: %v", err)
	}

	// 5. Verify update
	profile, _ = svc.GetProfile(ctx, &pb.GetProfileRequest{UserId: regResp.UserId})
	if profile.User.Nickname != "Alice" {
		t.Fatalf("expected nickname Alice, got %q", profile.User.Nickname)
	}
}

func TestE2E_RegisterDuplicate(t *testing.T) {
	svc := newE2EUserService()
	ctx := context.Background()

	svc.Register(ctx, &pb.RegisterRequest{Username: "bob", Password: "pass"})
	_, err := svc.Register(ctx, &pb.RegisterRequest{Username: "bob", Password: "another"})
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestE2E_LoginWrongPassword(t *testing.T) {
	svc := newE2EUserService()
	ctx := context.Background()

	svc.Register(ctx, &pb.RegisterRequest{Username: "charlie", Password: "correct"})
	_, err := svc.Login(ctx, &pb.LoginRequest{Username: "charlie", Password: "wrong"})
	if err == nil {
		t.Fatal("expected login error for wrong password")
	}
}

func TestE2E_BatchGetUsers(t *testing.T) {
	svc := newE2EUserService()
	ctx := context.Background()

	r1, _ := svc.Register(ctx, &pb.RegisterRequest{Username: "user1", Password: "pass"})
	r2, _ := svc.Register(ctx, &pb.RegisterRequest{Username: "user2", Password: "pass"})

	resp, err := svc.BatchGetUsers(ctx, &pb.BatchGetUsersRequest{UserIds: []string{r1.UserId, r2.UserId}})
	if err != nil {
		t.Fatalf("BatchGetUsers: %v", err)
	}
	if len(resp.Users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(resp.Users))
	}
}
