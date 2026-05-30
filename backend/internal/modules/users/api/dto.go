package users_api

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidInput      = errors.New("invalid input")
	ErrConflictEmail     = errors.New("user with this email already exist")
	ErrConflictLogin     = errors.New("user with this login already exist")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrSessionExpired    = errors.New("session is expired")
	ErrSessionRevoked    = errors.New("session is revoked")
	ErrForbidden         = errors.New("forbidden")
)

type Role string

const (
	RoleStudent         Role = "student"
	RoleContentEditor   Role = "content_editor"
	RoleContentReviewer Role = "content_reviewer"
	RoleAdmin           Role = "admin"
)

type RegisterUsecase interface {
	Exec(ctx context.Context, input RegisterInput) (User, error)
}

type LoginUsecase interface {
	Exec(ctx context.Context, input LoginInput) (TokenPair, error)
}

type RefreshUsecase interface {
	Exec(ctx context.Context, input RefreshInput) (TokenPair, error)
}

type LogoutUsecase interface {
	Exec(ctx context.Context, input LogoutInput) error
}

type GetMeUsecase interface {
	Exec(ctx context.Context, input GetMeInput) (User, error)
}

type ChangeUserRoleUsecase interface {
	Exec(ctx context.Context, input ChangeUserRoleInput) (User, error)
}

type FindUserUsecase interface {
	Exec(ctx context.Context, input FindUserInput) (User, error)
}

type RegisterInput struct {
	Email    string
	Login    string
	Password string
}

type LoginInput struct {
	Identity string
	Password string
}

type RefreshInput struct {
	RefreshToken string
}

type LogoutInput struct {
	UserID       string
	RefreshToken string
}

type GetMeInput struct {
	UserID string
}

type ChangeUserRoleInput struct {
	UserID string
	Role   Role
}

type FindUserInput struct {
	Identity string
}

type User struct {
	ID        string
	Email     string
	Login     string
	Role      Role
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TokenPair struct {
	AccessToken           string
	RefreshToken          string
	AccessExpires         time.Time
	RefreshTokenExpiresAt time.Time
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

type UserResponse struct {
	Email     string `json:"email"`
	Login     string `json:"login"`
	Role      Role   `json:"role"`
	CreatedAt string `json:"created_at"`
}

type AdminUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	Role      Role   `json:"role"`
	CreatedAt string `json:"created_at"`
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ChangeUserRoleRequest struct {
	Role Role `json:"role"`
}
