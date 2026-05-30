package infra

import (
	"encoding/json"
	"net/http"

	"github.com/SIniutin/history-app-backend/internal/modules/gamification/api"
	"github.com/SIniutin/history-app-backend/internal/modules/gamification/domain"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	"github.com/google/uuid"
)

type Handler struct {
	service api.Service
	auth    *users_infra.AuthMiddleware
}

type Dependencies struct {
	Service api.Service
	Auth    *users_infra.AuthMiddleware
}

func NewHandler(deps Dependencies) *Handler { return &Handler{service: deps.Service, auth: deps.Auth} }

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userIDRaw, ok := users_infra.UserIDFromContext(r.Context())
	if !ok || userIDRaw == "" {
		writeError(w, http.StatusUnauthorized, "user is not authenticated")
		return
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	profile, err := h.service.GetProfile(r.Context(), domain.UserID(userID))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, toAPIProfile(profile))
}

func toAPIProfile(profile *domain.GamificationProfile) api.Profile {
	out := api.Profile{TotalXP: profile.UserXP.TotalXP, Level: profile.UserXP.Level}
	if profile.Streak != nil {
		out.CurrentStreak = profile.Streak.CurrentDays
		out.LongestStreak = profile.Streak.LongestDays
	}
	for _, a := range profile.Achievements {
		out.Achievements = append(out.Achievements, api.Achievement{Code: a.Code, Title: a.Title, Description: a.Description, XPReward: a.XPReward})
	}
	return out
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
