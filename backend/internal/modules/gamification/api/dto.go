package api

import (
	"context"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/gamification/domain"
	"github.com/google/uuid"
)

type Service interface {
	RewardSessionCompleted(ctx context.Context, input domain.SessionRewardInput) (*domain.SessionRewardResult, error)
	AddXP(ctx context.Context, userID domain.UserID, amount int, reason domain.XPReason, sourceType string, sourceID uuid.UUID) error
	UpdateStreak(ctx context.Context, userID domain.UserID, activityDate time.Time) (*domain.UserStreak, error)
	GetProfile(ctx context.Context, userID domain.UserID) (*domain.GamificationProfile, error)
}

type Profile struct {
	TotalXP       int           `json:"total_xp"`
	Level         int           `json:"level"`
	CurrentStreak int           `json:"current_streak"`
	LongestStreak int           `json:"longest_streak"`
	Achievements  []Achievement `json:"achievements"`
}

type Achievement struct {
	Code        string `json:"code"`
	Title       string `json:"title"`
	Description string `json:"description"`
	XPReward    int    `json:"xp_reward"`
}
