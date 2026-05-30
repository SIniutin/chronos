package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/gamification/domain"
	learning_usecase "github.com/SIniutin/history-app-backend/internal/modules/learning/usecase"
	"github.com/google/uuid"
)

const baseSessionXP = 10
const perCorrectXP = 2
const perfectBonusXP = 5

type Service struct{ repo domain.Repository }

func NewService(repo domain.Repository) *Service { return &Service{repo: repo} }

func (s *Service) RewardSessionCompleted(ctx context.Context, input domain.SessionRewardInput) (*domain.SessionRewardResult, error) {
	if input.CompletedAt.IsZero() {
		input.CompletedAt = time.Now().UTC()
	}
	if input.CorrectAnswers < 0 || input.TotalAnswers < 0 || input.CorrectAnswers > input.TotalAnswers {
		return nil, domain.ErrInvalidInput
	}
	amount := baseSessionXP + input.CorrectAnswers*perCorrectXP
	if input.TotalAnswers > 0 && input.CorrectAnswers == input.TotalAnswers {
		amount += perfectBonusXP
	}
	err := s.AddXP(ctx, input.UserID, amount, domain.XPReasonSessionCompleted, "lesson_session", input.SessionID)
	duplicate := errors.Is(err, domain.ErrDuplicateXP)
	if err != nil && !duplicate {
		return nil, err
	}
	if duplicate {
		profile, err := s.GetProfile(ctx, input.UserID)
		if err != nil {
			return nil, err
		}
		currentStreak := 0
		if profile.Streak != nil {
			currentStreak = profile.Streak.CurrentDays
		}
		return &domain.SessionRewardResult{TotalXP: profile.UserXP.TotalXP, NewLevel: profile.UserXP.Level, CurrentStreak: currentStreak}, nil
	}
	streak, err := s.UpdateStreak(ctx, input.UserID, input.CompletedAt)
	if err != nil {
		return nil, err
	}
	unlocked, err := s.unlockSessionAchievements(ctx, input, streak.CurrentDays)
	if err != nil {
		return nil, err
	}
	profile, err := s.GetProfile(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	return &domain.SessionRewardResult{XPGained: amount, TotalXP: profile.UserXP.TotalXP, NewLevel: profile.UserXP.Level, StreakUpdated: true, CurrentStreak: streak.CurrentDays, UnlockedAchievements: unlocked}, nil
}

func (s *Service) AddXP(ctx context.Context, userID domain.UserID, amount int, reason domain.XPReason, sourceType string, sourceID uuid.UUID) error {
	if amount <= 0 {
		return domain.ErrInvalidInput
	}
	now := time.Now().UTC()
	tx := domain.XPTransaction{ID: uuid.New(), UserID: userID, Amount: amount, Reason: reason, SourceType: sourceType, SourceID: sourceID, CreatedAt: now}
	if err := s.repo.CreateXPTransaction(ctx, tx); err != nil {
		return err
	}
	xp, err := s.repo.GetUserXP(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}
	if xp == nil {
		xp = &domain.UserXP{UserID: userID}
	}
	xp.TotalXP += amount
	xp.Level = xp.TotalXP/100 + 1
	xp.UpdatedAt = now
	return s.repo.SaveUserXP(ctx, *xp)
}

func (s *Service) UpdateStreak(ctx context.Context, userID domain.UserID, activityDate time.Time) (*domain.UserStreak, error) {
	now := time.Now().UTC()
	day := dateOnly(activityDate.UTC())
	streak, err := s.repo.GetUserStreak(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if streak == nil {
		streak = &domain.UserStreak{UserID: userID, CurrentDays: 1, LongestDays: 1, LastActivityDate: day, UpdatedAt: now}
		return streak, s.repo.SaveUserStreak(ctx, *streak)
	}
	last := dateOnly(streak.LastActivityDate.UTC())
	switch {
	case last.Equal(day):
	case last.AddDate(0, 0, 1).Equal(day):
		streak.CurrentDays++
	default:
		streak.CurrentDays = 1
	}
	if streak.CurrentDays > streak.LongestDays {
		streak.LongestDays = streak.CurrentDays
	}
	streak.LastActivityDate = day
	streak.UpdatedAt = now
	return streak, s.repo.SaveUserStreak(ctx, *streak)
}

func (s *Service) GetProfile(ctx context.Context, userID domain.UserID) (*domain.GamificationProfile, error) {
	xp, err := s.repo.GetUserXP(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if xp == nil {
		xp = &domain.UserXP{UserID: userID, Level: 1}
	}
	streak, err := s.repo.GetUserStreak(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	achievements, err := s.repo.ListUserAchievements(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &domain.GamificationProfile{UserXP: *xp, Streak: streak, Achievements: achievements}, nil
}

func (s *Service) unlockSessionAchievements(ctx context.Context, input domain.SessionRewardInput, currentStreak int) ([]domain.Achievement, error) {
	all, err := s.repo.ListAchievements(ctx)
	if err != nil {
		return nil, err
	}
	var unlocked []domain.Achievement
	for _, a := range all {
		if shouldUnlock(a.Code, input, currentStreak) {
			ok, err := s.repo.UnlockAchievement(ctx, input.UserID, a.ID, input.CompletedAt)
			if err != nil {
				return nil, err
			}
			if ok {
				unlocked = append(unlocked, a)
			}
		}
	}
	return unlocked, nil
}

func shouldUnlock(code string, input domain.SessionRewardInput, currentStreak int) bool {
	switch code {
	case "first_session":
		return true
	case "perfect_session":
		return input.TotalAnswers > 0 && input.CorrectAnswers == input.TotalAnswers
	case "streak_3":
		return currentStreak >= 3
	default:
		return false
	}
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

type LearningRecorder struct{ service *Service }

func NewLearningRecorder(service *Service) LearningRecorder {
	return LearningRecorder{service: service}
}

func (r LearningRecorder) RewardSessionCompleted(ctx context.Context, input learning_usecase.GamificationInput) error {
	_, err := r.service.RewardSessionCompleted(ctx, domain.SessionRewardInput{UserID: input.UserID, SessionID: uuid.UUID(input.SessionID), CorrectAnswers: input.CorrectAnswers, TotalAnswers: input.TotalAnswers, CompletedAt: input.CompletedAt})
	return err
}
