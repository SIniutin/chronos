package users_usecase

import (
	"errors"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

func toAPIUser(user ud.User) users_api.User {
	return users_api.User{
		ID:        uuid.UUID(user.ID).String(),
		Email:     string(user.Email),
		Login:     string(user.Login),
		Role:      users_api.Role(user.Role),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

func toAPITokenPair(pair ud.TokenPair) users_api.TokenPair {
	return users_api.TokenPair{
		AccessToken:           pair.AccessToken,
		RefreshToken:          pair.RefreshToken,
		AccessExpires:         pair.AccessExpires,
		RefreshTokenExpiresAt: pair.RefreshTokenExpiresAt,
	}
}

func mapDomainError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, ud.ErrInvalidInput):
		return errors.Join(users_api.ErrInvalidInput, err)
	case errors.Is(err, ud.ErrConflictEmail):
		return users_api.ErrConflictEmail
	case errors.Is(err, ud.ErrConflictLogin):
		return users_api.ErrConflictLogin
	case errors.Is(err, ud.ErrUserNotFound):
		return users_api.ErrUserNotFound
	case errors.Is(err, ud.ErrInvalidCredential):
		return users_api.ErrInvalidCredential
	case errors.Is(err, users_api.ErrForbidden):
		return users_api.ErrForbidden
	case errors.Is(err, ud.ErrSessionExpired):
		return users_api.ErrSessionExpired
	case errors.Is(err, ud.ErrSessionRevoked):
		return users_api.ErrSessionRevoked
	default:
		return err
	}
}
