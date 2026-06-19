package data

import (
	"github.com/murphy-hc/h-im/services/group/internal/biz"
)

type groupRepo struct {
	data *Data
}

// NewGroupRepo creates a GroupRepo implementation.
func NewGroupRepo(data *Data) biz.GroupRepo {
	return &groupRepo{data: data}
}
