package infra

import (
	"encoding/json"
	"net/http"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/api"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service api.CatalogService
	auth    *users_infra.AuthMiddleware
}

type Dependencies struct {
	Service api.CatalogService
	Auth    *users_infra.AuthMiddleware
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{service: deps.Service, auth: deps.Auth}
}

func (h *Handler) GetCatalog(w http.ResponseWriter, r *http.Request) {
	userID, ok := users_infra.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return
	}
	courseID := r.URL.Query().Get("course_id")
	if courseID == "" {
		writeError(w, http.StatusBadRequest, "course_id is required")
		return
	}
	progress, err := h.service.GetCatalogProgress(r.Context(), userID, courseID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func (h *Handler) CompleteAllForUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}
	progress, err := h.service.CompleteAllForUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, progress)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
