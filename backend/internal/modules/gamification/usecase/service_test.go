package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/SIniutin/history-app-backend/internal/modules/gamification/domain"
	"github.com/google/uuid"
)

func TestRewardSessionCompletedXPPerfectAndDuplicate(t *testing.T) {
	repo := newMemoryRepo()
	uc := NewService(repo)
	userID := domain.UserID(uuid.New())
	sessionID := uuid.New()

	result, err := uc.RewardSessionCompleted(context.Background(), domain.SessionRewardInput{
		UserID: userID, SessionID: sessionID, CorrectAnswers: 10, TotalAnswers: 10, CompletedAt: day(1),
	})
	if err != nil {
		t.Fatalf("reward failed: %v", err)
	}
	if result.XPGained != 35 || result.TotalXP != 35 || result.NewLevel != 1 {
		t.Fatalf("unexpected reward: %+v", result)
	}
	if len(result.UnlockedAchievements) != 2 {
		t.Fatalf("expected first + perfect achievements, got %+v", result.UnlockedAchievements)
	}

	duplicate, err := uc.RewardSessionCompleted(context.Background(), domain.SessionRewardInput{
		UserID: userID, SessionID: sessionID, CorrectAnswers: 10, TotalAnswers: 10, CompletedAt: day(1),
	})
	if err != nil {
		t.Fatalf("duplicate reward failed: %v", err)
	}
	if duplicate.XPGained != 0 || duplicate.TotalXP != 35 {
		t.Fatalf("duplicate should not double XP: %+v", duplicate)
	}
}

func TestUpdateStreakSameYesterdayReset(t *testing.T) {
	repo := newMemoryRepo()
	uc := NewService(repo)
	userID := domain.UserID(uuid.New())

	if streak, _ := uc.UpdateStreak(context.Background(), userID, day(1)); streak.CurrentDays != 1 {
		t.Fatalf("expected first streak day, got %+v", streak)
	}
	if streak, _ := uc.UpdateStreak(context.Background(), userID, day(1)); streak.CurrentDays != 1 {
		t.Fatalf("same day should not increment, got %+v", streak)
	}
	if streak, _ := uc.UpdateStreak(context.Background(), userID, day(2)); streak.CurrentDays != 2 {
		t.Fatalf("yesterday should increment, got %+v", streak)
	}
	if streak, _ := uc.UpdateStreak(context.Background(), userID, day(5)); streak.CurrentDays != 1 || streak.LongestDays != 2 {
		t.Fatalf("gap should reset current only, got %+v", streak)
	}
}

type memoryRepo struct {
	xp           map[domain.UserID]domain.UserXP
	tx           map[string]domain.XPTransaction
	streaks      map[domain.UserID]domain.UserStreak
	achievements []domain.Achievement
	userAch      map[string]struct{}
}

func newMemoryRepo() *memoryRepo {
	return &memoryRepo{
		xp:      map[domain.UserID]domain.UserXP{},
		tx:      map[string]domain.XPTransaction{},
		streaks: map[domain.UserID]domain.UserStreak{},
		achievements: []domain.Achievement{
			{ID: uuid.New(), Code: "first_session", Title: "First"},
			{ID: uuid.New(), Code: "perfect_session", Title: "Perfect"},
			{ID: uuid.New(), Code: "streak_3", Title: "Streak"},
		},
		userAch: map[string]struct{}{},
	}
}

func (r *memoryRepo) GetUserXP(_ context.Context, userID domain.UserID) (*domain.UserXP, error) {
	xp, ok := r.xp[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &xp, nil
}

func (r *memoryRepo) SaveUserXP(_ context.Context, xp domain.UserXP) error {
	r.xp[xp.UserID] = xp
	return nil
}

func (r *memoryRepo) CreateXPTransaction(_ context.Context, tx domain.XPTransaction) error {
	key := uuid.UUID(tx.UserID).String() + ":" + string(tx.Reason) + ":" + tx.SourceType + ":" + tx.SourceID.String()
	if _, ok := r.tx[key]; ok {
		return domain.ErrDuplicateXP
	}
	r.tx[key] = tx
	return nil
}

func (r *memoryRepo) GetUserStreak(_ context.Context, userID domain.UserID) (*domain.UserStreak, error) {
	streak, ok := r.streaks[userID]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &streak, nil
}

func (r *memoryRepo) SaveUserStreak(_ context.Context, streak domain.UserStreak) error {
	r.streaks[streak.UserID] = streak
	return nil
}
func (r *memoryRepo) ListAchievements(context.Context) ([]domain.Achievement, error) {
	return r.achievements, nil
}

func (r *memoryRepo) ListUserAchievements(_ context.Context, userID domain.UserID) ([]domain.Achievement, error) {
	var out []domain.Achievement
	for _, a := range r.achievements {
		if _, ok := r.userAch[uuid.UUID(userID).String()+":"+a.ID.String()]; ok {
			out = append(out, a)
		}
	}
	return out, nil
}

func (r *memoryRepo) UnlockAchievement(_ context.Context, userID domain.UserID, achievementID uuid.UUID, _ time.Time) (bool, error) {
	key := uuid.UUID(userID).String() + ":" + achievementID.String()
	if _, ok := r.userAch[key]; ok {
		return false, nil
	}
	for _, a := range r.achievements {
		if a.ID == achievementID {
			r.userAch[key] = struct{}{}
			return true, nil
		}
	}
	return false, errors.New("missing achievement")
}

func day(n int) time.Time { return time.Date(2026, 5, n, 12, 0, 0, 0, time.UTC) }
