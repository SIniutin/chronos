package postgre

import (
	"context"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) CreateSession(ctx context.Context, session *domain.LessonSession) error {
	const query = `
		INSERT INTO lesson_sessions (id, user_id, skill_id, status, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(session.ID).String(), uuid.UUID(session.UserID).String(), uuid.UUID(session.SkillID).String(), session.Status, session.StartedAt, session.FinishedAt)
	return err
}

func (r *repoImpl) GetSession(ctx context.Context, id domain.LessonSessionID) (*domain.LessonSession, error) {
	const query = `
		SELECT id, user_id, skill_id, status, started_at, finished_at
		FROM lesson_sessions
		WHERE id = $1
	`
	session, err := scanSession(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *repoImpl) UpdateSession(ctx context.Context, session *domain.LessonSession) error {
	const query = `
		UPDATE lesson_sessions
		SET status = $2, finished_at = $3
		WHERE id = $1
	`
	tag, err := r.pool.Exec(ctx, query, uuid.UUID(session.ID).String(), session.Status, session.FinishedAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func scanSession(row scanner) (domain.LessonSession, error) {
	var session domain.LessonSession
	var idRaw, userIDRaw, skillIDRaw string
	var finishedAt *time.Time
	if err := row.Scan(&idRaw, &userIDRaw, &skillIDRaw, &session.Status, &session.StartedAt, &finishedAt); err != nil {
		return domain.LessonSession{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return domain.LessonSession{}, err
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return domain.LessonSession{}, err
	}
	skillID, err := uuid.Parse(skillIDRaw)
	if err != nil {
		return domain.LessonSession{}, err
	}
	session.ID = domain.LessonSessionID(id)
	session.UserID = domain.UserID(userID)
	session.SkillID = domain.SkillID(skillID)
	session.FinishedAt = finishedAt
	return session, nil
}
