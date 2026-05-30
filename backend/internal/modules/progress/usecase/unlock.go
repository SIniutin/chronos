package usecase

import (
	"context"
	"errors"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
)

func (s *Service) ensureFirstSkillAvailable(ctx context.Context, userID domain.UserID, courseID domain.CourseID, now time.Time) error {
	sections, err := s.content.ListPublishedSections(ctx, courseID)
	if err != nil || len(sections) == 0 {
		return err
	}
	units, err := s.content.ListPublishedUnits(ctx, sections[0].ID)
	if err != nil || len(units) == 0 {
		return err
	}
	skills, err := s.content.ListPublishedSkills(ctx, units[0].ID)
	if err != nil || len(skills) == 0 {
		return err
	}
	return s.makeSkillAvailable(ctx, userID, skills[0].ID, now)
}

func (s *Service) unlockNext(ctx context.Context, userID domain.UserID, skill content_domain.Skill, unit content_domain.Unit, section content_domain.Section, now time.Time) ([]domain.SkillID, []domain.UnitID, error) {
	skills, err := s.content.ListPublishedSkills(ctx, unit.ID)
	if err != nil {
		return nil, nil, err
	}
	for i, candidate := range skills {
		if candidate.ID == skill.ID && i+1 < len(skills) {
			if err := s.makeSkillAvailable(ctx, userID, skills[i+1].ID, now); err != nil {
				return nil, nil, err
			}
			return []domain.SkillID{skills[i+1].ID}, nil, nil
		}
	}

	units, err := s.publishedCourseUnits(ctx, section.CourseID)
	if err != nil {
		return nil, nil, err
	}
	for i, candidate := range units {
		if candidate.ID == unit.ID && i+1 < len(units) {
			nextUnit := units[i+1]
			if err := s.makeUnitAvailable(ctx, userID, nextUnit.ID, now); err != nil {
				return nil, nil, err
			}
			nextSkills, err := s.content.ListPublishedSkills(ctx, nextUnit.ID)
			if err != nil {
				return nil, nil, err
			}
			if len(nextSkills) == 0 {
				return nil, []domain.UnitID{nextUnit.ID}, nil
			}
			if err := s.makeSkillAvailable(ctx, userID, nextSkills[0].ID, now); err != nil {
				return nil, nil, err
			}
			return []domain.SkillID{nextSkills[0].ID}, []domain.UnitID{nextUnit.ID}, nil
		}
	}
	return nil, nil, nil
}

func (s *Service) publishedCourseUnits(ctx context.Context, courseID domain.CourseID) ([]content_domain.Unit, error) {
	sections, err := s.content.ListPublishedSections(ctx, courseID)
	if err != nil {
		return nil, err
	}
	var out []content_domain.Unit
	for _, section := range sections {
		units, err := s.content.ListPublishedUnits(ctx, section.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, units...)
	}
	return out, nil
}

func (s *Service) makeSkillAvailable(ctx context.Context, userID domain.UserID, skillID domain.SkillID, now time.Time) error {
	progress, err := s.repo.GetSkillProgress(ctx, userID, skillID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if progress != nil {
		if progress.Status == domain.ProgressStatusLocked {
			progress.Status = domain.ProgressStatusAvailable
			progress.UpdatedAt = now
			return s.repo.SaveSkillProgress(ctx, *progress)
		}
		return nil
	}
	return s.repo.SaveSkillProgress(ctx, domain.SkillProgress{UserID: userID, SkillID: skillID, Status: domain.ProgressStatusAvailable, StartedAt: now, UpdatedAt: now})
}

func (s *Service) makeUnitAvailable(ctx context.Context, userID domain.UserID, unitID domain.UnitID, now time.Time) error {
	progress, err := s.repo.GetUnitProgress(ctx, userID, unitID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if progress != nil {
		if progress.Status == domain.ProgressStatusLocked {
			progress.Status = domain.ProgressStatusAvailable
			progress.UpdatedAt = now
			return s.repo.SaveUnitProgress(ctx, *progress)
		}
		return nil
	}
	return s.repo.SaveUnitProgress(ctx, domain.UnitProgress{UserID: userID, UnitID: unitID, Status: domain.ProgressStatusAvailable, StartedAt: now, UpdatedAt: now})
}

func (s *Service) allSkillsCompleted(ctx context.Context, userID domain.UserID, unitID domain.UnitID) (bool, error) {
	skills, err := s.content.ListPublishedSkills(ctx, unitID)
	if err != nil || len(skills) == 0 {
		return false, err
	}
	for _, skill := range skills {
		progress, err := s.repo.GetSkillProgress(ctx, userID, skill.ID)
		if err != nil {
			if errors.Is(err, domain.ErrNotFound) {
				return false, nil
			}
			return false, err
		}
		if progress.Status != domain.ProgressStatusCompleted {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) allUnitsCompleted(ctx context.Context, userID domain.UserID, courseID domain.CourseID) (bool, error) {
	sections, err := s.content.ListPublishedSections(ctx, courseID)
	if err != nil || len(sections) == 0 {
		return false, err
	}
	seenUnit := false
	for _, section := range sections {
		units, err := s.content.ListPublishedUnits(ctx, section.ID)
		if err != nil {
			return false, err
		}
		for _, unit := range units {
			seenUnit = true
			progress, err := s.repo.GetUnitProgress(ctx, userID, unit.ID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return false, nil
				}
				return false, err
			}
			if progress.Status != domain.ProgressStatusCompleted {
				return false, nil
			}
		}
	}
	return seenUnit, nil
}
