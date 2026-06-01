package usecase

import (
	"context"
	"errors"

	"github.com/SIniutin/history-app-backend/internal/modules/progress/api"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
	"github.com/google/uuid"
)

func (s *Service) GetCatalogProgress(ctx context.Context, userIDRaw string, courseIDRaw string) (api.CatalogProgress, error) {
	userUUID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return api.CatalogProgress{}, domain.ErrInvalidInput
	}
	courseUUID, err := uuid.Parse(courseIDRaw)
	if err != nil {
		return api.CatalogProgress{}, domain.ErrInvalidInput
	}
	userID := domain.UserID(userUUID)
	courseID := domain.CourseID(courseUUID)

	courseProgress, err := s.repo.GetCourseProgress(ctx, userID, courseID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return api.CatalogProgress{}, err
	}
	out := api.CatalogProgress{CourseID: courseIDRaw, CourseStatus: string(domain.ProgressStatusAvailable)}
	if courseProgress != nil {
		out.CourseStatus = string(courseProgress.Status)
	}

	sections, err := s.content.ListPublishedSections(ctx, courseID)
	if err != nil {
		return api.CatalogProgress{}, err
	}

	previousCompleted := true
	for _, section := range sections {
		units, err := s.content.ListPublishedUnits(ctx, section.ID)
		if err != nil {
			return api.CatalogProgress{}, err
		}
		for _, unit := range units {
			unitStatus := domain.ProgressStatusLocked
			unitProgress, err := s.repo.GetUnitProgress(ctx, userID, unit.ID)
			if err != nil && !errors.Is(err, domain.ErrNotFound) {
				return api.CatalogProgress{}, err
			}
			if unitProgress != nil {
				unitStatus = unitProgress.Status
			}

			skills, err := s.content.ListPublishedSkills(ctx, unit.ID)
			if err != nil {
				return api.CatalogProgress{}, err
			}
			unitHasVisible := false
			unitCompleted := len(skills) > 0
			for _, skill := range skills {
				out.TotalLessons++
				progress, err := s.repo.GetSkillProgress(ctx, userID, skill.ID)
				if err != nil && !errors.Is(err, domain.ErrNotFound) {
					return api.CatalogProgress{}, err
				}
				status := domain.ProgressStatusLocked
				level := 0
				mastery := 0.0
				correct := 0
				wrong := 0
				if progress != nil {
					status = progress.Status
					level = progress.Level
					mastery = progress.Mastery
					correct = progress.CorrectAnswers
					wrong = progress.WrongAnswers
				} else if previousCompleted {
					status = domain.ProgressStatusAvailable
				}
				if status != domain.ProgressStatusLocked {
					out.AvailableLessons++
					unitHasVisible = true
				}
				if status == domain.ProgressStatusCompleted {
					out.CompletedLessons++
					previousCompleted = true
				} else {
					unitCompleted = false
					previousCompleted = false
				}
				out.Skills = append(out.Skills, api.SkillProgress{
					SkillID:        uuid.UUID(skill.ID).String(),
					UnitID:         uuid.UUID(unit.ID).String(),
					Status:         string(status),
					Level:          level,
					Mastery:        mastery,
					CorrectAnswers: correct,
					WrongAnswers:   wrong,
				})
			}
			if unitStatus == domain.ProgressStatusLocked && unitHasVisible {
				unitStatus = domain.ProgressStatusAvailable
			}
			if unitCompleted {
				unitStatus = domain.ProgressStatusCompleted
			}
			out.Units = append(out.Units, api.UnitProgress{UnitID: uuid.UUID(unit.ID).String(), Status: string(unitStatus)})
		}
	}
	return out, nil
}
