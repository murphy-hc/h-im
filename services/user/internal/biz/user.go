package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/murphy-hc/h-im/pkg/gp"
	"github.com/murphy-hc/h-im/pkg/jwt"
	"github.com/rs/xid"
	"golang.org/x/crypto/bcrypt"
)

// HeartbeatConfig holds heartbeat parameters for the user service.
type HeartbeatConfig struct {
	TimeoutSeconds int32
	SweepInterval  int32
}

// UserUseCase handles user business logic.
type UserUseCase struct {
	repo       UserRepo
	hbCfg      HeartbeatConfig
	jwtManager *jwt.Manager
}

// NewUserUseCase creates a UserUseCase.
func NewUserUseCase(repo UserRepo, hbCfg HeartbeatConfig, jwtManager *jwt.Manager) *UserUseCase {
	uc := &UserUseCase{repo: repo, hbCfg: hbCfg, jwtManager: jwtManager}
	gp.SafeGo(context.Background(), func(_ context.Context) { uc.sweepLoop() })
	return uc
}

func (uc *UserUseCase) sweepLoop() {
	interval := time.Duration(uc.hbCfg.SweepInterval) * time.Second
	if interval <= 0 {
		interval = 10 * time.Second
	}
	timeout := int64(uc.hbCfg.TimeoutSeconds)
	if timeout <= 0 {
		timeout = 180
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		uc.repo.SweepOffline(context.Background(), timeout)
	}
}

// ReportHeartbeat records a successful heartbeat from a gateway.
func (uc *UserUseCase) ReportHeartbeat(ctx context.Context, userID, deviceID, gatewayAddr string, timestamp int64) error {
	return uc.repo.SetOnline(ctx, userID, deviceID, gatewayAddr, timestamp)
}

// ReportDisconnect records a device disconnection.
func (uc *UserUseCase) ReportDisconnect(ctx context.Context, userID, deviceID string) error {
	return uc.repo.SetOffline(ctx, userID, deviceID)
}

// GetUserOnline returns all online devices for a user.
func (uc *UserUseCase) GetUserOnline(ctx context.Context, userID string) ([]OnlineDevice, error) {
	return uc.repo.GetUserOnline(ctx, userID)
}

// Register creates a new user account.
func (uc *UserUseCase) Register(ctx context.Context, username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password required")
	}
	if _, _, err := uc.repo.FindByUsername(ctx, username); err == nil {
		return "", fmt.Errorf("username already exists")
	}
	userID := xid.New().String()
	hash, err := hashPassword(password)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	if err := uc.repo.Register(ctx, userID, username, hash); err != nil {
		return "", fmt.Errorf("register: %w", err)
	}
	return userID, nil
}

// Login authenticates a user with username and password, returning JWT tokens.
func (uc *UserUseCase) Login(ctx context.Context, username, password string) (accessToken, refreshToken string, expiresAt int64, err error) {
	userID, hash, err := uc.repo.FindByUsername(ctx, username)
	if err != nil {
		return "", "", 0, fmt.Errorf("user not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return "", "", 0, fmt.Errorf("invalid password: %w", err)
	}
	accessToken, err = uc.jwtManager.IssueAccessToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("issue access token: %w", err)
	}
	refreshToken, err = uc.jwtManager.IssueRefreshToken(userID)
	if err != nil {
		return "", "", 0, fmt.Errorf("issue refresh token: %w", err)
	}
	expiresAt = time.Now().Add(uc.jwtManager.AccessTTL()).Unix()
	return accessToken, refreshToken, expiresAt, nil
}

// GetProfile returns a user's profile by ID.
func (uc *UserUseCase) GetProfile(ctx context.Context, userID string) (*User, error) {
	return uc.repo.FindByUserID(ctx, userID)
}

// UpdateProfile updates a user's nickname and/or avatar.
func (uc *UserUseCase) UpdateProfile(ctx context.Context, userID, nickname, avatar string) error {
	return uc.repo.UpdateProfile(ctx, userID, nickname, avatar)
}

// BatchGetUsers returns profiles for multiple user IDs.
func (uc *UserUseCase) BatchGetUsers(ctx context.Context, userIDs []string) ([]*User, error) {
	return uc.repo.BatchGetUsers(ctx, userIDs)
}


func hashPassword(pw string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("bcrypt: %w", err)
	}
	return string(hash), nil
}
