package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func (s *Service) ListPublishedUnits(ctx context.Context, input api.ListUnitsInput) ([]domain.Unit, error) {
	sectionID, err := domain.ParseSectionID(input.SectionID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	units, err := s.unitsRepo.ListPublishedUnits(ctx, sectionID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return units, nil
}

func (s *Service) ListAllUnits(ctx context.Context, input api.ListUnitsInput) ([]domain.Unit, error) {
	sectionID, err := domain.ParseSectionID(input.SectionID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	units, err := s.unitsRepo.ListAllUnits(ctx, sectionID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return units, nil
}

func (s *Service) CreateUnit(ctx context.Context, input api.UnitWriteInput) (domain.Unit, error) {
	actorID, err := parseActor(input.ActorID)
	if err != nil {
		return domain.Unit{}, err
	}
	sectionID, err := domain.ParseSectionID(input.SectionID)
	if err != nil {
		return domain.Unit{}, mapDomainError(err)
	}
	now := time.Now().UTC()
	unit := domain.Unit{
		ID:        domain.UnitID(uuid.New()),
		SectionID: sectionID,
		Title:     strings.TrimSpace(input.Title),
		Position:  input.Position,
		Status:    domain.ContentStatusDraft,
		Audit:     newAudit(actorID, now),
	}
	created, err := s.unitsRepo.CreateUnit(ctx, unit)
	return created, mapDomainError(err)
}

func (s *Service) UpdateUnit(ctx context.Context, input api.UnitWriteInput) (domain.Unit, error) {
	actorID, id, err := parseActorAndID(input.ActorID, input.ID)
	if err != nil {
		return domain.Unit{}, err
	}
	sectionID, err := domain.ParseSectionID(input.SectionID)
	if err != nil {
		return domain.Unit{}, mapDomainError(err)
	}
	unit := domain.Unit{
		ID:        domain.UnitID(id),
		SectionID: sectionID,
		Title:     strings.TrimSpace(input.Title),
		Position:  input.Position,
		Audit:     updateAudit(actorID),
	}
	updated, err := s.unitsRepo.UpdateUnit(ctx, unit)
	return updated, mapDomainError(err)
}
