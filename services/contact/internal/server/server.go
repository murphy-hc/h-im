package server

import (
	"github.com/google/wire"
	"github.com/murphy-hc/h-im/services/contact/internal/conf"
)

var GRPCProviderSet = wire.NewSet(NewGRPCServer, wire.FieldsOf(new(*conf.Bootstrap), "Server"))
var HTTPProviderSet = wire.NewSet(NewHTTPServer)
