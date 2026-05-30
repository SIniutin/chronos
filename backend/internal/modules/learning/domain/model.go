package domain

import (
	"encoding/json"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type UserID = users_domain.UserID
type SkillID = content_domain.SkillID
type ChallengeID = content_domain.ChallengeID

type LessonSessionID uuid.UUID
type LessonSessionChallengeID uuid.UUID
type ChallengeAttemptID uuid.UUID

type LessonSessionStatus string
type SessionChallengeStatus string

const (
	LessonSessionStatusActive   LessonSessionStatus = "active"
	LessonSessionStatusFinished LessonSessionStatus = "finished"

	SessionChallengeStatusPending  SessionChallengeStatus = "pending"
	SessionChallengeStatusAnswered SessionChallengeStatus = "answered"
)

type LessonSession struct {
	ID      LessonSessionID
	UserID  UserID
	SkillID SkillID

	Status     LessonSessionStatus
	StartedAt  time.Time
	FinishedAt *time.Time
}

type LessonSessionChallenge struct {
	ID          LessonSessionChallengeID
	SessionID   LessonSessionID
	ChallengeID ChallengeID

	Position int
	Status   SessionChallengeStatus
}

type ChallengeAttempt struct {
	ID                 ChallengeAttemptID
	SessionID          LessonSessionID
	SessionChallengeID LessonSessionChallengeID
	ChallengeID        ChallengeID
	UserAnswer         json.RawMessage

	IsCorrect bool
	Mistakes  []string

	AnsweredAt time.Time
}

type LessonSessionResult struct {
	SessionID LessonSessionID
	Total     int
	Correct   int
	Percent   int
}
