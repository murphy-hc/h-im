//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/murphy-hc/h-im/services/auth/internal/biz"
	"github.com/murphy-hc/h-im/services/auth/internal/conf"
	"github.com/murphy-hc/h-im/services/auth/internal/data"
	"github.com/murphy-hc/h-im/services/auth/internal/server"
	"github.com/murphy-hc/h-im/services/auth/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*grpc.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}
