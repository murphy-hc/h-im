//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/murphy-hc/h-im/services/group/internal/biz"
	"github.com/murphy-hc/h-im/services/group/internal/conf"
	"github.com/murphy-hc/h-im/services/group/internal/data"
	"github.com/murphy-hc/h-im/services/group/internal/server"
	"github.com/murphy-hc/h-im/services/group/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*grpc.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}
