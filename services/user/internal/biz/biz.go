package biz

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/user/internal/conf"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewUserUseCase, NewHeartbeatConfig, NewAuthUseCase)

// NewHeartbeatConfig extracts heartbeat configuration.
func NewHeartbeatConfig(bc *conf.Bootstrap) HeartbeatConfig {
	hb := bc.GetHeartbeat()
	return HeartbeatConfig{
		TimeoutSeconds: hb.GetTimeoutSeconds(),
		SweepInterval:  hb.GetSweepInterval(),
	}
}
