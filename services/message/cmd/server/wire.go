//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/google/wire"
	"go.opentelemetry.io/otel/metric"

	"github.com/murphy-hc/h-im/services/message/internal/biz"
	"github.com/murphy-hc/h-im/services/message/internal/conf"
	"github.com/murphy-hc/h-im/services/message/internal/data"
	"github.com/murphy-hc/h-im/services/message/internal/server"
	"github.com/murphy-hc/h-im/services/message/internal/service"
)

func wireApp(bc *conf.Bootstrap, meter metric.Meter) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.GRPCProviderSet,
		server.HTTPProviderSet,
		server.KafkaProviderSet,
		service.ProviderSet,
		biz.ProviderSet,
		data.ProviderSet,
		server.NewSequenceClient,
		newApp,
	))
}
