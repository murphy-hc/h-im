package data

import (
	"github.com/murphy-hc/h-im/services/contact/internal/biz"
)

type contactRepo struct {
	data *Data
}

// NewContactRepo creates a ContactRepo implementation.
func NewContactRepo(data *Data) biz.ContactRepo {
	return &contactRepo{data: data}
}
