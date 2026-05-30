package postgre

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) GetSkillProgress(ctx context.Context, userID domain.UserID, skillID domain.SkillID) (*domain.SkillProgress, error) {
	const query = `
		SELECT user_id, skill_id, status, level, mastery, correct_answers, wrong_answers, started_at, completed_at, updated_at
		FROM skill_progress
		WHERE user_id = $1 AND skill_id = $2
	`
	progress, err := scanSkillProgress(r.pool.QueryRow(ctx, query, uuid.UUID(userID).String(), uuid.UUID(skillID).String()))
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *repoImpl) SaveSkillProgress(ctx context.Context, progress domain.SkillProgress) error {
	const query = `
		INSERT INTO skill_progress (user_id, skill_id, status, level, mastery, correct_answers, wrong_answers, started_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (user_id, skill_id) DO UPDATE SET
			status = EXCLUDED.status,
			level = EXCLUDED.level,
			mastery = EXCLUDED.mastery,
			correct_answers = EXCLUDED.correct_answers,
			wrong_answers = EXCLUDED.wrong_answers,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(progress.UserID).String(), uuid.UUID(progress.SkillID).String(), progress.Status, progress.Level, progress.Mastery, progress.CorrectAnswers, progress.WrongAnswers, progress.StartedAt, progress.CompletedAt, progress.UpdatedAt)
	return err
}

func (r *repoImpl) ListSkillProgressByUser(ctx context.Context, userID domain.UserID) ([]domain.SkillProgress, error) {
	const query = `
		SELECT user_id, skill_id, status, level, mastery, correct_answers, wrong_answers, started_at, completed_at, updated_at
		FROM skill_progress
		WHERE user_id = $1
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(userID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SkillProgress
	for rows.Next() {
		progress, err := scanSkillProgress(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, progress)
	}
	return out, rows.Err()
}

func scanSkillProgress(row scanner) (domain.SkillProgress, error) {
	var progress domain.SkillProgress
	var userIDRaw, skillIDRaw string
	if err := row.Scan(&userIDRaw, &skillIDRaw, &progress.Status, &progress.Level, &progress.Mastery, &progress.CorrectAnswers, &progress.WrongAnswers, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt); err != nil {
		return domain.SkillProgress{}, mapPgError(err)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return domain.SkillProgress{}, err
	}
	skillID, err := uuid.Parse(skillIDRaw)
	if err != nil {
		return domain.SkillProgress{}, err
	}
	progress.UserID = domain.UserID(userID)
	progress.SkillID = domain.SkillID(skillID)
	return progress, nil
}
