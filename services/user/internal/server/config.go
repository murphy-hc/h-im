package server

import (
	"fmt"
	"os"
	"time"

	"github.com/murphy-hc/h-im/pkg/jwt"
	"github.com/murphy-hc/h-im/services/user/internal/biz"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
)

// NewHeartbeatConfig extracts heartbeat configuration from bootstrap.
// Lives in server layer to keep biz free of conf dependency.
func NewHeartbeatConfig(bc *conf.Bootstrap) biz.HeartbeatConfig {
	hb := bc.GetHeartbeat()
	return biz.HeartbeatConfig{
		TimeoutSeconds: hb.GetTimeoutSeconds(),
		SweepInterval:  hb.GetSweepInterval(),
	}
}

// NewJWTManager creates a JWT Manager from environment.
// Lives in server layer to keep biz free of conf/os dependencies.
func NewJWTManager(bc *conf.Bootstrap) *jwt.Manager {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" || secret == "change-me-in-production" {
		panic(fmt.Errorf("JWT_SECRET environment variable is not set or using insecure default — refusing to start"))
	}
	accessTTL := 24 * time.Hour
	refreshTTL := 7 * 24 * time.Hour
	return jwt.NewManager(secret, accessTTL, refreshTTL)
}
