package api

import "context"

type CoursesUsecase interface {
	ListPublishedCourses(ctx context.Context) ([]Course, error)
	ListAllCourses(ctx context.Context) ([]Course, error)
	CreateCourse(ctx context.Context, input CourseWriteInput) (Course, error)
	UpdateCourse(ctx context.Context, input CourseWriteInput) (Course, error)
}

type SectionsUsecase interface {
	ListPublishedSections(ctx context.Context, input ListSectionsInput) ([]Section, error)
	ListAllSections(ctx context.Context, input ListSectionsInput) ([]Section, error)
	CreateSection(ctx context.Context, input SectionWriteInput) (Section, error)
	UpdateSection(ctx context.Context, input SectionWriteInput) (Section, error)
}

type UnitsUsecase interface {
	ListPublishedUnits(ctx context.Context, input ListUnitsInput) ([]Unit, error)
	ListAllUnits(ctx context.Context, input ListUnitsInput) ([]Unit, error)
	CreateUnit(ctx context.Context, input UnitWriteInput) (Unit, error)
	UpdateUnit(ctx context.Context, input UnitWriteInput) (Unit, error)
}

type SkillsUsecase interface {
	ListPublishedSkills(ctx context.Context, input ListSkillsInput) ([]Skill, error)
	ListAllSkills(ctx context.Context, input ListSkillsInput) ([]Skill, error)
	CreateSkill(ctx context.Context, input SkillWriteInput) (Skill, error)
	UpdateSkill(ctx context.Context, input SkillWriteInput) (Skill, error)
}

type ChallengesUsecase interface {
	ListPublishedChallenges(ctx context.Context, input ListChallengesInput) ([]Challenge, error)
	ListAllChallenges(ctx context.Context, input ListChallengesInput) ([]AuthoringChallenge, error)
	CreateChallenge(ctx context.Context, input ChallengeWriteInput) (AuthoringChallenge, error)
	UpdateChallenge(ctx context.Context, input ChallengeWriteInput) (AuthoringChallenge, error)
}
