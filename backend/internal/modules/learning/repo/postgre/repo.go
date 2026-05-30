package postgre

import (
	"errors"

	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repoImpl struct {
	pool *pgxpool.Pool
}

func NewPostgreRepo(pool *pgxpool.Pool) *repoImpl {
	return &repoImpl{pool: pool}
}

type scanner interface {
	Scan(dest ...any) error
}

func mapPgError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return err
}
