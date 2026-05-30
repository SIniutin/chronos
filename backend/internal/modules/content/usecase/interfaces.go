package usecase

import (
	"context"

	"github.com/SIniutin/history-app-backend/internal/modules/content/api"
	"github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

type CoursesUsecase interface {
	ListPublishedCourses(ctx context.Context) ([]domain.Course, error)
	ListAllCourses(ctx context.Context) ([]domain.Course, error)
	CreateCourse(ctx context.Context, input api.CourseWriteInput) (domain.Course, error)
	UpdateCourse(ctx context.Context, input api.CourseWriteInput) (domain.Course, error)
	Publish(ctx context.Context, input api.StatusTransitionInput) error
	Archive(ctx context.Context, input api.StatusTransitionInput) error
}

type SectionsUsecase interface {
	ListPublishedSections(ctx context.Context, input api.ListSectionsInput) ([]domain.Section, error)
	ListAllSections(ctx context.Context, input api.ListSectionsInput) ([]domain.Section, error)
	CreateSection(ctx context.Context, input api.SectionWriteInput) (domain.Section, error)
	UpdateSection(ctx context.Context, input api.SectionWriteInput) (domain.Section, error)
	Publish(ctx context.Context, input api.StatusTransitionInput) error
	Archive(ctx context.Context, input api.StatusTransitionInput) error
}

type UnitsUsecase interface {
	ListPublishedUnits(ctx context.Context, input api.ListUnitsInput) ([]domain.Unit, error)
	ListAllUnits(ctx context.Context, input api.ListUnitsInput) ([]domain.Unit, error)
	CreateUnit(ctx context.Context, input api.UnitWriteInput) (domain.Unit, error)
	UpdateUnit(ctx context.Context, input api.UnitWriteInput) (domain.Unit, error)
	Publish(ctx context.Context, input api.StatusTransitionInput) error
	Archive(ctx context.Context, input api.StatusTransitionInput) error
}

type SkillsUsecase interface {
	ListPublishedSkills(ctx context.Context, input api.ListSkillsInput) ([]domain.Skill, error)
	ListAllSkills(ctx context.Context, input api.ListSkillsInput) ([]domain.Skill, error)
	CreateSkill(ctx context.Context, input api.SkillWriteInput) (domain.Skill, error)
	UpdateSkill(ctx context.Context, input api.SkillWriteInput) (domain.Skill, error)
	Publish(ctx context.Context, input api.StatusTransitionInput) error
	Archive(ctx context.Context, input api.StatusTransitionInput) error
}

type ChallengesUsecase interface {
	ListPublishedChallenges(ctx context.Context, input api.ListChallengesInput) ([]domain.Challenge, error)
	ListAllChallenges(ctx context.Context, input api.ListChallengesInput) ([]domain.Challenge, error)
	CreateChallenge(ctx context.Context, input api.ChallengeWriteInput) (domain.Challenge, error)
	UpdateChallenge(ctx context.Context, input api.ChallengeWriteInput) (domain.Challenge, error)
	Publish(ctx context.Context, input api.StatusTransitionInput) error
	Archive(ctx context.Context, input api.StatusTransitionInput) error
}

var _ CoursesUsecase = (*Service)(nil)
var _ SectionsUsecase = (*Service)(nil)
var _ UnitsUsecase = (*Service)(nil)
var _ SkillsUsecase = (*Service)(nil)
var _ ChallengesUsecase = (*Service)(nil)
