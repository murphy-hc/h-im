package biz

import (
	"context"
	"io"
	"time"
)

// Media is the domain entity for an uploaded file.
type Media struct {
	MediaID   string
	UserID    string
	MediaType int32
	URL       string
	ThumbURL  string
	FileName  string
	MimeType  string
	Size      int64
}

// MediaRepo defines the media repository interface.
type MediaRepo interface {
	Save(ctx context.Context, m *Media) error
	FindByID(ctx context.Context, mediaID string) (*Media, error)
}

// Storage abstracts the object storage backend.
type Storage interface {
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)
	PresignedUploadURL(key string, contentType string, expires time.Duration) (string, error)
	HeadObject(key string) (metadata map[string]string, err error)
	URL(key string) string
}
