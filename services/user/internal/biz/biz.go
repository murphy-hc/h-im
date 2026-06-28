package biz

import (
	"os"
	"time"

	"github.com/google/wire"
	"github.com/murphy-hc/h-im/pkg/jwt"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUserUseCase, NewHeartbeatConfig, NewAuthUseCase, NewJWTManager)

// NewHeartbeatConfig extracts heartbeat configuration.
func NewHeartbeatConfig(bc *conf.Bootstrap) HeartbeatConfig {
	hb := bc.GetHeartbeat()
	return HeartbeatConfig{
		TimeoutSeconds: hb.GetTimeoutSeconds(),
		SweepInterval:  hb.GetSweepInterval(),
	}
}

// NewJWTManager creates a JWT Manager from config or env.
func NewJWTManager(bc *conf.Bootstrap) *jwt.Manager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "change-me-in-production"
	}
	accessTTL := 24 * time.Hour
	refreshTTL := 7 * 24 * time.Hour
	if bc.GetJwt() != nil {
		if bc.GetJwt().GetAccessTtl().GetSeconds() > 0 {
			accessTTL = time.Duration(bc.GetJwt().GetAccessTtl().GetSeconds()) * time.Second
		}
		if bc.GetJwt().GetRefreshTtl().GetSeconds() > 0 {
			refreshTTL = time.Duration(bc.GetJwt().GetRefreshTtl().GetSeconds()) * time.Second
		}
	}
	return jwt.NewManager(secret, accessTTL, refreshTTL)
}
