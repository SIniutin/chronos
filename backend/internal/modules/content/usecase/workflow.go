package usecase

import (
	"context"
	"strings"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

func (s *Service) Publish(ctx context.Context, input api.StatusTransitionInput) error {
	actorID, id, err := parseTransition(input)
	if err != nil {
		return err
	}
	switch strings.TrimSpace(input.Entity) {
	case "courses", "course", "sections", "section", "units", "unit", "skills", "skill", "challenges", "challenge":
		return mapDomainError(s.statusRepo.SetStatus(ctx, input.Entity, id, domain.ContentStatusPublished, actorID))
	default:
		return api.ErrInvalidInput
	}
}

func (s *Service) Archive(ctx context.Context, input api.StatusTransitionInput) error {
	actorID, id, err := parseTransition(input)
	if err != nil {
		return err
	}
	switch strings.TrimSpace(input.Entity) {
	case "courses", "course", "sections", "section", "units", "unit", "skills", "skill", "challenges", "challenge":
		return mapDomainError(s.statusRepo.SetStatus(ctx, input.Entity, id, domain.ContentStatusArchived, actorID))
	default:
		return api.ErrInvalidInput
	}
}
