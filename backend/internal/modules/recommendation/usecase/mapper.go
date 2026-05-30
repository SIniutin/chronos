package usecase

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/recommendation/api"
	"github.com/SIniutin/history-app-backend/internal/modules/recommendation/domain"
	"github.com/google/uuid"
)

type APIService struct{ service *Service }

func NewAPIService(service *Service) APIService { return APIService{service: service} }

func (s APIService) GetNextSkill(ctx context.Context, userIDRaw, courseIDRaw string) (*api.Recommendation, error) {
	userID, err := uuid.Parse(userIDRaw)
	if err != nil {
		return nil, err
	}
	courseID, err := uuid.Parse(courseIDRaw)
	if err != nil {
		return nil, err
	}
	rec, err := s.service.GetNextSkill(ctx, domain.UserID(userID), domain.CourseID(courseID))
	if err != nil {
		return nil, err
	}
	return toAPI(rec), nil
}

func toAPI(rec *domain.Recommendation) *api.Recommendation {
	out := &api.Recommendation{Type: string(rec.Type), CourseID: uuid.UUID(rec.CourseID).String(), Reason: rec.Reason}
	if rec.UnitID != nil {
		v := uuid.UUID(*rec.UnitID).String()
		out.UnitID = &v
	}
	if rec.SkillID != nil {
		v := uuid.UUID(*rec.SkillID).String()
		out.SkillID = &v
	}
	return out
}
