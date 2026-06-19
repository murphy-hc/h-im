package server

import (
	"github.com/google/wire"

	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
)

// GRPCProviderSet is providers for the gRPC server.
var GRPCProviderSet = wire.NewSet(NewGRPCServer, wire.FieldsOf(new(*conf.Bootstrap), "Server"))

// HTTPProviderSet is providers for the HTTP server (metrics/health).
var HTTPProviderSet = wire.NewSet(NewHTTPServer)
