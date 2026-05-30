package users_usecase

import (
	"context"
	"errors"
	"time"

	users_api "github.com/SIniutin/history-app-backend/internal/modules/users/api"
	ud "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type RegisterUsecase struct {
	users  ud.UserRepository
	hasher passwordHasher
}

type passwordHasher interface {
	Hash(password string) (ud.PasswordHash, error)
	Compare(password string, hash ud.PasswordHash) bool
}

func NewRegisterUsecase(users ud.UserRepository, hasher passwordHasher) *RegisterUsecase {
	return &RegisterUsecase{
		users:  users,
		hasher: hasher,
	}
}

func (uc *RegisterUsecase) Exec(ctx context.Context, input users_api.RegisterInput) (users_api.User, error) {
	if uc.users == nil || uc.hasher == nil {
		return users_api.User{}, errors.New("register usecase is not configured")
	}

	email, err := ud.NewEmail(input.Email)
	if err != nil {
		return users_api.User{}, errors.Join(users_api.ErrInvalidInput, err)
	}
	login, err := ud.NewLogin(input.Login)
	if err != nil {
		return users_api.User{}, errors.Join(users_api.ErrInvalidInput, err)
	}
	if err := ud.ValidatePassword(input.Password); err != nil {
		return users_api.User{}, errors.Join(users_api.ErrInvalidInput, err)
	}

	passwordHash, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return users_api.User{}, err
	}

	now := time.Now().UTC()
	userID := uuid.New()
	user, err := uc.users.Create(ctx, ud.CreateUserParams{
		ID:                ud.UserID(userID),
		Email:             *email,
		Login:             *login,
		Role:              ud.RoleStudent,
		PasswordHash:      passwordHash,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return users_api.User{}, mapDomainError(err)
	}

	return toAPIUser(user), nil
}
