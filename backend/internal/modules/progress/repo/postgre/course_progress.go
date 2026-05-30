package postgre

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) GetCourseProgress(ctx context.Context, userID domain.UserID, courseID domain.CourseID) (*domain.CourseProgress, error) {
	const query = `
		SELECT user_id, course_id, status, started_at, completed_at, updated_at
		FROM course_progress
		WHERE user_id = $1 AND course_id = $2
	`
	progress, err := scanCourseProgress(r.pool.QueryRow(ctx, query, uuid.UUID(userID).String(), uuid.UUID(courseID).String()))
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *repoImpl) SaveCourseProgress(ctx context.Context, progress domain.CourseProgress) error {
	const query = `
		INSERT INTO course_progress (user_id, course_id, status, started_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, course_id) DO UPDATE SET
			status = EXCLUDED.status,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(progress.UserID).String(), uuid.UUID(progress.CourseID).String(), progress.Status, progress.StartedAt, progress.CompletedAt, progress.UpdatedAt)
	return err
}

func scanCourseProgress(row scanner) (domain.CourseProgress, error) {
	var progress domain.CourseProgress
	var userIDRaw, courseIDRaw string
	if err := row.Scan(&userIDRaw, &courseIDRaw, &progress.Status, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt); err != nil {
		return domain.CourseProgress{}, mapPgError(err)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return domain.CourseProgress{}, err
	}
	courseID, err := uuid.Parse(courseIDRaw)
	if err != nil {
		return domain.CourseProgress{}, err
	}
	progress.UserID = domain.UserID(userID)
	progress.CourseID = domain.CourseID(courseID)
	return progress, nil
}
