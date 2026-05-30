package postgre

import (
	"context"
	"errors"
	"time"

	d "github.com/SIniutin/history-app-backend/internal/modules/users/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repoImpl struct {
	pool *pgxpool.Pool
}

func NewPostgreRepo(pool *pgxpool.Pool) *repoImpl {
	return &repoImpl{pool: pool}
}

func (r *repoImpl) Create(ctx context.Context, params d.CreateUserParams) (d.User, error) {
	const query = `
		INSERT INTO users (
			id,
			email,
			login,
			role,
			password_hash,
			password_hash_algo,
			password_changed_at,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, email, login, role, created_at, updated_at
	`

	row := r.pool.QueryRow(
		ctx,
		query,
		uuid.UUID(params.ID).String(),
		string(params.Email),
		string(params.Login),
		string(params.Role),
		params.PasswordHash.Value,
		string(params.PasswordHash.Algo),
		params.PasswordChangedAt,
		params.CreatedAt,
		params.UpdatedAt,
	)

	user, err := scanUser(row)
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) Update(ctx context.Context, u *d.User) (d.User, error) {
	const query = `
		UPDATE users
		SET email = $2, login = $3, updated_at = now()
		WHERE id = $1
		RETURNING id, email, login, role, created_at, updated_at
	`
	user, err := scanUser(r.pool.QueryRow(
		ctx,
		query,
		uuid.UUID(u.ID).String(),
		string(u.Email),
		string(u.Login),
	))
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) Delete(ctx context.Context, u *d.User) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, uuid.UUID(u.ID).String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return d.ErrUserNotFound
	}
	return nil
}

