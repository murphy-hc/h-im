package data

import (
	"context"

	"github.com/murphy-hc/h-im/services/media/internal/biz"
)

var _ biz.MediaRepo = (*mediaRepo)(nil)

type mediaRepo struct {
	data *Data
}

// NewMediaRepo creates a MediaRepo implementation.
func NewMediaRepo(data *Data) biz.MediaRepo {
	return &mediaRepo{data: data}
}

func (r *mediaRepo) Save(ctx context.Context, m *biz.Media) error {
	return r.data.DB.WithContext(ctx).Create(&MediaModel{
		MediaID:   m.MediaID,
		UserID:    m.UserID,
		MediaType: m.MediaType,
		URL:       m.URL,
		ThumbURL:  m.ThumbURL,
		FileName:  m.FileName,
		MimeType:  m.MimeType,
		Size:      m.Size,
	}).Error
}

func (r *mediaRepo) FindByID(ctx context.Context, mediaID string) (*biz.Media, error) {
	var m MediaModel
	if err := r.data.DB.WithContext(ctx).Where("media_id = ?", mediaID).First(&m).Error; err != nil {
		return nil, err
	}
	return &biz.Media{
		MediaID:   m.MediaID,
		UserID:    m.UserID,
		MediaType: m.MediaType,
		URL:       m.URL,
		ThumbURL:  m.ThumbURL,
		FileName:  m.FileName,
		MimeType:  m.MimeType,
		Size:      m.Size,
	}, nil
}
