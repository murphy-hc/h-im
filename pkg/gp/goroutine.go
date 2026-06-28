package gp

import (
	"context"
	"runtime/debug"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/trace"
)

// SafeGo runs f in a new goroutine with panic recovery. The goroutine receives
// a context detached from the caller's lifecycle (background context with only
// the OpenTelemetry span propagated), so it survives cancellation/deadline.
func SafeGo(ctx context.Context, f func(ctx context.Context)) {
	bgCtx := trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Context(bgCtx).Errorf( "SafeGo panic: %s", string(debug.Stack()))
			}
		}()
		f(bgCtx)
	}()
}
