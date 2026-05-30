package postgre

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) CreateMany(ctx context.Context, challenges []domain.LessonSessionChallenge) error {
	for _, challenge := range challenges {
		const query = `
			INSERT INTO lesson_session_challenges (id, session_id, challenge_id, position, status)
			VALUES ($1, $2, $3, $4, $5)
		`
		if _, err := r.pool.Exec(ctx, query, uuid.UUID(challenge.ID).String(), uuid.UUID(challenge.SessionID).String(), uuid.UUID(challenge.ChallengeID).String(), challenge.Position, challenge.Status); err != nil {
			return err
		}
	}
	return nil
}

func (r *repoImpl) GetCurrentPending(ctx context.Context, sessionID domain.LessonSessionID) (*domain.LessonSessionChallenge, error) {
	const query = `
		SELECT id, session_id, challenge_id, position, status
		FROM lesson_session_challenges
		WHERE session_id = $1 AND status = $2
		ORDER BY position
		LIMIT 1
	`
	challenge, err := scanSessionChallenge(r.pool.QueryRow(ctx, query, uuid.UUID(sessionID).String(), domain.SessionChallengeStatusPending))
	if err != nil {
		if err == domain.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &challenge, nil
}

func (r *repoImpl) MarkAnswered(ctx context.Context, id domain.LessonSessionChallengeID) error {
	const query = `
		UPDATE lesson_session_challenges
		SET status = $2
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, uuid.UUID(id).String(), domain.SessionChallengeStatusAnswered)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *repoImpl) ListBySession(ctx context.Context, sessionID domain.LessonSessionID) ([]domain.LessonSessionChallenge, error) {
	const query = `
		SELECT id, session_id, challenge_id, position, status
		FROM lesson_session_challenges
		WHERE session_id = $1
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(sessionID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []domain.LessonSessionChallenge
	for rows.Next() {
		challenge, err := scanSessionChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func scanSessionChallenge(row scanner) (domain.LessonSessionChallenge, error) {
	var challenge domain.LessonSessionChallenge
	var idRaw, sessionIDRaw, challengeIDRaw string
	if err := row.Scan(&idRaw, &sessionIDRaw, &challengeIDRaw, &challenge.Position, &challenge.Status); err != nil {
		return domain.LessonSessionChallenge{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return domain.LessonSessionChallenge{}, err
	}
	sessionID, err := uuid.Parse(sessionIDRaw)
	if err != nil {
		return domain.LessonSessionChallenge{}, err
	}
	challengeID, err := uuid.Parse(challengeIDRaw)
	if err != nil {
		return domain.LessonSessionChallenge{}, err
	}
	challenge.ID = domain.LessonSessionChallengeID(id)
	challenge.SessionID = domain.LessonSessionID(sessionID)
	challenge.ChallengeID = domain.ChallengeID(challengeID)
	return challenge, nil
}
