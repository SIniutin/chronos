package users_infra

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	registerCase users_api.RegisterUsecase
	loginCase    users_api.LoginUsecase
	refreshCase  users_api.RefreshUsecase
	getMeCase    users_api.GetMeUsecase
	logoutCase   users_api.LogoutUsecase
	changeRole   users_api.ChangeUserRoleUsecase
	findUser     users_api.FindUserUsecase
	auth         *AuthMiddleware
}

type Dependencies struct {
	Register   users_api.RegisterUsecase
	Login      users_api.LoginUsecase
	Refresh    users_api.RefreshUsecase
	GetMe      users_api.GetMeUsecase
	Logout     users_api.LogoutUsecase
	ChangeRole users_api.ChangeUserRoleUsecase
	FindUser   users_api.FindUserUsecase
	Auth       *AuthMiddleware
}

func NewHandler(deps Dependencies) *Handler {
	return &Handler{
		registerCase: deps.Register,
		loginCase:    deps.Login,
		refreshCase:  deps.Refresh,
		getMeCase:    deps.GetMe,
		logoutCase:   deps.Logout,
		changeRole:   deps.ChangeRole,
		findUser:     deps.FindUser,
		auth:         deps.Auth,
	}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req users_api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, err := h.registerCase.Exec(r.Context(), users_api.RegisterInput{
		Email:    req.Email,
		Login:    req.Login,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrConflictEmail):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, users_api.ErrConflictLogin):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, users_api.UserResponse{
		Email:     user.Email,
		Login:     user.Login,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.DateTime),
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if h.loginCase == nil {
		writeError(w, http.StatusNotImplemented, "login usecase is not configured")
		return
	}

	var req users_api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokenPair, err := h.loginCase.Exec(r.Context(), users_api.LoginInput{
		Identity: req.Identity,
		Password: req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidCredential):
			writeError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, users_api.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if h.logoutCase == nil {
		writeError(w, http.StatusNotImplemented, "logout usecase is not configured")
		return
	}

	idRaw, ok := UserIDFromContext(r.Context())
	if !ok || idRaw == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return
	}

	var req users_api.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	if err := h.logoutCase.Exec(r.Context(), users_api.LogoutInput{
		UserID:       idRaw,
		RefreshToken: req.RefreshToken,
	}); err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrInvalidCredential):
			writeError(w, http.StatusUnauthorized, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	if h.refreshCase == nil {
		writeError(w, http.StatusNotImplemented, "refresh usecase is not configured")
		return
	}

	var req users_api.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	tokenPair, err := h.refreshCase.Exec(r.Context(), users_api.RefreshInput{RefreshToken: req.RefreshToken})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrInvalidCredential), errors.Is(err, users_api.ErrSessionExpired), errors.Is(err, users_api.ErrSessionRevoked):
			writeError(w, http.StatusUnauthorized, users_api.ErrInvalidCredential.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, users_api.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	if h.getMeCase == nil {
		writeError(w, http.StatusNotImplemented, "get me usecase is not configured")
		return
	}

	idRaw, ok := UserIDFromContext(r.Context())
	if !ok || idRaw == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return
	}

	user, err := h.getMeCase.Exec(r.Context(), users_api.GetMeInput{UserID: idRaw})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, users_api.UserResponse{
		Email:     user.Email,
		Login:     user.Login,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) ChangeUserRole(w http.ResponseWriter, r *http.Request) {
	if h.changeRole == nil {
		writeError(w, http.StatusNotImplemented, "change user role usecase is not configured")
		return
	}

	var req users_api.ChangeUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}

	user, err := h.changeRole.Exec(r.Context(), users_api.ChangeUserRoleInput{
		UserID: chi.URLParam(r, "user_id"),
		Role:   req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, users_api.UserResponse{
		Email:     user.Email,
		Login:     user.Login,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) FindUser(w http.ResponseWriter, r *http.Request) {
	if h.findUser == nil {
		writeError(w, http.StatusNotImplemented, "find user usecase is not configured")
		return
	}

	user, err := h.findUser.Exec(r.Context(), users_api.FindUserInput{
		Identity: r.URL.Query().Get("identity"),
	})
	if err != nil {
		switch {
		case errors.Is(err, users_api.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, users_api.ErrUserNotFound):
			writeError(w, http.StatusNotFound, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, users_api.AdminUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Login:     user.Login,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

func (h *Handler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "update me is not implemented yet")
}

func (h *Handler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	writeError(w, http.StatusNotImplemented, "delete me is not implemented yet")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, `{"error":"failed to encode response"}`, http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}
