package jwt

import (
	"context"
	"os"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var (
	ErrServiceAuthMissing = errors.New(401, "SERVICE_AUTH_MISSING", "inter-service authorization token is missing")
	ErrServiceAuthInvalid = errors.New(401, "SERVICE_AUTH_INVALID", "inter-service authorization token is invalid")
)

// ServiceAuth returns a middleware that validates a shared inter-service token.
// The token is read from SERVICE_AUTH_TOKEN env var. If not set, auth is skipped
// (dev mode). In production, all services must share the same token.
func ServiceAuth() middleware.Middleware {
	expected := os.Getenv("SERVICE_AUTH_TOKEN")
	if expected == "" {
		// Dev mode: no inter-service auth
		return func(handler middleware.Handler) middleware.Handler {
			return func(ctx context.Context, req interface{}) (interface{}, error) {
				return handler(ctx, req)
			}
		}
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				token := tr.RequestHeader().Get("x-service-token")
				if token == "" {
					return nil, ErrServiceAuthMissing
				}
				if token != expected {
					return nil, ErrServiceAuthInvalid
				}
			}
			return handler(ctx, req)
		}
	}
}
