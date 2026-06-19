//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/murphy-hc/h-im/services/media/internal/biz"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	"github.com/murphy-hc/h-im/services/media/internal/data"
	"github.com/murphy-hc/h-im/services/media/internal/server"
	"github.com/murphy-hc/h-im/services/media/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*grpc.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}
