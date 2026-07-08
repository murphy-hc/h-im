package server

import (
	"fmt"
	"net/http"

	pb "github.com/murphy-hc/h-im/gen/go/him/media/v1"
	"github.com/murphy-hc/h-im/services/media/internal/biz"
	"github.com/murphy-hc/h-im/services/media/internal/conf"
)

// MediaHTTPHandler handles HTTP upload/callback requests.
type MediaHTTPHandler struct {
	uc     *biz.MediaUseCase
	secret string
}

// NewMediaHTTPHandler creates a MediaHTTPHandler from config.
func NewMediaHTTPHandler(uc *biz.MediaUseCase, bc *conf.Bootstrap) *MediaHTTPHandler {
	return &MediaHTTPHandler{uc: uc, secret: bc.GetMediaSecret()}
}

func (h *MediaHTTPHandler) auth(r *http.Request) bool {
	if h.secret == "" {
		return false // fail closed: no secret configured
	}
	token := r.Header.Get("X-Media-Secret")
	if token == "" {
		token = r.FormValue("secret")
	}
	return token == h.secret
}

// Upload handles POST /media/v1/upload.
func (h *MediaHTTPHandler) Upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, fmt.Sprintf("parse form: %v", err), http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("read file: %v", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	m, err := h.uc.UploadAndSave(r.Context(), userID, mediaTypeInt(r.FormValue("media_type")), header.Filename, mimeType, file, header.Size)
	if err != nil {
		http.Error(w, fmt.Sprintf("upload: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"media_id":"%s","url":"%s"}`, m.MediaID, m.URL)
}

// Token handles POST /media/v1/token.
func (h *MediaHTTPHandler) Token(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	userID := r.FormValue("user_id")
	if userID == "" {
		http.Error(w, "missing user_id", http.StatusBadRequest)
		return
	}

	mediaID, uploadURL, publicURL, err := h.uc.GenerateUploadToken(r.Context(), userID, r.FormValue("file_name"), r.FormValue("mime_type"))
	if err != nil {
		http.Error(w, fmt.Sprintf("token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"media_id":"%s","upload_url":"%s","public_url":"%s"}`,
		mediaID, uploadURL, publicURL)
}

// Confirm handles POST /media/v1/confirm.
func (h *MediaHTTPHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !h.auth(r) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	mediaID := r.FormValue("media_id")
	userID := r.FormValue("user_id")
	key := r.FormValue("key")
	fileName := r.FormValue("file_name")
	mimeType := r.FormValue("mime_type")
	if mediaID == "" || userID == "" || key == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	m, err := h.uc.ConfirmUpload(r.Context(), mediaID, userID, key, fileName, mimeType, mediaTypeInt(r.FormValue("media_type")))
	if err != nil {
		http.Error(w, fmt.Sprintf("confirm: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"media_id":"%s","url":"%s"}`, m.MediaID, m.URL)
}

func mediaTypeInt(s string) int32 {
	switch s {
	case "IMAGE":
		return int32(pb.MediaType_MEDIA_TYPE_IMAGE)
	case "VOICE":
		return int32(pb.MediaType_MEDIA_TYPE_VOICE)
	case "VIDEO":
		return int32(pb.MediaType_MEDIA_TYPE_VIDEO)
	case "FILE":
		return int32(pb.MediaType_MEDIA_TYPE_FILE)
	default:
		return 0
	}
}
