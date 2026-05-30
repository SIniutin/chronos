package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	GetUserXP(ctx context.Context, userID UserID) (*UserXP, error)
	SaveUserXP(ctx context.Context, xp UserXP) error
	CreateXPTransaction(ctx context.Context, tx XPTransaction) error
	GetUserStreak(ctx context.Context, userID UserID) (*UserStreak, error)
	SaveUserStreak(ctx context.Context, streak UserStreak) error
	ListAchievements(ctx context.Context) ([]Achievement, error)
	ListUserAchievements(ctx context.Context, userID UserID) ([]Achievement, error)
	UnlockAchievement(ctx context.Context, userID UserID, achievementID uuid.UUID, unlockedAt time.Time) (bool, error)
}
