package usecase

import (
	"context"

	learning_usecase "github.com/SIniutin/history-app-backend/internal/modules/learning/usecase"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
)

type LearningRecorder struct {
	service *Service
}

func NewLearningRecorder(service *Service) LearningRecorder {
	return LearningRecorder{service: service}
}

func (r LearningRecorder) RecordSessionResult(ctx context.Context, input learning_usecase.ProgressInput) error {
	_, err := r.service.ApplySessionResult(ctx, domain.SessionProgressInput{
		UserID:         input.UserID,
		SkillID:        input.SkillID,
		CorrectAnswers: input.CorrectAnswers,
		TotalAnswers:   input.TotalAnswers,
		CompletedAt:    input.CompletedAt,
	})
	return err
}
