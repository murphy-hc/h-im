package biz

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

var ProviderSet = wire.NewSet(NewGatewayUseCase, NewHeartbeatConfig)

// NewHeartbeatConfig extracts heartbeat configuration from bootstrap.
func NewHeartbeatConfig(bc *conf.Bootstrap) HeartbeatConfig {
	hb := bc.GetHeartbeat()
	return HeartbeatConfig{
		IntervalSeconds: hb.GetIntervalSeconds(),
		TimeoutSeconds:  hb.GetTimeoutSeconds(),
		SweepInterval:   hb.GetSweepInterval(),
	}
}
