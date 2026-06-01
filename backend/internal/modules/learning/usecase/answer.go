package usecase

import (
	"context"
	"encoding/json"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	learning_api "github.com/SIniutin/history-app-backend/internal/modules/learning/api"
	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func (s *Service) SubmitAnswer(ctx context.Context, input learning_api.SubmitAnswerInput) (learning_api.SubmitAnswerResult, error) {
	session, err := s.getOwnedSession(ctx, learning_api.SessionInput{
		UserID:    input.UserID,
		SessionID: input.SessionID,
	})
	if err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	if session.Status == domain.LessonSessionStatusFinished {
		return learning_api.SubmitAnswerResult{}, mapDomainError(domain.ErrSessionFinished)
	}
	malformedAnswer := !json.Valid(input.UserAnswer)
	if malformedAnswer {
		input.UserAnswer = json.RawMessage("null")
	}

	current, err := s.queue.GetCurrentPending(ctx, session.ID)
	if err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	if current == nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(domain.ErrNoCurrentChallenge)
	}

	challenge, err := s.content.GetChallenge(ctx, content_domain.ChallengeID(current.ChallengeID))
	if err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	checked := checkAnswer(challenge, input.UserAnswer)
	if malformedAnswer {
		checked = incorrect("answer is malformed")
	}
	attempt := domain.ChallengeAttempt{
		ID:                 domain.ChallengeAttemptID(uuid.New()),
		SessionID:          session.ID,
		SessionChallengeID: current.ID,
		ChallengeID:        current.ChallengeID,
		UserAnswer:         input.UserAnswer,
		IsCorrect:          checked.isCorrect,
		Mistakes:           checked.mistakes,
		AnsweredAt:         time.Now().UTC(),
	}
	if attempt.Mistakes == nil {
		attempt.Mistakes = []string{}
	}
	if err := s.attempts.CreateAttempt(ctx, &attempt); err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	if err := s.queue.MarkAnswered(ctx, current.ID); err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	if !attempt.IsCorrect {
		if err := s.queue.Append(ctx, domain.LessonSessionChallenge{
			ID:          domain.LessonSessionChallengeID(uuid.New()),
			SessionID:   session.ID,
			ChallengeID: current.ChallengeID,
			Status:      domain.SessionChallengeStatusPending,
		}); err != nil {
			return learning_api.SubmitAnswerResult{}, mapDomainError(err)
		}
	}

	next, err := s.queue.GetCurrentPending(ctx, session.ID)
	if err != nil {
		return learning_api.SubmitAnswerResult{}, mapDomainError(err)
	}
	return learning_api.SubmitAnswerResult{
		AttemptID: uuid.UUID(attempt.ID).String(),
		IsCorrect: attempt.IsCorrect,
		Mistakes:  attempt.Mistakes,
		HasNext:   next != nil,
	}, nil
}
