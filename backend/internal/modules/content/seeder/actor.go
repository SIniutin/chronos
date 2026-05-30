package seeder

import (
	"context"
	"fmt"
	"strings"

	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

func ResolveActor(ctx context.Context, users UserFinder, emailRaw string) (string, error) {
	emailRaw = strings.TrimSpace(emailRaw)
	if emailRaw == "" {
		return "", fmt.Errorf("seed actor email is empty")
	}
	email, err := users_domain.NewEmail(emailRaw)
	if err != nil {
		return "", fmt.Errorf("invalid seed actor email %q: %w", emailRaw, err)
	}
	user, err := users.GetByEmail(ctx, *email)
	if err != nil {
		return "", fmt.Errorf("seed actor %q not found: create/bootstrap this admin user before running seed: %w", emailRaw, err)
	}
	return uuid.UUID(user.ID).String(), nil
}
