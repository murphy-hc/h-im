package data

import (
	"github.com/murphy-hc/h-im/services/push/internal/biz"
)

type pushRepo struct {
	data *Data
}

// NewPushRepo creates a PushRepo implementation.
func NewPushRepo(data *Data) biz.PushRepo {
	return &pushRepo{data: data}
}
