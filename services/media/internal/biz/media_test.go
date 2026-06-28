package biz

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"
)

type mockMediaRepo struct {
	media map[string]*Media
}

func newMockMediaRepo() *mockMediaRepo {
	return &mockMediaRepo{media: make(map[string]*Media)}
}

func (m *mockMediaRepo) Save(ctx context.Context, media *Media) error {
	m.media[media.MediaID] = media
	return nil
}
func (m *mockMediaRepo) FindByID(ctx context.Context, mediaID string) (*Media, error) {
	md, ok := m.media[mediaID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return md, nil
}

type mockStorage struct {
	files map[string]string
}

func newMockStorage() *mockStorage {
	return &mockStorage{files: make(map[string]string)}
}

func (s *mockStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	b, _ := io.ReadAll(data)
	s.files[key] = string(b)
	return "https://img.example.com/" + key, nil
}
func (s *mockStorage) PresignedUploadURL(key string, contentType string, expires time.Duration) (string, error) {
	return "https://upload.example.com/" + key, nil
}
func (s *mockStorage) HeadObject(key string) (map[string]string, error) {
	return nil, nil
}
func (s *mockStorage) URL(key string) string {
	return "https://img.example.com/" + key
}

func TestUploadAndSave(t *testing.T) {
	repo := newMockMediaRepo()
	storage := newMockStorage()
	uc := NewMediaUseCase(repo, storage)

	m, err := uc.UploadAndSave(context.Background(), "user-1", 1, "photo.jpg", "image/jpeg", strings.NewReader("fake-image-data"), 1024)
	if err != nil {
		t.Fatalf("UploadAndSave: %v", err)
	}
	if m.MediaID == "" {
		t.Fatal("expected media ID")
	}
	if m.URL == "" {
		t.Fatal("expected URL")
	}
}

func TestGetMedia(t *testing.T) {
	repo := newMockMediaRepo()
	storage := newMockStorage()
	uc := NewMediaUseCase(repo, storage)

	m, _ := uc.UploadAndSave(context.Background(), "user-1", 1, "doc.pdf", "application/pdf", strings.NewReader("data"), 512)

	found, err := uc.GetMedia(context.Background(), m.MediaID)
	if err != nil {
		t.Fatalf("GetMedia: %v", err)
	}
	if found.MediaID != m.MediaID {
		t.Fatal("media ID mismatch")
	}
}

func TestGetMediaNotFound(t *testing.T) {
	repo := newMockMediaRepo()
	storage := newMockStorage()
	uc := NewMediaUseCase(repo, storage)

	_, err := uc.GetMedia(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent media")
	}
}

func TestGenerateUploadToken(t *testing.T) {
	repo := newMockMediaRepo()
	storage := newMockStorage()
	uc := NewMediaUseCase(repo, storage)

	mediaID, uploadURL, publicURL, err := uc.GenerateUploadToken(context.Background(), "user-1", "video.mp4", "video/mp4")
	if err != nil {
		t.Fatalf("GenerateUploadToken: %v", err)
	}
	if mediaID == "" || uploadURL == "" || publicURL == "" {
		t.Fatal("expected non-empty return values")
	}
}
