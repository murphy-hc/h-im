package data

import (
	"github.com/murphy-hc/h-im/services/media/internal/biz"
)

type mediaRepo struct {
	data *Data
}

// NewMediaRepo creates a MediaRepo implementation.
func NewMediaRepo(data *Data) biz.MediaRepo {
	return &mediaRepo{data: data}
}
