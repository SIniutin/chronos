package users_usecase

import (
	"context"
	"errors"
	"strings"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type RefreshUsecase struct {
	users    ud.UserRepository
	sessions ud.SessionRepository
	tokens   tokenService
}

func NewRefreshUsecase(users ud.UserRepository, sessions ud.SessionRepository, tokens tokenService) *RefreshUsecase {
	return &RefreshUsecase{users: users, sessions: sessions, tokens: tokens}
}

func (uc *RefreshUsecase) Exec(ctx context.Context, input users_api.RefreshInput) (users_api.TokenPair, error) {
	if uc.users == nil || uc.sessions == nil || uc.tokens == nil {
		return users_api.TokenPair{}, errors.New("refresh usecase is not configured")
	}
	if strings.TrimSpace(input.RefreshToken) == "" {
		return users_api.TokenPair{}, users_api.ErrInvalidInput
	}

	tokenHash := uc.tokens.HashRefresh(input.RefreshToken)
	session, err := uc.sessions.GetRefresh(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ud.ErrSessionNotFound) {
			return users_api.TokenPair{}, users_api.ErrInvalidCredential
		}
		return users_api.TokenPair{}, mapDomainError(err)
	}
	if session.RevokedAt != nil {
		return users_api.TokenPair{}, users_api.ErrSessionRevoked
	}
	if uc.tokens.RefreshExpired(session.ExpiresAt) {
		return users_api.TokenPair{}, users_api.ErrSessionExpired
	}

	user, err := uc.users.GetByID(ctx, session.UserID)
	if err != nil {
		return users_api.TokenPair{}, mapDomainError(err)
	}

	pair, nextSession, err := uc.tokens.NewPair(session.UserID, user.Role)
	if err != nil {
		return users_api.TokenPair{}, err
	}
	if err := uc.sessions.RevokeRefresh(ctx, tokenHash); err != nil {
		return users_api.TokenPair{}, mapDomainError(err)
	}
	if err := uc.sessions.CreateRefresh(ctx, nextSession); err != nil {
		return users_api.TokenPair{}, mapDomainError(err)
	}

	return toAPITokenPair(pair), nil
}
