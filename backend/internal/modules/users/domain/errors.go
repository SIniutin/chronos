package users_domain

import "errors"

var (
	ErrEmptyPassword     = errors.New("the password is empty")
	ErrPasswordTooShort  = errors.New("password is too short")
	ErrConflictEmail     = errors.New("user with this email already exist")
	ErrConflictLogin     = errors.New("user with this login already exist")
	ErrUserNotFound      = errors.New("user not found")
	ErrSessionNotFound   = errors.New("session not found")
	ErrSessionRevoked    = errors.New("session is revoked")
	ErrSessionExpired    = errors.New("session is expired")
	ErrInvalidCredential = errors.New("invalid credentials")
	ErrInvalidInput      = errors.New("invalid input")
)
