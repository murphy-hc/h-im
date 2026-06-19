//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
	"github.com/murphy-hc/h-im/services/contact/internal/conf"
	"github.com/murphy-hc/h-im/services/contact/internal/data"
	"github.com/murphy-hc/h-im/services/contact/internal/server"
	"github.com/murphy-hc/h-im/services/contact/internal/service"
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
