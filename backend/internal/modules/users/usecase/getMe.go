package users_usecase

import (
	"context"
	"errors"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type GetMeUsecase struct {
	users ud.UserRepository
}

func NewGetMeUsecase(users ud.UserRepository) *GetMeUsecase {
	return &GetMeUsecase{users: users}
}

func (uc *GetMeUsecase) Exec(ctx context.Context, input users_api.GetMeInput) (users_api.User, error) {
	if uc.users == nil {
		return users_api.User{}, errors.New("get me usecase is not configured")
	}
	id, err := uuid.Parse(input.UserID)
	if err != nil {
		return users_api.User{}, users_api.ErrInvalidInput
	}
	user, err := uc.users.GetByID(ctx, ud.UserID(id))
	if err != nil {
		return users_api.User{}, mapDomainError(err)
	}
	return toAPIUser(user), nil
}
