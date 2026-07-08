package server

import (
	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

// NewHeartbeatConfig extracts heartbeat configuration from bootstrap.
// Lives in server layer to keep biz free of conf dependency.
func NewHeartbeatConfig(bc *conf.Bootstrap) biz.HeartbeatConfig {
	hb := bc.GetHeartbeat()
	return biz.HeartbeatConfig{
		IntervalSeconds: hb.GetIntervalSeconds(),
		TimeoutSeconds:  hb.GetTimeoutSeconds(),
		SweepInterval:   hb.GetSweepInterval(),
	}
}
