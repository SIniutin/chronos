package usecase

import (
	"context"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/api"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func (s *Service) CompleteAllForUser(ctx context.Context, userIDRaw string) (api.CatalogProgress, error) {
	userUUID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return api.CatalogProgress{}, domain.ErrInvalidInput
	}
	userID := domain.UserID(userUUID)
	now := time.Now().UTC()
	completedAt := now

	courses, err := s.content.ListPublishedCourses(ctx)
	if err != nil {
		return api.CatalogProgress{}, err
	}

	var lastCourseID domain.CourseID
	for _, course := range courses {
		lastCourseID = course.ID
		if err := s.repo.SaveCourseProgress(ctx, domain.CourseProgress{
			UserID:      userID,
			CourseID:    course.ID,
			Status:      domain.ProgressStatusCompleted,
			StartedAt:   now,
			CompletedAt: &completedAt,
			UpdatedAt:   now,
		}); err != nil {
			return api.CatalogProgress{}, err
		}

		sections, err := s.content.ListPublishedSections(ctx, course.ID)
		if err != nil {
			return api.CatalogProgress{}, err
		}
		for _, section := range sections {
			units, err := s.content.ListPublishedUnits(ctx, section.ID)
			if err != nil {
				return api.CatalogProgress{}, err
			}
			for _, unit := range units {
				if err := s.repo.SaveUnitProgress(ctx, domain.UnitProgress{
					UserID:      userID,
					UnitID:      unit.ID,
					Status:      domain.ProgressStatusCompleted,
					StartedAt:   now,
					CompletedAt: &completedAt,
					UpdatedAt:   now,
				}); err != nil {
					return api.CatalogProgress{}, err
				}

				skills, err := s.content.ListPublishedSkills(ctx, unit.ID)
				if err != nil {
					return api.CatalogProgress{}, err
				}
				for _, skill := range skills {
					if err := s.repo.SaveSkillProgress(ctx, domain.SkillProgress{
						UserID:         userID,
						SkillID:        skill.ID,
						Status:         domain.ProgressStatusCompleted,
						Level:          maxSkillLevel,
						Mastery:        1,
						CorrectAnswers: 0,
						WrongAnswers:   0,
						StartedAt:      now,
						CompletedAt:    &completedAt,
						UpdatedAt:      now,
					}); err != nil {
						return api.CatalogProgress{}, err
					}
				}
			}
		}
	}

	if uuid.UUID(lastCourseID) == uuid.Nil {
		return api.CatalogProgress{}, nil
	}
	return s.GetCatalogProgress(ctx, userIDRaw, uuid.UUID(lastCourseID).String())
}
