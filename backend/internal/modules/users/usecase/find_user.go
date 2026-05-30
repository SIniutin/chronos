package users_usecase

import (
	"context"
	"strings"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type FindUserUsecase struct {
	users ud.UserRepository
}

func NewFindUserUsecase(users ud.UserRepository) *FindUserUsecase {
	return &FindUserUsecase{users: users}
}

func (uc *FindUserUsecase) Exec(ctx context.Context, input users_api.FindUserInput) (users_api.User, error) {
	identity := strings.TrimSpace(input.Identity)
	if identity == "" {
		return users_api.User{}, users_api.ErrInvalidInput
	}

	if email, err := ud.NewEmail(identity); err == nil {
		user, err := uc.users.GetByEmail(ctx, *email)
		return toAPIUser(user), mapDomainError(err)
	}

	login, err := ud.NewLogin(identity)
	if err != nil {
		return users_api.User{}, users_api.ErrInvalidInput
	}
	user, err := uc.users.GetByLogin(ctx, *login)
	return toAPIUser(user), mapDomainError(err)
}
