package data

import (
	"context"
	"io"
	"time"

	"github.com/murphy-hc/h-im/pkg/oss"
	"github.com/murphy-hc/h-im/services/media/internal/biz"
)

var _ biz.Storage = (*ossStorage)(nil)

type ossStorage struct {
	client *oss.Client
}

func NewOSSStorage(client *oss.Client) *ossStorage {
	return &ossStorage{client: client}
}

func (s *ossStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	return s.client.Upload(key, data, contentType)
}

func (s *ossStorage) PresignedUploadURL(key, contentType string, expires time.Duration) (string, error) {
	return s.client.PresignedUploadURL(key, contentType, expires)
}

func (s *ossStorage) HeadObject(key string) (map[string]string, error) {
	headers, err := s.client.HeadObject(key)
	if err != nil {
		return nil, err
	}
	m := make(map[string]string)
	for k := range headers {
		m[k] = headers.Get(k)
	}
	return m, nil
}

func (s *ossStorage) URL(key string) string {
	return s.client.URL(key)
}
