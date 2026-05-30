package users_domain

import (
	"context"
)

type SessionRepository interface {
	CreateRefresh(ctx context.Context, s RefreshSession) error
	GetRefresh(ctx context.Context, tokenHash string) (RefreshSession, error)
	RevokeRefresh(ctx context.Context, tokenHash string) error
}

type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (User, error)
	Update(ctx context.Context, u *User) (User, error)
	Delete(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id UserID) (User, error)
	GetByEmail(ctx context.Context, email Email) (User, error)
	GetByLogin(ctx context.Context, login Login) (User, error)
	GetCredentialsByEmail(ctx context.Context, email Email) (Credentials, error)
	GetCredentialsByLogin(ctx context.Context, login Login) (Credentials, error)
	GetCredentials(ctx context.Context, id UserID) (Credentials, error)
	ChangeRole(ctx context.Context, id UserID, role Role) (User, error)
}
