package data

import (
	"github.com/murphy-hc/h-im/services/auth/internal/biz"
)

type authRepo struct {
	data *Data
}

// NewAuthRepo creates a AuthRepo implementation.
func NewAuthRepo(data *Data) biz.AuthRepo {
	return &authRepo{data: data}
}
