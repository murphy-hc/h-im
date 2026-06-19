package data

import (
	"github.com/murphy-hc/h-im/services/user/internal/biz"
)

type userRepo struct {
	data *Data
}

// NewUserRepo creates a UserRepo implementation.
func NewUserRepo(data *Data) biz.UserRepo {
	return &userRepo{data: data}
}
