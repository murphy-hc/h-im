// Package logger provides a thin structured logging facade over log/slog.
package logger

import (
	"log/slog"
	"os"
)

// New creates a *slog.Logger configured for the given level.
// Set level to "debug" for verbose output; defaults to "info".
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

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})
	return slog.New(handler)
}

// WithService attaches a service name to the logger.
func WithService(logger *slog.Logger, name string) *slog.Logger {
	return logger.With("service", name)
}
