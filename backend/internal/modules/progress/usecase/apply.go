package usecase

import (
	"context"
	"errors"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
)

func (s *Service) ApplySessionResult(ctx context.Context, input domain.SessionProgressInput) (domain.SessionProgressResult, error) {
	if input.CompletedAt.IsZero() {
		input.CompletedAt = time.Now().UTC()
	}
	if input.TotalAnswers < 0 || input.CorrectAnswers < 0 || input.CorrectAnswers > input.TotalAnswers {
		return domain.SessionProgressResult{}, domain.ErrInvalidInput
	}

	skill, unit, section, err := s.resolveHierarchy(ctx, input.SkillID)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	courseID := section.CourseID

	courseProgress, err := s.ensureCourseProgress(ctx, input.UserID, courseID, input.CompletedAt)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	unitProgress, err := s.ensureUnitProgress(ctx, input.UserID, unit.ID, input.CompletedAt)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	if err := s.ensureFirstSkillAvailable(ctx, input.UserID, courseID, input.CompletedAt); err != nil {
		return domain.SessionProgressResult{}, err
	}

	skillProgress, err := s.ensureSkillProgress(ctx, input.UserID, skill.ID, input.CompletedAt)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	applySkillSession(&skillProgress, input)
	if err := s.repo.SaveSkillProgress(ctx, skillProgress); err != nil {
		return domain.SessionProgressResult{}, err
	}

	result := domain.SessionProgressResult{
		SkillCompleted: skillProgress.Status == domain.ProgressStatusCompleted,
		NewSkillLevel:  skillProgress.Level,
		NewMastery:     skillProgress.Mastery,
	}

	if result.SkillCompleted {
		unlockedSkills, unlockedUnits, err := s.unlockNext(ctx, input.UserID, skill, unit, section, input.CompletedAt)
		if err != nil {
			return domain.SessionProgressResult{}, err
		}
		result.UnlockedSkillIDs = unlockedSkills
		result.UnlockedUnitIDs = unlockedUnits
	}

	unitCompleted, err := s.allSkillsCompleted(ctx, input.UserID, unit.ID)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	if unitCompleted {
		completedAt := input.CompletedAt
		unitProgress.Status = domain.ProgressStatusCompleted
		unitProgress.CompletedAt = &completedAt
		unitProgress.UpdatedAt = input.CompletedAt
		if err := s.repo.SaveUnitProgress(ctx, unitProgress); err != nil {
			return domain.SessionProgressResult{}, err
		}
		result.UnitCompleted = true
	}

	courseCompleted, err := s.allUnitsCompleted(ctx, input.UserID, courseID)
	if err != nil {
		return domain.SessionProgressResult{}, err
	}
	if courseCompleted {
		completedAt := input.CompletedAt
		courseProgress.Status = domain.ProgressStatusCompleted
		courseProgress.CompletedAt = &completedAt
		courseProgress.UpdatedAt = input.CompletedAt
		if err := s.repo.SaveCourseProgress(ctx, courseProgress); err != nil {
			return domain.SessionProgressResult{}, err
		}
		result.CourseCompleted = true
	}

	return result, nil
}

func (s *Service) resolveHierarchy(ctx context.Context, skillID content_domain.SkillID) (content_domain.Skill, content_domain.Unit, content_domain.Section, error) {
	skill, err := s.content.GetSkill(ctx, skillID)
	if err != nil {
		return content_domain.Skill{}, content_domain.Unit{}, content_domain.Section{}, err
	}
	unit, err := s.content.GetUnit(ctx, skill.UnitID)
	if err != nil {
		return content_domain.Skill{}, content_domain.Unit{}, content_domain.Section{}, err
	}
	section, err := s.content.GetSection(ctx, unit.SectionID)
	if err != nil {
		return content_domain.Skill{}, content_domain.Unit{}, content_domain.Section{}, err
	}
	return skill, unit, section, nil
}

func (s *Service) ensureCourseProgress(ctx context.Context, userID domain.UserID, courseID domain.CourseID, now time.Time) (domain.CourseProgress, error) {
	progress, err := s.repo.GetCourseProgress(ctx, userID, courseID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.CourseProgress{}, err
	}
	if progress != nil {
		if progress.Status == domain.ProgressStatusAvailable || progress.Status == domain.ProgressStatusLocked {
			progress.Status = domain.ProgressStatusInProgress
		}
		progress.UpdatedAt = now
		return *progress, s.repo.SaveCourseProgress(ctx, *progress)
	}
	created := domain.CourseProgress{UserID: userID, CourseID: courseID, Status: domain.ProgressStatusInProgress, StartedAt: now, UpdatedAt: now}
	return created, s.repo.SaveCourseProgress(ctx, created)
}

func (s *Service) ensureUnitProgress(ctx context.Context, userID domain.UserID, unitID domain.UnitID, now time.Time) (domain.UnitProgress, error) {
	progress, err := s.repo.GetUnitProgress(ctx, userID, unitID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.UnitProgress{}, err
	}
	if progress != nil {
		if progress.Status == domain.ProgressStatusAvailable || progress.Status == domain.ProgressStatusLocked {
			progress.Status = domain.ProgressStatusInProgress
		}
		progress.UpdatedAt = now
		return *progress, s.repo.SaveUnitProgress(ctx, *progress)
	}
	created := domain.UnitProgress{UserID: userID, UnitID: unitID, Status: domain.ProgressStatusInProgress, StartedAt: now, UpdatedAt: now}
	return created, s.repo.SaveUnitProgress(ctx, created)
}

func (s *Service) ensureSkillProgress(ctx context.Context, userID domain.UserID, skillID domain.SkillID, now time.Time) (domain.SkillProgress, error) {
	progress, err := s.repo.GetSkillProgress(ctx, userID, skillID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return domain.SkillProgress{}, err
	}
	if progress != nil {
		if progress.Status == domain.ProgressStatusAvailable || progress.Status == domain.ProgressStatusLocked {
			progress.Status = domain.ProgressStatusInProgress
		}
		progress.UpdatedAt = now
		return *progress, nil
	}
	return domain.SkillProgress{UserID: userID, SkillID: skillID, Status: domain.ProgressStatusInProgress, StartedAt: now, UpdatedAt: now}, nil
}

func applySkillSession(progress *domain.SkillProgress, input domain.SessionProgressInput) {
	accuracy := 0.0
	if input.TotalAnswers > 0 {
		accuracy = float64(input.CorrectAnswers) / float64(input.TotalAnswers)
	}
	progress.Mastery = progress.Mastery*oldMasteryWeight + accuracy*sessionAccuracyWeight
	progress.CorrectAnswers += input.CorrectAnswers
	progress.WrongAnswers += input.TotalAnswers - input.CorrectAnswers
	if progress.Level < maxSkillLevel {
		progress.Level++
	}
	completedAt := input.CompletedAt
	progress.Status = domain.ProgressStatusCompleted
	progress.CompletedAt = &completedAt
	progress.UpdatedAt = input.CompletedAt
}
