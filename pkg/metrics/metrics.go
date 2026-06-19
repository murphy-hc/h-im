package metrics

import (
	"fmt"

	"github.com/go-kratos/kratos/v2/middleware/metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// NewMeterProvider creates a MeterProvider with the given reader, service name,
// and environment and registers it as the global provider. Returns the meter
// for the named service.
func NewMeterProvider(reader sdkmetric.Reader, name, env string) (metric.Meter, error) {
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(name),
				attribute.String("environment", env),
			),
		),
		sdkmetric.WithView(
			metrics.DefaultSecondsHistogramView(metrics.DefaultServerSecondsHistogramName),
		),
	)
	otel.SetMeterProvider(provider)

	return provider.Meter(name), nil
}

// NewPrometheusMeter is a convenience helper that creates a Prometheus-backed
// meter. For other exporters, use NewMeterProvider with a custom reader.
func NewPrometheusMeter(name, env string) (metric.Meter, error) {
	exporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("create prometheus exporter: %w", err)
	}
	return NewMeterProvider(exporter, name, env)
}
