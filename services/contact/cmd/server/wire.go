//go:build wireinject
// +build wireinject

package main

import (
	"github.com/google/wire"
	"google.golang.org/grpc"

	"github.com/murphy-hc/h-im/services/contact/internal/biz"
	"github.com/murphy-hc/h-im/services/contact/internal/conf"
	"github.com/murphy-hc/h-im/services/contact/internal/data"
	"github.com/murphy-hc/h-im/services/contact/internal/server"
	"github.com/murphy-hc/h-im/services/contact/internal/service"
)

func wireApp(*conf.Server, *conf.Data) (*grpc.Server, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
	))
}
