// Package logger provides a structured logging facade over log/slog
// with automatic OpenTelemetry trace context injection.
package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go.opentelemetry.io/otel/trace"
)

// traceHandler wraps a slog.Handler and enriches each record with
// trace_id and span_id from the OpenTelemetry span context.
type traceHandler struct {
	handler slog.Handler
}

func (h *traceHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *traceHandler) Handle(ctx context.Context, r slog.Record) error {
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.handler.Handle(ctx, r)
}

func (h *traceHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &traceHandler{handler: h.handler.WithAttrs(attrs)}
}

func (h *traceHandler) WithGroup(name string) slog.Handler {
	return &traceHandler{handler: h.handler.WithGroup(name)}
}

// New creates a *slog.Logger with trace ID enrichment for the given level.
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}
	base := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: lvl})
	return slog.New(&traceHandler{handler: base})
}

// Default is the default logger (level=info) with trace enrichment.
var Default = New("info")

// Infof logs an info message.
func Infof(format string, args ...any) {
	Default.LogAttrs(context.Background(), slog.LevelInfo, fmt.Sprintf(format, args...))
}

// Errorf logs an error message.
func Errorf(format string, args ...any) {
	Default.LogAttrs(context.Background(), slog.LevelError, fmt.Sprintf(format, args...))
}

// Warnf logs a warning message.
func Warnf(format string, args ...any) {
	Default.LogAttrs(context.Background(), slog.LevelWarn, fmt.Sprintf(format, args...))
}

// ContextInfof logs an info message with context (trace propagation).
func ContextInfof(ctx context.Context, format string, args ...any) {
	Default.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf(format, args...))
}

// ContextErrorf logs an error with context.
func ContextErrorf(ctx context.Context, format string, args ...any) {
	Default.LogAttrs(ctx, slog.LevelError, fmt.Sprintf(format, args...))
}

// ContextWarnf logs a warning with context.
func ContextWarnf(ctx context.Context, format string, args ...any) {
	Default.LogAttrs(ctx, slog.LevelWarn, fmt.Sprintf(format, args...))
}
