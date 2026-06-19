//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"

	"github.com/murphy-hc/h-im/services/gateway/internal/biz"
	"github.com/murphy-hc/h-im/services/gateway/internal/conf"
	"github.com/murphy-hc/h-im/services/gateway/internal/server"
	"github.com/murphy-hc/h-im/services/gateway/internal/service"
)

func wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.WSServerProviderSet,
		server.HTTPProviderSet,
		server.GRPCProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		newApp,
	))
}
