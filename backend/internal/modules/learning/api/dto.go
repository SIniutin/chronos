package api

import (
	"context"
	"encoding/json"
	"errors"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrNotFound           = errors.New("learning item not found")
	ErrForbidden          = errors.New("forbidden")
	ErrSessionFinished    = errors.New("lesson session is finished")
	ErrNoChallenges       = errors.New("no challenges available")
	ErrNoCurrentChallenge = errors.New("no current challenge")
)

type StartSessionInput struct {
	UserID  string `json:"-"`
	SkillID string `json:"skill_id"`
	Limit   int    `json:"limit"`
}

type SessionInput struct {
	UserID    string
	SessionID string
}

type SubmitAnswerInput struct {
	UserID     string          `json:"-"`
	SessionID  string          `json:"-"`
	UserAnswer json.RawMessage `json:"user_answer"`
}

type LessonSession struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	SkillID    string  `json:"skill_id"`
	Status     string  `json:"status"`
	StartedAt  string  `json:"started_at"`
	FinishedAt *string `json:"finished_at,omitempty"`
}

type Challenge struct {
	ID          string          `json:"id"`
	SkillID     string          `json:"skill_id"`
	Type        string          `json:"type"`
	Difficulty  string          `json:"difficulty"`
	Tags        json.RawMessage `json:"tags"`
	Level       int             `json:"level"`
	LessonCount int             `json:"lesson_count"`
	Prompt      string          `json:"prompt"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Options     json.RawMessage `json:"options"`
	Explanation string          `json:"explanation"`
	Position    int             `json:"position"`
	Status      string          `json:"status"`
}

type CurrentChallenge struct {
	SessionChallengeID string    `json:"session_challenge_id"`
	Position           int       `json:"position"`
	Challenge          Challenge `json:"challenge"`
}

type SubmitAnswerResult struct {
	AttemptID string   `json:"attempt_id"`
	IsCorrect bool     `json:"is_correct"`
	Mistakes  []string `json:"mistakes"`
	HasNext   bool     `json:"has_next"`
}

type LessonSessionResult struct {
	SessionID string `json:"session_id"`
	Total     int    `json:"total"`
	Correct   int    `json:"correct"`
	Percent   int    `json:"percent"`
}

type SessionsUsecase interface {
	StartSession(ctx context.Context, input StartSessionInput) (LessonSession, error)
	GetCurrentChallenge(ctx context.Context, input SessionInput) (CurrentChallenge, error)
	SubmitAnswer(ctx context.Context, input SubmitAnswerInput) (SubmitAnswerResult, error)
	FinishSession(ctx context.Context, input SessionInput) (LessonSessionResult, error)
}
