package infra

import (
	"encoding/json"
	"errors"
	"net/http"

	learning_api "github.com/SIniutin/history-app-backend/internal/modules/learning/api"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	sessions learning_api.SessionsUsecase
	auth     *users_infra.AuthMiddleware
}

type Dependencies struct {
	Sessions learning_api.SessionsUsecase
	Auth     *users_infra.AuthMiddleware
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		sessions: deps.Sessions,
		auth:     deps.Auth,
	}
}

func (h *Handler) StartSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req learning_api.StartSessionInput
	if !decodeJSON(w, r, &req) {
		return
	}
	req.UserID = userID
	session, err := h.sessions.StartSession(r.Context(), req)
	writeResult(w, http.StatusCreated, session, err)
}

func (h *Handler) GetCurrentChallenge(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	current, err := h.sessions.GetCurrentChallenge(r.Context(), learning_api.SessionInput{
		UserID:    userID,
		SessionID: chi.URLParam(r, "session_id"),
	})
	writeResult(w, http.StatusOK, current, err)
}

func (h *Handler) SubmitAnswer(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	var req learning_api.SubmitAnswerInput
	if !decodeJSON(w, r, &req) {
		return
	}
	req.UserID = userID
	req.SessionID = chi.URLParam(r, "session_id")
	result, err := h.sessions.SubmitAnswer(r.Context(), req)
	writeResult(w, http.StatusOK, result, err)
}

func (h *Handler) FinishSession(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromRequest(w, r)
	if !ok {
		return
	}
	result, err := h.sessions.FinishSession(r.Context(), learning_api.SessionInput{
		UserID:    userID,
		SessionID: chi.URLParam(r, "session_id"),
	})
	writeResult(w, http.StatusOK, result, err)
}

func userIDFromRequest(w http.ResponseWriter, r *http.Request) (string, bool) {
	userID, ok := users_infra.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return "", false
	}
	return userID, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return false
	}
	return true
}

func writeResult(w http.ResponseWriter, status int, v any, err error) {
	if err != nil {
		writeUsecaseError(w, err)
		return
	}
	writeJSON(w, status, v)
}

func writeUsecaseError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, learning_api.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, learning_api.ErrNoChallenges):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, learning_api.ErrNoCurrentChallenge):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, learning_api.ErrSessionFinished):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, learning_api.ErrForbidden):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, learning_api.ErrNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
