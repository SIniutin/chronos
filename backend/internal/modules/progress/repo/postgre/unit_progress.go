package postgre

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func (r *repoImpl) GetUnitProgress(ctx context.Context, userID domain.UserID, unitID domain.UnitID) (*domain.UnitProgress, error) {
	const query = `
		SELECT user_id, unit_id, status, started_at, completed_at, updated_at
		FROM unit_progress
		WHERE user_id = $1 AND unit_id = $2
	`
	progress, err := scanUnitProgress(r.pool.QueryRow(ctx, query, uuid.UUID(userID).String(), uuid.UUID(unitID).String()))
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

func (r *repoImpl) SaveUnitProgress(ctx context.Context, progress domain.UnitProgress) error {
	const query = `
		INSERT INTO unit_progress (user_id, unit_id, status, started_at, completed_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id, unit_id) DO UPDATE SET
			status = EXCLUDED.status,
			completed_at = EXCLUDED.completed_at,
			updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query, uuid.UUID(progress.UserID).String(), uuid.UUID(progress.UnitID).String(), progress.Status, progress.StartedAt, progress.CompletedAt, progress.UpdatedAt)
	return err
}

func scanUnitProgress(row scanner) (domain.UnitProgress, error) {
	var progress domain.UnitProgress
	var userIDRaw, unitIDRaw string
	if err := row.Scan(&userIDRaw, &unitIDRaw, &progress.Status, &progress.StartedAt, &progress.CompletedAt, &progress.UpdatedAt); err != nil {
		return domain.UnitProgress{}, mapPgError(err)
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return domain.UnitProgress{}, err
	}
	unitID, err := uuid.Parse(unitIDRaw)
	if err != nil {
		return domain.UnitProgress{}, err
	}
	progress.UserID = domain.UserID(userID)
	progress.UnitID = domain.UnitID(unitID)
	return progress, nil
}
