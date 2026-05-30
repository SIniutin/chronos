package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func (s *Service) ListPublishedCourses(ctx context.Context) ([]domain.Course, error) {
	courses, err := s.coursesRepo.ListPublishedCourses(ctx)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return courses, nil
}

func (s *Service) ListAllCourses(ctx context.Context) ([]domain.Course, error) {
	courses, err := s.coursesRepo.ListAllCourses(ctx)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return courses, nil
}

func (s *Service) CreateCourse(ctx context.Context, input api.CourseWriteInput) (domain.Course, error) {
	actorID, err := parseActor(input.ActorID)
	if err != nil {
		return domain.Course{}, err
	}
	now := time.Now().UTC()
	course := domain.Course{
		ID:         domain.CourseID(uuid.New()),
		SourceLang: strings.TrimSpace(input.SourceLang),
		TargetLang: strings.TrimSpace(input.TargetLang),
		Title:      strings.TrimSpace(input.Title),
		Status:     domain.ContentStatusDraft,
		Audit:      newAudit(actorID, now),
	}
	if course.SourceLang == "" || course.TargetLang == "" || course.Title == "" {
		return domain.Course{}, api.ErrInvalidInput
	}
	created, err := s.coursesRepo.CreateCourse(ctx, course)
	return created, mapDomainError(err)
}

func (s *Service) UpdateCourse(ctx context.Context, input api.CourseWriteInput) (domain.Course, error) {
	actorID, id, err := parseActorAndID(input.ActorID, input.ID)
	if err != nil {
		return domain.Course{}, err
	}
	course := domain.Course{
		ID:         domain.CourseID(id),
		SourceLang: strings.TrimSpace(input.SourceLang),
		TargetLang: strings.TrimSpace(input.TargetLang),
		Title:      strings.TrimSpace(input.Title),
		Audit:      updateAudit(actorID),
	}
	updated, err := s.coursesRepo.UpdateCourse(ctx, course)
	return updated, mapDomainError(err)
}
