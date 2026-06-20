package server

import (
	"github.com/google/wire"

	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
)

var WSServerProviderSet = wire.NewSet(
	NewWSServer,
	NewUpgrader,
	wire.FieldsOf(new(*conf.Bootstrap), "Server", "User"),
	wire.FieldsOf(new(*conf.Server), "Ws"),
)
var HTTPProviderSet = wire.NewSet(NewHTTPServer)
var GRPCProviderSet = wire.NewSet(NewGRPCServer)
