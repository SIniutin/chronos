package users_usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type LoginUsecase struct {
	users    ud.UserRepository
	sessions ud.SessionRepository
	hasher   passwordHasher
	tokens   tokenService
}

type tokenService interface {
	NewPair(userID ud.UserID, role ud.Role) (ud.TokenPair, ud.RefreshSession, error)
	HashRefresh(token string) string
	RefreshExpired(expiresAt time.Time) bool
}

func NewLoginUsecase(users ud.UserRepository, sessions ud.SessionRepository, hasher passwordHasher, tokens tokenService) *LoginUsecase {
	return &LoginUsecase{
		users:    users,
		sessions: sessions,
		hasher:   hasher,
		tokens:   tokens,
	}
}

func (uc *LoginUsecase) Exec(ctx context.Context, input users_api.LoginInput) (users_api.TokenPair, error) {
	if uc.users == nil || uc.sessions == nil || uc.hasher == nil || uc.tokens == nil {
		return users_api.TokenPair{}, errors.New("login usecase is not configured")
	}
	if strings.TrimSpace(input.Identity) == "" || strings.TrimSpace(input.Password) == "" {
		return users_api.TokenPair{}, users_api.ErrInvalidInput
	}

	credentials, err := uc.getCredentials(ctx, input.Identity)
	if err != nil {
		if errors.Is(err, ud.ErrUserNotFound) {
			return users_api.TokenPair{}, users_api.ErrInvalidCredential
		}
		return users_api.TokenPair{}, mapDomainError(err)
	}
	if !uc.hasher.Compare(input.Password, credentials.PasswordHash) {
		return users_api.TokenPair{}, users_api.ErrInvalidCredential
	}

	pair, session, err := uc.tokens.NewPair(credentials.UserID, credentials.Role)
	if err != nil {
		return users_api.TokenPair{}, err
	}
	if err := uc.sessions.CreateRefresh(ctx, session); err != nil {
		return users_api.TokenPair{}, mapDomainError(err)
	}

	return toAPITokenPair(pair), nil
}

func (uc *LoginUsecase) getCredentials(ctx context.Context, identity string) (ud.Credentials, error) {
	if email, err := ud.NewEmail(identity); err == nil {
		return uc.users.GetCredentialsByEmail(ctx, *email)
	}
	login, err := ud.NewLogin(identity)
	if err != nil {
		return ud.Credentials{}, ud.ErrInvalidCredential
	}
	return uc.users.GetCredentialsByLogin(ctx, *login)
}
