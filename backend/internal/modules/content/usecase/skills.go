package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func (s *Service) ListPublishedSkills(ctx context.Context, input api.ListSkillsInput) ([]domain.Skill, error) {
	unitID, err := domain.ParseUnitID(input.UnitID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	skills, err := s.skillsRepo.ListPublishedSkills(ctx, unitID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return skills, nil
}

func (s *Service) ListAllSkills(ctx context.Context, input api.ListSkillsInput) ([]domain.Skill, error) {
	unitID, err := domain.ParseUnitID(input.UnitID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	skills, err := s.skillsRepo.ListAllSkills(ctx, unitID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return skills, nil
}

func (s *Service) CreateSkill(ctx context.Context, input api.SkillWriteInput) (domain.Skill, error) {
	actorID, err := parseActor(input.ActorID)
	if err != nil {
		return domain.Skill{}, err
	}
	unitID, err := domain.ParseUnitID(input.UnitID)
	if err != nil {
		return domain.Skill{}, mapDomainError(err)
	}
	now := time.Now().UTC()
	skill := domain.Skill{
		ID:       domain.SkillID(uuid.New()),
		UnitID:   unitID,
		Title:    strings.TrimSpace(input.Title),
		Icon:     strings.TrimSpace(input.Icon),
		Position: input.Position,
		Status:   domain.ContentStatusDraft,
		Audit:    newAudit(actorID, now),
	}
	created, err := s.skillsRepo.CreateSkill(ctx, skill)
	return created, mapDomainError(err)
}

func (s *Service) UpdateSkill(ctx context.Context, input api.SkillWriteInput) (domain.Skill, error) {
	actorID, id, err := parseActorAndID(input.ActorID, input.ID)
	if err != nil {
		return domain.Skill{}, err
	}
	unitID, err := domain.ParseUnitID(input.UnitID)
	if err != nil {
		return domain.Skill{}, mapDomainError(err)
	}
	skill := domain.Skill{
		ID:       domain.SkillID(id),
		UnitID:   unitID,
		Title:    strings.TrimSpace(input.Title),
		Icon:     strings.TrimSpace(input.Icon),
		Position: input.Position,
		Audit:    updateAudit(actorID),
	}
	updated, err := s.skillsRepo.UpdateSkill(ctx, skill)
	return updated, mapDomainError(err)
}
