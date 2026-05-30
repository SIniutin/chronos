package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func (s *Service) ListPublishedSections(ctx context.Context, input api.ListSectionsInput) ([]domain.Section, error) {
	courseID, err := domain.ParseCourseID(input.CourseID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	sections, err := s.sectionsRepo.ListPublishedSections(ctx, courseID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return sections, nil
}

func (s *Service) ListAllSections(ctx context.Context, input api.ListSectionsInput) ([]domain.Section, error) {
	courseID, err := domain.ParseCourseID(input.CourseID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	sections, err := s.sectionsRepo.ListAllSections(ctx, courseID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return sections, nil
}

func (s *Service) CreateSection(ctx context.Context, input api.SectionWriteInput) (domain.Section, error) {
	actorID, err := parseActor(input.ActorID)
	if err != nil {
		return domain.Section{}, err
	}
	courseID, err := domain.ParseCourseID(input.CourseID)
	if err != nil {
		return domain.Section{}, mapDomainError(err)
	}
	now := time.Now().UTC()
	section := domain.Section{
		ID:          domain.SectionID(uuid.New()),
		CourseID:    courseID,
		Theme:       strings.TrimSpace(input.Theme),
		Description: strings.TrimSpace(input.Description),
		Position:    input.Position,
		Status:      domain.ContentStatusDraft,
		Audit:       newAudit(actorID, now),
	}
	created, err := s.sectionsRepo.CreateSection(ctx, section)
	return created, mapDomainError(err)
}

func (s *Service) UpdateSection(ctx context.Context, input api.SectionWriteInput) (domain.Section, error) {
	actorID, id, err := parseActorAndID(input.ActorID, input.ID)
	if err != nil {
		return domain.Section{}, err
	}
	courseID, err := domain.ParseCourseID(input.CourseID)
	if err != nil {
		return domain.Section{}, mapDomainError(err)
	}
	section := domain.Section{
		ID:          domain.SectionID(id),
		CourseID:    courseID,
		Theme:       strings.TrimSpace(input.Theme),
		Description: strings.TrimSpace(input.Description),
		Position:    input.Position,
		Audit:       updateAudit(actorID),
	}
	updated, err := s.sectionsRepo.UpdateSection(ctx, section)
	return updated, mapDomainError(err)
}
