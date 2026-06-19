package data

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewPushRepo)

// Data holds all data source clients.
type Data struct {
	// TODO: add db, redis, mq clients
}

// NewData creates a Data instance.
func NewData() (*Data, func(), error) {
	d := &Data{}
	cleanup := func() {
		// TODO: close connections
	}
	return d, cleanup, nil
}

var _ biz.PushRepo = (*pushRepo)(nil)
