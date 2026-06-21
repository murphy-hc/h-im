package biz

import (
	"context"
	"time"
)

// HeartbeatConfig holds heartbeat parameters for the user service.
type HeartbeatConfig struct {
	TimeoutSeconds int32
	SweepInterval  int32
}

// UserUseCase handles user business logic.
type UserUseCase struct {
	repo  UserRepo
	hbCfg HeartbeatConfig
}

// NewUserUseCase creates a UserUseCase.
func NewUserUseCase(repo UserRepo, hbCfg HeartbeatConfig) *UserUseCase {
	uc := &UserUseCase{repo: repo, hbCfg: hbCfg}
	go uc.sweepLoop()
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
