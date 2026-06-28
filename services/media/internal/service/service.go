package service

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewMediaService, NewMediaHTTPHandler, NewMediaSecret)

func NewMediaSecret(bc *conf.Bootstrap) string {
	return bc.GetMediaSecret()
}
