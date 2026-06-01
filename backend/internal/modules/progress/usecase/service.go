package usecase

import (
	"context"

	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/SIniutin/history-app-backend/internal/modules/progress/domain"
)

const (
	maxSkillLevel         = 5
	levelUpAccuracy       = 0.7
	oldMasteryWeight      = 0.7
	sessionAccuracyWeight = 0.3
)

type Service struct {
	repo    domain.Repository
	content ContentRepository
}

type ContentRepository interface {
	ListPublishedCourses(ctx context.Context) ([]content_domain.Course, error)
	GetSkill(ctx context.Context, id content_domain.SkillID) (content_domain.Skill, error)
	GetUnit(ctx context.Context, id content_domain.UnitID) (content_domain.Unit, error)
	GetSection(ctx context.Context, id content_domain.SectionID) (content_domain.Section, error)
	ListPublishedSkills(ctx context.Context, unitID content_domain.UnitID) ([]content_domain.Skill, error)
	ListPublishedUnits(ctx context.Context, sectionID content_domain.SectionID) ([]content_domain.Unit, error)
	ListPublishedSections(ctx context.Context, courseID content_domain.CourseID) ([]content_domain.Section, error)
}

type Dependencies struct {
	Repository domain.Repository
	Content    ContentRepository
}

func NewService(deps Dependencies) *Service {
	return &Service{repo: deps.Repository, content: deps.Content}
}
