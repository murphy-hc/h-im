package tracing

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const defaultTimeout = 5 * time.Second

// Transport is the OTLP transport protocol.
type Transport string

const (
	TransportHTTP Transport = "http"
	TransportGRPC Transport = "grpc"
)

// Config holds OpenTelemetry tracing configuration.
type Config struct {
	ServiceName string
	Rate        float64
	Endpoint    string
	Path        string
	Timeout     time.Duration
	Insecure    bool
	Transport   Transport
}

func (c *Config) timeout() time.Duration {
	if c.Timeout <= 0 {
		return defaultTimeout
	}
	return c.Timeout
}

func (c *Config) transport() Transport {
	if c.Transport == "" {
		return TransportHTTP
	}
	return c.Transport
}

// InitTracer initializes the OpenTelemetry tracer provider and returns a
// shutdown function.
func InitTracer(conf *Config) (func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.timeout())
	defer cancel()

	hostName, _ := os.Hostname()
	otelResource, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(conf.ServiceName),
			semconv.HostNameKey.String(hostName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("create otel resource: %w", err)
	}

	var traceExporter *otlptrace.Exporter
	switch conf.transport() {
	case TransportGRPC:
		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(conf.Endpoint),
			otlptracegrpc.WithCompressor("gzip"),
		}
		if conf.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		traceExporter, err = otlptrace.New(ctx, otlptracegrpc.NewClient(opts...))
	default:
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(conf.Endpoint),
			otlptracehttp.WithURLPath(conf.Path),
			otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		}
		if conf.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		traceExporter, err = otlptrace.New(ctx, otlptracehttp.NewClient(opts...))
	}
	if err != nil {
		return nil, fmt.Errorf("create trace exporter: %w", err)
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(conf.Rate)),
		tracesdk.WithResource(otelResource),
		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(traceExporter)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return func() {
		cxt, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := traceExporter.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}, nil
}

// FinishSpan ends a span with stack trace and timestamp.
func FinishSpan(span trace.Span) {
	span.End(trace.WithStackTrace(true), trace.WithTimestamp(time.Now()))
}

// Tracer returns a tracer instance for the given service name.
func Tracer(name string) trace.Tracer {
	return otel.Tracer(name)
}
