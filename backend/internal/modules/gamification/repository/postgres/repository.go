package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/gamification/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type scanner interface{ Scan(dest ...any) error }

func mapErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrDuplicateXP
	}
	return err
}

func (r *Repository) GetUserXP(ctx context.Context, userID domain.UserID) (*domain.UserXP, error) {
	row := r.pool.QueryRow(ctx, `SELECT user_id,total_xp,level,updated_at FROM user_xp WHERE user_id=$1`, uuid.UUID(userID).String())
	xp, err := scanUserXP(row)
	if err != nil {
		return nil, err
	}
	return &xp, nil
}

func (r *Repository) SaveUserXP(ctx context.Context, xp domain.UserXP) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO user_xp (user_id,total_xp,level,updated_at) VALUES ($1,$2,$3,$4) ON CONFLICT (user_id) DO UPDATE SET total_xp=EXCLUDED.total_xp, level=EXCLUDED.level, updated_at=EXCLUDED.updated_at`, uuid.UUID(xp.UserID).String(), xp.TotalXP, xp.Level, xp.UpdatedAt)
	return err
}

func (r *Repository) CreateXPTransaction(ctx context.Context, tx domain.XPTransaction) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO xp_transactions (id,user_id,amount,reason,source_type,source_id,created_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`, tx.ID.String(), uuid.UUID(tx.UserID).String(), tx.Amount, tx.Reason, tx.SourceType, tx.SourceID.String(), tx.CreatedAt)
	return mapErr(err)
}

func (r *Repository) GetUserStreak(ctx context.Context, userID domain.UserID) (*domain.UserStreak, error) {
	row := r.pool.QueryRow(ctx, `SELECT user_id,current_days,longest_days,last_activity_date,updated_at FROM user_streaks WHERE user_id=$1`, uuid.UUID(userID).String())
	streak, err := scanStreak(row)
	if err != nil {
		return nil, err
	}
	return &streak, nil
}

func (r *Repository) SaveUserStreak(ctx context.Context, streak domain.UserStreak) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO user_streaks (user_id,current_days,longest_days,last_activity_date,updated_at) VALUES ($1,$2,$3,$4,$5) ON CONFLICT (user_id) DO UPDATE SET current_days=EXCLUDED.current_days,longest_days=EXCLUDED.longest_days,last_activity_date=EXCLUDED.last_activity_date,updated_at=EXCLUDED.updated_at`, uuid.UUID(streak.UserID).String(), streak.CurrentDays, streak.LongestDays, streak.LastActivityDate, streak.UpdatedAt)
	return err
}

func (r *Repository) ListAchievements(ctx context.Context) ([]domain.Achievement, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,code,title,description,xp_reward,condition,created_at FROM achievements ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Achievement
	for rows.Next() {
		a, err := scanAchievement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) ListUserAchievements(ctx context.Context, userID domain.UserID) ([]domain.Achievement, error) {
	rows, err := r.pool.Query(ctx, `SELECT a.id,a.code,a.title,a.description,a.xp_reward,a.condition,a.created_at FROM achievements a JOIN user_achievements ua ON ua.achievement_id=a.id WHERE ua.user_id=$1 ORDER BY ua.unlocked_at`, uuid.UUID(userID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Achievement
	for rows.Next() {
		a, err := scanAchievement(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *Repository) UnlockAchievement(ctx context.Context, userID domain.UserID, achievementID uuid.UUID, unlockedAt time.Time) (bool, error) {
	tag, err := r.pool.Exec(ctx, `INSERT INTO user_achievements (user_id,achievement_id,unlocked_at) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, uuid.UUID(userID).String(), achievementID.String(), unlockedAt)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func scanUserXP(row scanner) (domain.UserXP, error) {
	var xp domain.UserXP
	var userID string
	if err := row.Scan(&userID, &xp.TotalXP, &xp.Level, &xp.UpdatedAt); err != nil {
		return domain.UserXP{}, mapErr(err)
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return domain.UserXP{}, err
	}
	xp.UserID = domain.UserID(id)
	return xp, nil
}

func scanStreak(row scanner) (domain.UserStreak, error) {
	var s domain.UserStreak
	var userID string
	if err := row.Scan(&userID, &s.CurrentDays, &s.LongestDays, &s.LastActivityDate, &s.UpdatedAt); err != nil {
		return domain.UserStreak{}, mapErr(err)
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return domain.UserStreak{}, err
	}
	s.UserID = domain.UserID(id)
	return s, nil
}

func scanAchievement(row scanner) (domain.Achievement, error) {
	var a domain.Achievement
	var id string
	if err := row.Scan(&id, &a.Code, &a.Title, &a.Description, &a.XPReward, &a.Condition, &a.CreatedAt); err != nil {
		return domain.Achievement{}, mapErr(err)
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return domain.Achievement{}, err
	}
	a.ID = parsed
	return a, nil
}
