package users_domain

import (
	"time"

	"github.com/google/uuid"
)

type UserID uuid.UUID
type SessionID uuid.UUID

type PasswordHashAlgorithm string
type Role string

const (
	PasswordHashArgon2id PasswordHashAlgorithm = "argon2id"
	PasswordHashBcrypt   PasswordHashAlgorithm = "bcrypt"

	RoleStudent         Role = "student"
	RoleContentEditor   Role = "content_editor"
	RoleContentReviewer Role = "content_reviewer"
	RoleAdmin           Role = "admin"
)

type CreateUserParams struct {
	ID                UserID
	Login             Login
	Email             Email
	Role              Role
	PasswordHash      PasswordHash
	PasswordChangedAt time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type User struct {
	ID        UserID
	Login     Login
	Email     Email
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PasswordHash struct {
	Algo  PasswordHashAlgorithm
	Value string // better than byte[] because of DB
}

type Credentials struct {
	UserID            UserID
	Login             Login
	Email             Email
	Role              Role
	PasswordHash      PasswordHash
	PasswordChangedAt time.Time
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessExpires         time.Time
	RefreshTokenExpiresAt time.Time
}

func NewRole(raw string) (Role, error) {
	role := Role(raw)
	switch role {
	case RoleStudent, RoleContentEditor, RoleContentReviewer, RoleAdmin:
		return role, nil
	default:
		return "", ErrInvalidInput
	}
}

type RefreshSession struct {
	ID        SessionID
	UserID    UserID
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}
