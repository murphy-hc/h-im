package jwt

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

type contextKey string

const userIDKey contextKey = "user_id"

// UserIDFromContext returns the authenticated user ID from context.
func UserIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userIDKey).(string); ok {
		return v
	}
	return ""
}

// Server returns a gRPC middleware that validates JWT tokens.
// Methods listed in whitelist are allowed without authentication.
func Server(mgr *Manager, whitelist ...string) middleware.Middleware {
	allow := make(map[string]bool, len(whitelist))
	for _, method := range whitelist {
		allow[method] = true
	}

	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				operation := tr.Operation()
				if allow[operation] {
					return handler(ctx, req)
				}
			}

			token := extractToken(ctx)
			if token == "" {
				return nil, ErrMissingToken
			}

			userID, err := mgr.Validate(token)
			if err != nil {
				return nil, ErrInvalidToken
			}

			ctx = context.WithValue(ctx, userIDKey, userID)
			return handler(ctx, req)
		}
	}
}

func extractToken(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		auth := tr.RequestHeader().Get("authorization")
		if auth == "" {
			return ""
		}
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
		return auth
	}
	return ""
}