func (r *repoImpl) GetByID(ctx context.Context, id d.UserID) (d.User, error) {
	const query = `SELECT id, email, login, role, created_at, updated_at FROM users WHERE id = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) GetByEmail(ctx context.Context, email d.Email) (d.User, error) {
	const query = `SELECT id, email, login, role, created_at, updated_at FROM users WHERE email = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, query, string(email)))
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) GetByLogin(ctx context.Context, login d.Login) (d.User, error) {
	const query = `SELECT id, email, login, role, created_at, updated_at FROM users WHERE login = $1`
	user, err := scanUser(r.pool.QueryRow(ctx, query, string(login)))
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) GetCredentials(ctx context.Context, id d.UserID) (d.Credentials, error) {
	const query = `
		SELECT id, email, login, role, password_hash, password_hash_algo, password_changed_at
		FROM users
		WHERE id = $1
	`
	credentials, err := scanCredentials(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
	if err != nil {
		return d.Credentials{}, mapPgError(err)
	}
	return credentials, nil
}

func (r *repoImpl) GetCredentialsByEmail(ctx context.Context, email d.Email) (d.Credentials, error) {
	const query = `
		SELECT id, email, login, role, password_hash, password_hash_algo, password_changed_at
		FROM users
		WHERE email = $1
	`
	credentials, err := scanCredentials(r.pool.QueryRow(ctx, query, string(email)))
	if err != nil {
		return d.Credentials{}, mapPgError(err)
	}
	return credentials, nil
}

func (r *repoImpl) GetCredentialsByLogin(ctx context.Context, login d.Login) (d.Credentials, error) {
	const query = `
		SELECT id, email, login, role, password_hash, password_hash_algo, password_changed_at
		FROM users
		WHERE login = $1
	`
	credentials, err := scanCredentials(r.pool.QueryRow(ctx, query, string(login)))
	if err != nil {
		return d.Credentials{}, mapPgError(err)
	}
	return credentials, nil
}

func (r *repoImpl) CreateRefresh(ctx context.Context, s d.RefreshSession) error {
	const query = `
		INSERT INTO refresh_sessions (id, user_id, token_hash, expires_at, revoked_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		uuid.UUID(s.ID).String(),
		uuid.UUID(s.UserID).String(),
		s.TokenHash,
		s.ExpiresAt,
		s.RevokedAt,
		s.CreatedAt,
	)
	return mapPgError(err)
}

func (r *repoImpl) ChangeRole(ctx context.Context, id d.UserID, role d.Role) (d.User, error) {
	const query = `
		UPDATE users
		SET role = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, email, login, role, created_at, updated_at
	`
	user, err := scanUser(r.pool.QueryRow(ctx, query, uuid.UUID(id).String(), string(role)))
	if err != nil {
		return d.User{}, mapPgError(err)
	}
	return user, nil
}

func (r *repoImpl) GetRefresh(ctx context.Context, tokenHash string) (d.RefreshSession, error) {
	const query = `
		SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
		FROM refresh_sessions
		WHERE token_hash = $1
	`
	var (
		idRaw     string
		userIDRaw string
		session   d.RefreshSession
		revokedAt *time.Time
	)
	err := r.pool.QueryRow(ctx, query, tokenHash).Scan(
		&idRaw,
		&userIDRaw,
		&session.TokenHash,
		&session.ExpiresAt,
		&revokedAt,
		&session.CreatedAt,
	)
	if err != nil {
		return d.RefreshSession{}, mapSessionError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return d.RefreshSession{}, err
	}
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return d.RefreshSession{}, err
	}
	session.ID = d.SessionID(id)
	session.UserID = d.UserID(userID)
	session.RevokedAt = revokedAt
	return session, nil
}

func (r *repoImpl) RevokeRefresh(ctx context.Context, tokenHash string) error {
	tag, err := r.pool.Exec(
		ctx,
		`UPDATE refresh_sessions SET revoked_at = now() WHERE token_hash = $1 AND revoked_at IS NULL`,
		tokenHash,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return d.ErrSessionNotFound
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanUser(row rowScanner) (d.User, error) {
	var (
		idRaw    string
		emailRaw string
		loginRaw string
		user     d.User
	)
	var roleRaw string
	if err := row.Scan(&idRaw, &emailRaw, &loginRaw, &roleRaw, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return d.User{}, err
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return d.User{}, err
	}
	email, err := d.NewEmail(emailRaw)
	if err != nil {
		return d.User{}, err
	}
	login, err := d.NewLogin(loginRaw)
	if err != nil {
		return d.User{}, err
	}
	user.ID = d.UserID(id)
	user.Email = *email
	user.Login = *login
	role, err := d.NewRole(roleRaw)
	if err != nil {
		return d.User{}, err
	}
	user.Role = role
	return user, nil
}

func scanCredentials(row rowScanner) (d.Credentials, error) {
	var (
		idRaw    string
		emailRaw string
		loginRaw string
		roleRaw  string
		algoRaw  string
		c        d.Credentials
	)
	if err := row.Scan(
		&idRaw,
		&emailRaw,
		&loginRaw,
		&roleRaw,
		&c.PasswordHash.Value,
		&algoRaw,
		&c.PasswordChangedAt,
	); err != nil {
		return d.Credentials{}, err
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return d.Credentials{}, err
	}
	email, err := d.NewEmail(emailRaw)
	if err != nil {
		return d.Credentials{}, err
	}
	login, err := d.NewLogin(loginRaw)
	if err != nil {
		return d.Credentials{}, err
	}
	c.UserID = d.UserID(id)
	c.Email = *email
	c.Login = *login
	role, err := d.NewRole(roleRaw)
	if err != nil {
		return d.Credentials{}, err
	}
	c.Role = role
	c.PasswordHash.Algo = d.PasswordHashAlgorithm(algoRaw)
	return c, nil
}

func mapPgError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return d.ErrUserNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		switch pgErr.ConstraintName {
		case "users_email_key":
			return d.ErrConflictEmail
		case "users_login_key":
			return d.ErrConflictLogin
		default:
			return d.ErrInvalidInput
		}
	}
	return err
}

func mapSessionError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return d.ErrSessionNotFound
	}
	return mapPgError(err)
}
