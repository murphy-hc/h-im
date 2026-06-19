//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/google/wire"

	"github.com/murphy-hc/h-im/services/sequence/internal/biz"
	"github.com/murphy-hc/h-im/services/sequence/internal/conf"
	"github.com/murphy-hc/h-im/services/sequence/internal/data"
	"github.com/murphy-hc/h-im/services/sequence/internal/server"
	"github.com/murphy-hc/h-im/services/sequence/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}

func newApp(gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.Server(gs),
	)
}
