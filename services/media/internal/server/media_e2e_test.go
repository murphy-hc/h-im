package server_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/murphy-hc/h-im/services/media/internal/biz"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
	"github.com/murphy-hc/h-im/services/media/internal/server"
)

type e2eMediaRepo struct {
	media map[string]*biz.Media
}

func newE2EMediaRepo() *e2eMediaRepo {
	return &e2eMediaRepo{media: make(map[string]*biz.Media)}
}

func (r *e2eMediaRepo) Save(_ context.Context, m *biz.Media) error {
	r.media[m.MediaID] = m
	return nil
}
func (r *e2eMediaRepo) FindByID(_ context.Context, mediaID string) (*biz.Media, error) {
	m, ok := r.media[mediaID]
	if !ok {
		return nil, context.DeadlineExceeded
	}
	return m, nil
}

type e2eStorage struct {
	files map[string]string
}

func newE2EStorage() *e2eStorage { return &e2eStorage{files: make(map[string]string)} }
func (s *e2eStorage) Upload(_ context.Context, key string, data io.Reader, _ string) (string, error) {
	b, _ := io.ReadAll(data)
	s.files[key] = string(b)
	return "https://img.example.com/" + key, nil
}
func (s *e2eStorage) PresignedUploadURL(key string, _ string, _ time.Duration) (string, error) {
	return "https://upload.example.com/" + key, nil
}
func (s *e2eStorage) HeadObject(_ string) (map[string]string, error) { return nil, nil }
func (s *e2eStorage) URL(key string) string                         { return "https://img.example.com/" + key }

func newE2EMediaHandler() *server.MediaHTTPHandler {
	repo := newE2EMediaRepo()
	storage := newE2EStorage()
	uc := biz.NewMediaUseCase(repo, storage)
	bc := &conf.Bootstrap{MediaSecret: "test-secret"}
	return server.NewMediaHTTPHandler(uc, bc)
}

func TestE2E_UploadAndToken(t *testing.T) {
	h := newE2EMediaHandler()

	// 1. Upload
	body := bytes.NewBuffer(nil)
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"user_id\"\r\n\r\nuser-1\r\n")
	body.WriteString("--boundary\r\n")
	body.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"test.jpg\"\r\n")
	body.WriteString("Content-Type: image/jpeg\r\n\r\nfake-image-bytes\r\n")
	body.WriteString("--boundary--\r\n")

	req := httptest.NewRequest("POST", "/media/v1/upload", body)
	req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")
	req.Header.Set("X-Media-Secret", "test-secret")
	w := httptest.NewRecorder()

	h.Upload(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("Upload returned %d: %s", w.Code, w.Body.String())
	}

	// 2. Token
	tokenBody := bytes.NewBufferString("user_id=user-1&file_name=video.mp4&mime_type=video/mp4")
	req2 := httptest.NewRequest("POST", "/media/v1/token", tokenBody)
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("X-Media-Secret", "test-secret")
	w2 := httptest.NewRecorder()

	h.Token(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("Token returned %d: %s", w2.Code, w2.Body.String())
	}
}
