package users_usecase

import (
	"context"
	"errors"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type ChangeUserRoleUsecase struct {
	users ud.UserRepository
}

func NewChangeUserRoleUsecase(users ud.UserRepository) *ChangeUserRoleUsecase {
	return &ChangeUserRoleUsecase{users: users}
}

func (uc *ChangeUserRoleUsecase) Exec(ctx context.Context, input users_api.ChangeUserRoleInput) (users_api.User, error) {
	if uc.users == nil {
		return users_api.User{}, errors.New("change user role usecase is not configured")
	}
	userID, err := uuid.Parse(input.UserID)
	if err != nil {
		return users_api.User{}, users_api.ErrInvalidInput
	}
	role, err := ud.NewRole(string(input.Role))
	if err != nil {
		return users_api.User{}, users_api.ErrInvalidInput
	}
	user, err := uc.users.ChangeRole(ctx, ud.UserID(userID), role)
	if err != nil {
		return users_api.User{}, mapDomainError(err)
	}
	return toAPIUser(user), nil
}
