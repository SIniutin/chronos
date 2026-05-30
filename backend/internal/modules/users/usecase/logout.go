package users_usecase

import (
	"context"
	"errors"
	"strings"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type LogoutUsecase struct {
	sessions ud.SessionRepository
	tokens   tokenService
}

func NewLogoutUsecase(sessions ud.SessionRepository, tokens tokenService) *LogoutUsecase {
	return &LogoutUsecase{sessions: sessions, tokens: tokens}
}

func (uc *LogoutUsecase) Exec(ctx context.Context, input users_api.LogoutInput) error {
	if uc.sessions == nil || uc.tokens == nil {
		return errors.New("logout usecase is not configured")
	}
	if strings.TrimSpace(input.RefreshToken) == "" {
		return users_api.ErrInvalidInput
	}
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return users_api.ErrInvalidInput
	}
	tokenHash := uc.tokens.HashRefresh(input.RefreshToken)
	session, err := uc.sessions.GetRefresh(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ud.ErrSessionNotFound) {
			return nil
		}
		return mapDomainError(err)
	}
	if session.UserID != ud.UserID(userID) {
		return users_api.ErrInvalidCredential
	}
	return mapDomainError(uc.sessions.RevokeRefresh(ctx, tokenHash))
}
