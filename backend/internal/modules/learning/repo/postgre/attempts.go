package postgre

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) CreateAttempt(ctx context.Context, attempt *domain.ChallengeAttempt) error {
	const query = `
		INSERT INTO challenge_attempts (id, session_id, session_challenge_id, challenge_id, user_answer, is_correct, mistakes, answered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(attempt.ID).String(), uuid.UUID(attempt.SessionID).String(), uuid.UUID(attempt.SessionChallengeID).String(), uuid.UUID(attempt.ChallengeID).String(), attempt.UserAnswer, attempt.IsCorrect, attempt.Mistakes, attempt.AnsweredAt)
	return err
}

func (r *repoImpl) ListAttemptsBySession(ctx context.Context, sessionID domain.LessonSessionID) ([]domain.ChallengeAttempt, error) {
	const query = `
		SELECT id, session_id, session_challenge_id, challenge_id, user_answer, is_correct, mistakes, answered_at
		FROM challenge_attempts
		WHERE session_id = $1
		ORDER BY answered_at
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(sessionID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attempts []domain.ChallengeAttempt
	for rows.Next() {
		attempt, err := scanAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}
	return attempts, rows.Err()
}

func scanAttempt(row scanner) (domain.ChallengeAttempt, error) {
	var attempt domain.ChallengeAttempt
	var idRaw, sessionIDRaw, sessionChallengeIDRaw, challengeIDRaw string
	if err := row.Scan(&idRaw, &sessionIDRaw, &sessionChallengeIDRaw, &challengeIDRaw, &attempt.UserAnswer, &attempt.IsCorrect, &attempt.Mistakes, &attempt.AnsweredAt); err != nil {
		return domain.ChallengeAttempt{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return domain.ChallengeAttempt{}, err
	}
	sessionID, err := uuid.Parse(sessionIDRaw)
	if err != nil {
		return domain.ChallengeAttempt{}, err
	}
	sessionChallengeID, err := uuid.Parse(sessionChallengeIDRaw)
	if err != nil {
		return domain.ChallengeAttempt{}, err
	}
	challengeID, err := uuid.Parse(challengeIDRaw)
	if err != nil {
		return domain.ChallengeAttempt{}, err
	}
	attempt.ID = domain.ChallengeAttemptID(id)
	attempt.SessionID = domain.LessonSessionID(sessionID)
	attempt.SessionChallengeID = domain.LessonSessionChallengeID(sessionChallengeID)
	attempt.ChallengeID = domain.ChallengeID(challengeID)
	return attempt, nil
}
