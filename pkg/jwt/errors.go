package jwt

import "github.com/go-kratos/kratos/v2/errors"

var (
	ErrMissingToken = errors.New(401, "JWT_MISSING", "authorization token is missing")
	ErrInvalidToken = errors.New(401, "JWT_INVALID", "authorization token is invalid")
)
