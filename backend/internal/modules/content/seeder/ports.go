package seeder

import (
	"context"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type ContentService interface {
	ListAllCourses(ctx context.Context) ([]content_domain.Course, error)
	CreateCourse(ctx context.Context, input content_api.CourseWriteInput) (content_domain.Course, error)
	CreateSection(ctx context.Context, input content_api.SectionWriteInput) (content_domain.Section, error)
	CreateUnit(ctx context.Context, input content_api.UnitWriteInput) (content_domain.Unit, error)
	CreateSkill(ctx context.Context, input content_api.SkillWriteInput) (content_domain.Skill, error)
	CreateChallenge(ctx context.Context, input content_api.ChallengeWriteInput) (content_domain.Challenge, error)
	Publish(ctx context.Context, input content_api.StatusTransitionInput) error
}

type UserFinder interface {
	GetByEmail(ctx context.Context, email users_domain.Email) (users_domain.User, error)
}
