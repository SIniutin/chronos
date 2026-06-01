package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/config"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const maxImageUploadSize = 8 << 20

type Handler struct {
	client *minio.Client
	cfg    config.S3Config
	auth   *users_infra.AuthMiddleware
}

type Dependencies struct {
	Config config.S3Config
	Auth   *users_infra.AuthMiddleware
}

type UploadResponse struct {
	URL         string `json:"url"`
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

func NewHandler(deps Dependencies) (*Handler, error) {
	client, err := minio.New(deps.Config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(deps.Config.AccessKey, deps.Config.SecretKey, ""),
		Secure: deps.Config.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Handler{client: client, cfg: deps.Config, auth: deps.Auth}, nil
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/editor/media", func(r chi.Router) {
		if h.auth != nil {
			r.Use(h.auth.Auth)
			r.Use(users_infra.RequireAnyRole("content_editor", "admin"))
		}
		r.Post("/images", h.UploadImage)
	})
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxImageUploadSize)
	if err := r.ParseMultipartForm(maxImageUploadSize); err != nil {
		writeError(w, http.StatusBadRequest, "image is too large or multipart body is invalid")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file field is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !allowedImageType(contentType) {
		writeError(w, http.StatusBadRequest, "only png, jpeg and webp images are allowed")
		return
	}
	if header.Size <= 0 || header.Size > maxImageUploadSize {
		writeError(w, http.StatusBadRequest, "image is too large")
		return
	}

	key := objectKey(header.Filename, contentType)
	if err := h.EnsureBucket(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to prepare media bucket")
		return
	}
	if _, err := h.client.PutObject(r.Context(), h.cfg.Bucket, key, file, header.Size, minio.PutObjectOptions{ContentType: contentType}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	writeJSON(w, http.StatusCreated, UploadResponse{
		URL:         h.publicURL(key),
		Key:         key,
		ContentType: contentType,
		Size:        header.Size,
	})
}

func (h *Handler) EnsureBucket(ctx context.Context) error {
	exists, err := h.client.BucketExists(ctx, h.cfg.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return h.client.MakeBucket(ctx, h.cfg.Bucket, minio.MakeBucketOptions{})
}

func (h *Handler) publicURL(key string) string {
	return strings.TrimRight(h.cfg.PublicBaseURL, "/") + "/" + strings.TrimLeft(key, "/")
}

func allowedImageType(contentType string) bool {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/png", "image/jpeg", "image/webp":
		return true
	default:
		return false
	}
}

func objectKey(filename, contentType string) string {
	ext := strings.ToLower(path.Ext(filename))
	if ext == "" {
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/webp":
			ext = ".webp"
		default:
			ext = ".jpg"
		}
	}
	now := time.Now().UTC()
	return fmt.Sprintf("images/%04d/%02d/%s%s", now.Year(), now.Month(), uuid.NewString(), ext)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
