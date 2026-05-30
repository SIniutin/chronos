package domain

import (
	"context"
)

type SessionRepository interface {
	CreateSession(ctx context.Context, session *LessonSession) error
	GetSession(ctx context.Context, id LessonSessionID) (*LessonSession, error)
	UpdateSession(ctx context.Context, session *LessonSession) error
}

type SessionChallengeRepository interface {
	CreateMany(ctx context.Context, challenges []LessonSessionChallenge) error
	GetCurrentPending(ctx context.Context, sessionID LessonSessionID) (*LessonSessionChallenge, error)
	MarkAnswered(ctx context.Context, id LessonSessionChallengeID) error
	ListBySession(ctx context.Context, sessionID LessonSessionID) ([]LessonSessionChallenge, error)
}

type AttemptRepository interface {
	CreateAttempt(ctx context.Context, attempt *ChallengeAttempt) error
	ListAttemptsBySession(ctx context.Context, sessionID LessonSessionID) ([]ChallengeAttempt, error)
}
