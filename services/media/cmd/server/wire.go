//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	"github.com/murphy-hc/h-im/services/media/internal/biz"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	"github.com/murphy-hc/h-im/services/media/internal/data"
	"github.com/murphy-hc/h-im/services/media/internal/server"
	"github.com/murphy-hc/h-im/services/media/internal/service"
)

func wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.GRPCProviderSet,
		server.HTTPProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		newApp,
	))
}
