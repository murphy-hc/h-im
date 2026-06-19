package tracing

import (
	"context"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// TracingConfig holds OpenTelemetry tracing configuration.
type TracingConfig struct {
	ServiceName string  // application name
	Rate        float64 // sample rate
	Endpoint    string  // collector endpoint
	Path        string  // collector path
}

// InitTracer initializes the OpenTelemetry tracer provider and returns a
// shutdown function. Callers should defer the returned function.
func InitTracer(conf *TracingConfig) func() {
	ctx := context.Background()
	var traceExporter *otlptrace.Exporter
	var batchSpanProcessor tracesdk.SpanProcessor
	traceExporter, batchSpanProcessor = newHTTPExporterAndSpanProcessor(ctx, conf)
	otelResource := newResource(ctx, conf)

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(conf.Rate)),
		tracesdk.WithResource(otelResource),
		tracesdk.WithSpanProcessor(batchSpanProcessor),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() {
		cxt, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := traceExporter.Shutdown(cxt); err != nil {
			otel.Handle(err)
		}
	}
}

// GetTracer returns a tracer instance.
func GetTracer() trace.Tracer {
	return otel.Tracer("kratos")
}

// FinishSpan ends a span with stack trace and timestamp.
func FinishSpan(span trace.Span) {
	span.End(trace.WithStackTrace(true), trace.WithTimestamp(time.Now()))
}

func newResource(ctx context.Context, conf *TracingConfig) *resource.Resource {
	hostName, _ := os.Hostname()

	r, err := resource.New(
		ctx,
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
		log.Fatalf("%s: %v", "Failed to create OpenTelemetry resource", err)
	}
	return r
}

func newHTTPExporterAndSpanProcessor(ctx context.Context, conf *TracingConfig) (*otlptrace.Exporter, tracesdk.SpanProcessor) {
	traceExporter, err := otlptrace.New(
		ctx,
		otlptracehttp.NewClient(
			otlptracehttp.WithEndpoint(conf.Endpoint),
			otlptracehttp.WithURLPath(conf.Path),
			otlptracehttp.WithInsecure(),
			otlptracehttp.WithCompression(otlptracehttp.GzipCompression),
		),
	)

	if err != nil {
		log.Fatalf("%s: %v", "Failed to create the OpenTelemetry trace exporter", err)
	}

	batchSpanProcessor := tracesdk.NewBatchSpanProcessor(traceExporter)

	return traceExporter, batchSpanProcessor
}
