package domain

import (
	"encoding/json"
	"time"

	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type UserID = users_domain.UserID

type XPReason string

const (
	XPReasonSessionCompleted XPReason = "session_completed"
	XPReasonCorrectAnswer    XPReason = "correct_answer"
	XPReasonPerfectSession   XPReason = "perfect_session"
	XPReasonAchievement      XPReason = "achievement"
	XPReasonDailyGoal        XPReason = "daily_goal"
)

type UserXP struct {
	UserID    UserID
	TotalXP   int
	Level     int
	UpdatedAt time.Time
}

type XPTransaction struct {
	ID         uuid.UUID
	UserID     UserID
	Amount     int
	Reason     XPReason
	SourceType string
	SourceID   uuid.UUID
	CreatedAt  time.Time
}

type UserStreak struct {
	UserID           UserID
	CurrentDays      int
	LongestDays      int
	LastActivityDate time.Time
	UpdatedAt        time.Time
}

type Achievement struct {
	ID          uuid.UUID
	Code        string
	Title       string
	Description string
	XPReward    int
	Condition   json.RawMessage
	CreatedAt   time.Time
}

type UserAchievement struct {
	UserID        UserID
	AchievementID uuid.UUID
	UnlockedAt    time.Time
}

type SessionRewardInput struct {
	UserID         UserID
	SessionID      uuid.UUID
	CorrectAnswers int
	TotalAnswers   int
	CompletedAt    time.Time
}

type SessionRewardResult struct {
	XPGained             int
	TotalXP              int
	NewLevel             int
	StreakUpdated        bool
	CurrentStreak        int
	UnlockedAchievements []Achievement
}

type GamificationProfile struct {
	UserXP       UserXP
	Streak       *UserStreak
	Achievements []Achievement
}
