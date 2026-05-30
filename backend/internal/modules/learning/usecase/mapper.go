package usecase

import (
	"errors"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	learning_api "github.com/SIniutin/history-app-backend/internal/modules/learning/api"
	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func mapDomainError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrInvalidInput):
		return errors.Join(learning_api.ErrInvalidInput, err)
	case errors.Is(err, domain.ErrNoChallenges):
		return errors.Join(learning_api.ErrNoChallenges, err)
	case errors.Is(err, domain.ErrNoCurrentChallenge):
		return errors.Join(learning_api.ErrNoCurrentChallenge, err)
	case errors.Is(err, domain.ErrSessionFinished):
		return errors.Join(learning_api.ErrSessionFinished, err)
	case errors.Is(err, domain.ErrForbidden):
		return learning_api.ErrForbidden
	case errors.Is(err, domain.ErrNotFound), errors.Is(err, content_domain.ErrNotFound):
		return learning_api.ErrNotFound
	case errors.Is(err, content_domain.ErrInvalidInput):
		return errors.Join(learning_api.ErrInvalidInput, err)
	default:
		return err
	}
}

func toAPISession(s domain.LessonSession) learning_api.LessonSession {
	var finishedAt *string
	if s.FinishedAt != nil {
		formatted := formatTime(*s.FinishedAt)
		finishedAt = &formatted
	}
	return learning_api.LessonSession{
		ID:         uuid.UUID(s.ID).String(),
		UserID:     uuid.UUID(s.UserID).String(),
		SkillID:    uuid.UUID(s.SkillID).String(),
		Status:     string(s.Status),
		StartedAt:  formatTime(s.StartedAt),
		FinishedAt: finishedAt,
	}
}

func toAPIChallenge(c content_domain.Challenge) learning_api.Challenge {
	return learning_api.Challenge{
		ID:          uuid.UUID(c.ID).String(),
		SkillID:     uuid.UUID(c.SkillID).String(),
		Type:        string(c.Type),
		Difficulty:  string(c.Difficulty),
		Tags:        c.Tags,
		Level:       c.Level,
		LessonCount: c.LessonCount,
		Prompt:      c.Prompt,
		Body:        c.Body,
		Payload:     c.Payload,
		Options:     c.Options,
		Explanation: c.Explanation,
		Position:    c.Position,
		Status:      string(c.Status),
	}
}

func toAPIResult(r domain.LessonSessionResult) learning_api.LessonSessionResult {
	return learning_api.LessonSessionResult{
		SessionID: uuid.UUID(r.SessionID).String(),
		Total:     r.Total,
		Correct:   r.Correct,
		Percent:   r.Percent,
	}
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
