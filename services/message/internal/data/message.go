package data

import (
	"github.com/murphy-hc/h-im/services/message/internal/biz"
)

type messageRepo struct {
	data *Data
}

// NewMessageRepo creates a MessageRepo implementation.
func NewMessageRepo(data *Data) biz.MessageRepo {
	return &messageRepo{data: data}
}
