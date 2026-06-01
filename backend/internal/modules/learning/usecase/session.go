package usecase

import (
	"context"
	crand "crypto/rand"
	"errors"
	"math/big"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	learning_api "github.com/SIniutin/history-app-backend/internal/modules/learning/api"
	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

const defaultSessionLimit = 10

func (s *Service) StartSession(ctx context.Context, input learning_api.StartSessionInput) (learning_api.LessonSession, error) {
	userID, skillID, limit, err := parseStartInput(input)
	if err != nil {
		return learning_api.LessonSession{}, mapDomainError(err)
	}

	challenges, err := s.pickSessionChallenges(ctx, userID, skillID, limit)
	if err != nil {
		return learning_api.LessonSession{}, mapDomainError(err)
	}
	if len(challenges) == 0 {
		return learning_api.LessonSession{}, mapDomainError(domain.ErrNoChallenges)
	}

	now := time.Now().UTC()
	session := domain.LessonSession{
		ID:        domain.LessonSessionID(uuid.New()),
		UserID:    userID,
		SkillID:   domain.SkillID(skillID),
		Status:    domain.LessonSessionStatusActive,
		StartedAt: now,
	}
	if err := s.sessions.CreateSession(ctx, &session); err != nil {
		return learning_api.LessonSession{}, mapDomainError(err)
	}

	queue := make([]domain.LessonSessionChallenge, 0, len(challenges))
	for i, challengeID := range challenges {
		queue = append(queue, domain.LessonSessionChallenge{
			ID:          domain.LessonSessionChallengeID(uuid.New()),
			SessionID:   session.ID,
			ChallengeID: challengeID,
			Position:    i + 1,
			Status:      domain.SessionChallengeStatusPending,
		})
	}
	if err := s.queue.CreateMany(ctx, queue); err != nil {
		return learning_api.LessonSession{}, mapDomainError(err)
	}

	return toAPISession(session), nil
}

func (s *Service) GetCurrentChallenge(ctx context.Context, input learning_api.SessionInput) (learning_api.CurrentChallenge, error) {
	session, err := s.getOwnedSession(ctx, input)
	if err != nil {
		return learning_api.CurrentChallenge{}, mapDomainError(err)
	}
	if session.Status == domain.LessonSessionStatusFinished {
		return learning_api.CurrentChallenge{}, mapDomainError(domain.ErrSessionFinished)
	}

	current, err := s.queue.GetCurrentPending(ctx, session.ID)
	if err != nil {
		return learning_api.CurrentChallenge{}, mapDomainError(err)
	}
	if current == nil {
		return learning_api.CurrentChallenge{}, mapDomainError(domain.ErrNoCurrentChallenge)
	}

	challenge, err := s.content.GetChallenge(ctx, content_domain.ChallengeID(current.ChallengeID))
	if err != nil {
		return learning_api.CurrentChallenge{}, mapDomainError(err)
	}
	return learning_api.CurrentChallenge{
		SessionChallengeID: uuid.UUID(current.ID).String(),
		Position:           current.Position,
		Challenge:          toAPIChallenge(challenge),
	}, nil
}

func (s *Service) FinishSession(ctx context.Context, input learning_api.SessionInput) (learning_api.LessonSessionResult, error) {
	session, err := s.getOwnedSession(ctx, input)
	if err != nil {
		return learning_api.LessonSessionResult{}, mapDomainError(err)
	}
	if session.Status != domain.LessonSessionStatusFinished {
		pending, err := s.queue.GetCurrentPending(ctx, session.ID)
		if err != nil {
			return learning_api.LessonSessionResult{}, mapDomainError(err)
		}
		if pending != nil {
			return learning_api.LessonSessionResult{}, mapDomainError(domain.ErrInvalidInput)
		}
	}
	result, err := s.calculateResult(ctx, session.ID)
	if err != nil {
		return learning_api.LessonSessionResult{}, mapDomainError(err)
	}

	if session.Status != domain.LessonSessionStatusFinished {
		now := time.Now().UTC()
		session.Status = domain.LessonSessionStatusFinished
		session.FinishedAt = &now
		if err := s.sessions.UpdateSession(ctx, session); err != nil {
			return learning_api.LessonSessionResult{}, mapDomainError(err)
		}
		if err := s.progress.RecordSessionResult(ctx, ProgressInput{
			UserID:         session.UserID,
			SkillID:        session.SkillID,
			CorrectAnswers: result.Correct,
			TotalAnswers:   result.Total,
			CompletedAt:    now,
		}); err != nil {
			return learning_api.LessonSessionResult{}, err
		}
		if err := s.gamification.RewardSessionCompleted(ctx, GamificationInput{
			UserID:         session.UserID,
			SessionID:      session.ID,
			CorrectAnswers: result.Correct,
			TotalAnswers:   result.Total,
			CompletedAt:    now,
		}); err != nil {
			return learning_api.LessonSessionResult{}, err
		}
	}

	return toAPIResult(result), nil
}

func (s *Service) pickSessionChallenges(ctx context.Context, userID domain.UserID, skillID content_domain.SkillID, limit int) ([]domain.ChallengeID, error) {
	if s.picker != nil {
		picked, err := s.picker.PickChallengesForSession(ctx, userID, domain.SkillID(skillID), limit)
		if err != nil {
			return nil, err
		}
		if len(picked) > 0 {
			return picked, nil
		}
	}
	challenges, err := s.content.ListPublishedChallenges(ctx, skillID)
	if err != nil {
		return nil, err
	}
	if len(challenges) > limit {
		shuffleChallenges(challenges)
		challenges = challenges[:limit]
	} else {
		shuffleChallenges(challenges)
	}
	out := make([]domain.ChallengeID, 0, len(challenges))
	for _, challenge := range challenges {
		out = append(out, domain.ChallengeID(challenge.ID))
	}
	return out, nil
}

func (s *Service) calculateResult(ctx context.Context, sessionID domain.LessonSessionID) (domain.LessonSessionResult, error) {
	attempts, err := s.attempts.ListAttemptsBySession(ctx, sessionID)
	if err != nil {
		return domain.LessonSessionResult{}, err
	}
	correct := 0
	for _, attempt := range attempts {
		if attempt.IsCorrect {
			correct++
		}
	}
	total := len(attempts)
	percent := 0
	if total > 0 {
		percent = correct * 100 / total
	}
	return domain.LessonSessionResult{
		SessionID: sessionID,
		Total:     total,
		Correct:   correct,
		Percent:   percent,
	}, nil
}

func (s *Service) getOwnedSession(ctx context.Context, input learning_api.SessionInput) (*domain.LessonSession, error) {
	userID, sessionID, err := parseSessionInput(input)
	if err != nil {
		return nil, err
	}
	session, err := s.sessions.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if session.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return session, nil
}

func parseStartInput(input learning_api.StartSessionInput) (domain.UserID, content_domain.SkillID, int, error) {
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return domain.UserID{}, content_domain.SkillID{}, 0, errors.Join(domain.ErrInvalidInput, err)
	}
	skillUUID, err := uuid.Parse(input.SkillID)
	if err != nil {
		return domain.UserID{}, content_domain.SkillID{}, 0, errors.Join(domain.ErrInvalidInput, err)
	}
	limit := input.Limit
	if limit <= 0 {
		limit = defaultSessionLimit
	}
	return domain.UserID(userUUID), content_domain.SkillID(skillUUID), limit, nil
}

func parseSessionInput(input learning_api.SessionInput) (domain.UserID, domain.LessonSessionID, error) {
	userUUID, err := uuid.Parse(input.UserID)
	if err != nil {
		return domain.UserID{}, domain.LessonSessionID{}, errors.Join(domain.ErrInvalidInput, err)
	}
	sessionUUID, err := uuid.Parse(input.SessionID)
	if err != nil {
		return domain.UserID{}, domain.LessonSessionID{}, errors.Join(domain.ErrInvalidInput, err)
	}
	return domain.UserID(userUUID), domain.LessonSessionID(sessionUUID), nil
}

func shuffleChallenges(challenges []content_domain.Challenge) {
	for i := len(challenges) - 1; i > 0; i-- {
		n, err := crand.Int(crand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			continue
		}
		j := int(n.Int64())
		challenges[i], challenges[j] = challenges[j], challenges[i]
	}
}
