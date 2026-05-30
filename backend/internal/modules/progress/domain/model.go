package domain

import (
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type UserID = users_domain.UserID
type CourseID = content_domain.CourseID
type UnitID = content_domain.UnitID
type SkillID = content_domain.SkillID

type ProgressStatus string

const (
	ProgressStatusLocked     ProgressStatus = "locked"
	ProgressStatusAvailable  ProgressStatus = "available"
	ProgressStatusInProgress ProgressStatus = "in_progress"
	ProgressStatusCompleted  ProgressStatus = "completed"
)

type SkillProgress struct {
	UserID  UserID
	SkillID SkillID

	Status  ProgressStatus
	Level   int
	Mastery float64

	CorrectAnswers int
	WrongAnswers   int

	StartedAt   time.Time
	CompletedAt *time.Time
	UpdatedAt   time.Time
}

type UnitProgress struct {
	UserID UserID
	UnitID UnitID

	Status ProgressStatus

	StartedAt   time.Time
	CompletedAt *time.Time
	UpdatedAt   time.Time
}

type CourseProgress struct {
	UserID   UserID
	CourseID CourseID

	Status ProgressStatus

	StartedAt   time.Time
	CompletedAt *time.Time
	UpdatedAt   time.Time
}

type SessionProgressInput struct {
	UserID UserID

	SkillID SkillID

	CorrectAnswers int
	TotalAnswers   int

	CompletedAt time.Time
}

type SessionProgressResult struct {
	SkillCompleted  bool
	UnitCompleted   bool
	CourseCompleted bool

	NewSkillLevel int
	NewMastery    float64

	UnlockedSkillIDs []SkillID
	UnlockedUnitIDs  []UnitID
}
