package usecase

import (
	"context"
	"time"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/learning/domain"
)

type Service struct {
	sessions     domain.SessionRepository
	queue        domain.SessionChallengeRepository
	attempts     domain.AttemptRepository
	content      ContentRepository
	picker       ChallengePicker
	progress     ProgressRecorder
	gamification GamificationRecorder
}

type ContentRepository interface {
	ListPublishedChallenges(ctx context.Context, skillID content_domain.SkillID) ([]content_domain.Challenge, error)
	GetChallenge(ctx context.Context, id content_domain.ChallengeID) (content_domain.Challenge, error)
}

type ProgressRecorder interface {
	RecordSessionResult(ctx context.Context, input ProgressInput) error
}

type GamificationRecorder interface {
	RewardSessionCompleted(ctx context.Context, input GamificationInput) error
}

type ChallengePicker interface {
	PickChallengesForSession(ctx context.Context, userID domain.UserID, skillID domain.SkillID, limit int) ([]domain.ChallengeID, error)
}

type Dependencies struct {
	Sessions     domain.SessionRepository
	Queue        domain.SessionChallengeRepository
	Attempts     domain.AttemptRepository
	Content      ContentRepository
	Picker       ChallengePicker
	Progress     ProgressRecorder
	Gamification GamificationRecorder
}

func NewService(deps Dependencies) *Service {
	progress := deps.Progress
	if progress == nil {
		progress = NoopProgressRecorder{}
	}
	gamification := deps.Gamification
	if gamification == nil {
		gamification = NoopGamificationRecorder{}
	}
	return &Service{
		sessions:     deps.Sessions,
		queue:        deps.Queue,
		attempts:     deps.Attempts,
		content:      deps.Content,
		picker:       deps.Picker,
		progress:     progress,
		gamification: gamification,
	}
}

func NewServiceFromRepository(repo interface {
	domain.SessionRepository
	domain.SessionChallengeRepository
	domain.AttemptRepository
}, content ContentRepository) *Service {
	return NewService(Dependencies{
		Sessions: repo,
		Queue:    repo,
		Attempts: repo,
		Content:  content,
	})
}

type NoopProgressRecorder struct{}

type ProgressInput struct {
	UserID         domain.UserID
	SkillID        domain.SkillID
	CorrectAnswers int
	TotalAnswers   int
	CompletedAt    time.Time
}

type GamificationInput struct {
	UserID         domain.UserID
	SessionID      domain.LessonSessionID
	CorrectAnswers int
	TotalAnswers   int
	CompletedAt    time.Time
}

func (NoopProgressRecorder) RecordSessionResult(context.Context, ProgressInput) error {
	return nil
}

type NoopGamificationRecorder struct{}

func (NoopGamificationRecorder) RewardSessionCompleted(context.Context, GamificationInput) error {
	return nil
}
