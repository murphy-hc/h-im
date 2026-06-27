package biz

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/rs/xid"
)

// MediaUseCase handles media business logic.
type MediaUseCase struct {
	repo    MediaRepo
	storage Storage
}

// NewMediaUseCase creates a MediaUseCase.
func NewMediaUseCase(repo MediaRepo, storage Storage) *MediaUseCase {
	return &MediaUseCase{repo: repo, storage: storage}
}

// Save persists a media record.
func (uc *MediaUseCase) Save(ctx context.Context, m *Media) error {
	return uc.repo.Save(ctx, m)
}

// GetMedia retrieves media by ID.
func (uc *MediaUseCase) GetMedia(ctx context.Context, mediaID string) (*Media, error) {
	return uc.repo.FindByID(ctx, mediaID)
}

// UploadAndSave uploads data to storage and persists the media record.
func (uc *MediaUseCase) UploadAndSave(ctx context.Context, userID string, mediaType int32, fileName, mimeType string, data io.Reader, size int64) (*Media, error) {
	key := objectKey(userID, fileName)
	url, err := uc.storage.Upload(ctx, key, data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("upload: %w", err)
	}
	m := &Media{
		MediaID:   xid.New().String(),
		UserID:    userID,
		MediaType: mediaType,
		URL:       url,
		FileName:  fileName,
		MimeType:  mimeType,
		Size:      size,
	}
	if err := uc.repo.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return m, nil
}

// GenerateUploadToken creates a media ID and returns a pre-signed upload URL.
func (uc *MediaUseCase) GenerateUploadToken(ctx context.Context, userID, fileName, mimeType string) (mediaID, uploadURL, publicURL string, err error) {
	mediaID = xid.New().String()
	key := objectKey(userID, fileName)
	uploadURL, err = uc.storage.PresignedUploadURL(key, mimeType, 10*time.Minute)
	if err != nil {
		return "", "", "", fmt.Errorf("token: %w", err)
	}
	publicURL = uc.storage.URL(key)
	return mediaID, uploadURL, publicURL, nil
}

// ConfirmUpload verifies the file exists in storage and persists the record.
func (uc *MediaUseCase) ConfirmUpload(ctx context.Context, mediaID, userID, key, fileName, mimeType string, mediaType int32) (*Media, error) {
	metadata, err := uc.storage.HeadObject(key)
	if err != nil {
		return nil, fmt.Errorf("object not found: %w", err)
	}
	size := int64(0)
	if l := metadata["Content-Length"]; l != "" {
		fmt.Sscanf(l, "%d", &size)
	}
	m := &Media{
		MediaID:   mediaID,
		UserID:    userID,
		MediaType: mediaType,
		URL:       uc.storage.URL(key),
		FileName:  fileName,
		MimeType:  mimeType,
		Size:      size,
	}
	if err := uc.repo.Save(ctx, m); err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}
	return m, nil
}

func objectKey(userID, fileName string) string {
	ext := filepath.Ext(fileName)
	return fmt.Sprintf("media/%s/%s%s", userID, xid.New().String(), ext)
}
