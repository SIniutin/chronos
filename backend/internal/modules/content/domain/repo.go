package domain

import (
	"context"

	"github.com/google/uuid"
)

type CoursesRepository interface {
	ListPublishedCourses(ctx context.Context) ([]Course, error)
	ListAllCourses(ctx context.Context) ([]Course, error)
	CreateCourse(ctx context.Context, course Course) (Course, error)
	UpdateCourse(ctx context.Context, course Course) (Course, error)
}

type SectionsRepository interface {
	ListPublishedSections(ctx context.Context, courseID CourseID) ([]Section, error)
	ListAllSections(ctx context.Context, courseID CourseID) ([]Section, error)
	GetSection(ctx context.Context, id SectionID) (Section, error)
	CreateSection(ctx context.Context, section Section) (Section, error)
	UpdateSection(ctx context.Context, section Section) (Section, error)
}

type UnitsRepository interface {
	ListPublishedUnits(ctx context.Context, sectionID SectionID) ([]Unit, error)
	ListAllUnits(ctx context.Context, sectionID SectionID) ([]Unit, error)
	GetUnit(ctx context.Context, id UnitID) (Unit, error)
	CreateUnit(ctx context.Context, unit Unit) (Unit, error)
	UpdateUnit(ctx context.Context, unit Unit) (Unit, error)
}

type SkillsRepository interface {
	ListPublishedSkills(ctx context.Context, unitID UnitID) ([]Skill, error)
	ListAllSkills(ctx context.Context, unitID UnitID) ([]Skill, error)
	GetSkill(ctx context.Context, id SkillID) (Skill, error)
	CreateSkill(ctx context.Context, skill Skill) (Skill, error)
	UpdateSkill(ctx context.Context, skill Skill) (Skill, error)
}

type ChallengeRepository interface {
	ListPublishedChallenges(ctx context.Context, skillID SkillID) ([]Challenge, error)
	ListAllChallenges(ctx context.Context, skillID SkillID) ([]Challenge, error)
	GetChallenge(ctx context.Context, id ChallengeID) (Challenge, error)
	CreateChallenge(ctx context.Context, challenge Challenge) (Challenge, error)
	UpdateChallenge(ctx context.Context, challenge Challenge) (Challenge, error)
}

type StatusRepository interface {
	SetStatus(ctx context.Context, entity string, id uuid.UUID, status ContentStatus, actorID UserID) error
}
