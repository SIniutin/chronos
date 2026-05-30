package usecase

import (
	"context"
	"testing"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

var _ CoursesUsecase = (*Service)(nil)
var _ SectionsUsecase = (*Service)(nil)
var _ UnitsUsecase = (*Service)(nil)
var _ SkillsUsecase = (*Service)(nil)
var _ ChallengesUsecase = (*Service)(nil)

func TestListCoursesFiltersPublished(t *testing.T) {
	repo := &memoryRepo{
		courses: []cd.Course{
			{ID: cd.CourseID(uuid.New()), Title: "Published", Status: cd.ContentStatusPublished},
			{ID: cd.CourseID(uuid.New()), Title: "Draft", Status: cd.ContentStatusDraft},
		},
	}
	uc := NewServiceFromRepository(repo)

	courses, err := uc.ListPublishedCourses(context.Background())
	if err != nil {
		t.Fatalf("list courses failed: %v", err)
	}
	if len(courses) != 1 || courses[0].Title != "Published" {
		t.Fatalf("expected one published course, got %+v", courses)
	}
}

func TestListChallengesDoesNotExposeAnswers(t *testing.T) {
	skillID := uuid.New()
	repo := &memoryRepo{
		challenges: []cd.Challenge{
			{
				ID:          cd.ChallengeID(uuid.New()),
				SkillID:     cd.SkillID(skillID),
				Type:        cd.ChallengeTypeSingleChoice,
				Difficulty:  cd.DifficultyEasy,
				Tags:        []byte(`["seed"]`),
				Prompt:      "Question?",
				Options:     []byte(`["a","b"]`),
				Answers:     []byte(`["a"]`),
				Explanation: "Because",
				Status:      cd.ContentStatusPublished,
			},
		},
	}
	uc := NewServiceFromRepository(repo)

	challenges, err := uc.ListPublishedChallenges(context.Background(), content_api.ListChallengesInput{SkillID: skillID.String()})
	if err != nil {
		t.Fatalf("list challenges failed: %v", err)
	}
	if len(challenges) != 1 {
		t.Fatalf("expected one challenge, got %+v", challenges)
	}
	if string(challenges[0].Options) == "" || challenges[0].Explanation == "" {
		t.Fatalf("expected public challenge data, got %+v", challenges[0])
	}
}

type memoryRepo struct {
	courses    []cd.Course
	sections   []cd.Section
	units      []cd.Unit
	skills     []cd.Skill
	challenges []cd.Challenge
}

func (r *memoryRepo) ListPublishedCourses(context.Context) ([]cd.Course, error) {
	var out []cd.Course
	for _, course := range r.courses {
		if course.Status.IsPublished() {
			out = append(out, course)
		}
	}
	return out, nil
}

func (r *memoryRepo) ListAllCourses(context.Context) ([]cd.Course, error) {
	return r.courses, nil
}

func (r *memoryRepo) CreateCourse(context.Context, cd.Course) (cd.Course, error) {
	return cd.Course{}, nil
}

func (r *memoryRepo) UpdateCourse(context.Context, cd.Course) (cd.Course, error) {
	return cd.Course{}, nil
}

func (r *memoryRepo) ListPublishedSections(context.Context, cd.CourseID) ([]cd.Section, error) {
	var out []cd.Section
	for _, section := range r.sections {
		if section.Status.IsPublished() {
			out = append(out, section)
		}
	}
	return out, nil
}

func (r *memoryRepo) ListAllSections(context.Context, cd.CourseID) ([]cd.Section, error) {
	return r.sections, nil
}

func (r *memoryRepo) GetSection(_ context.Context, id cd.SectionID) (cd.Section, error) {
	for _, section := range r.sections {
		if section.ID == id {
			return section, nil
		}
	}
	return cd.Section{}, cd.ErrNotFound
}

func (r *memoryRepo) CreateSection(context.Context, cd.Section) (cd.Section, error) {
	return cd.Section{}, nil
}

func (r *memoryRepo) UpdateSection(context.Context, cd.Section) (cd.Section, error) {
	return cd.Section{}, nil
}

func (r *memoryRepo) ListPublishedUnits(context.Context, cd.SectionID) ([]cd.Unit, error) {
	var out []cd.Unit
	for _, unit := range r.units {
		if unit.Status.IsPublished() {
			out = append(out, unit)
		}
	}
	return out, nil
}

func (r *memoryRepo) ListAllUnits(context.Context, cd.SectionID) ([]cd.Unit, error) {
	return r.units, nil
}

func (r *memoryRepo) GetUnit(_ context.Context, id cd.UnitID) (cd.Unit, error) {
	for _, unit := range r.units {
		if unit.ID == id {
			return unit, nil
		}
	}
	return cd.Unit{}, cd.ErrNotFound
}

func (r *memoryRepo) CreateUnit(context.Context, cd.Unit) (cd.Unit, error) {
	return cd.Unit{}, nil
}

func (r *memoryRepo) UpdateUnit(context.Context, cd.Unit) (cd.Unit, error) {
	return cd.Unit{}, nil
}

func (r *memoryRepo) ListPublishedSkills(context.Context, cd.UnitID) ([]cd.Skill, error) {
	var out []cd.Skill
	for _, skill := range r.skills {
		if skill.Status.IsPublished() {
			out = append(out, skill)
		}
	}
	return out, nil
}

func (r *memoryRepo) ListAllSkills(context.Context, cd.UnitID) ([]cd.Skill, error) {
	return r.skills, nil
}

func (r *memoryRepo) GetSkill(_ context.Context, id cd.SkillID) (cd.Skill, error) {
	for _, skill := range r.skills {
		if skill.ID == id {
			return skill, nil
		}
	}
	return cd.Skill{}, cd.ErrNotFound
}

func (r *memoryRepo) CreateSkill(context.Context, cd.Skill) (cd.Skill, error) {
	return cd.Skill{}, nil
}

func (r *memoryRepo) UpdateSkill(context.Context, cd.Skill) (cd.Skill, error) {
	return cd.Skill{}, nil
}

func (r *memoryRepo) ListPublishedChallenges(context.Context, cd.SkillID) ([]cd.Challenge, error) {
	var out []cd.Challenge
	for _, challenge := range r.challenges {
		if challenge.Status.IsPublished() {
			out = append(out, challenge)
		}
	}
	return out, nil
}

func (r *memoryRepo) ListAllChallenges(context.Context, cd.SkillID) ([]cd.Challenge, error) {
	return r.challenges, nil
}

func (r *memoryRepo) GetChallenge(_ context.Context, id cd.ChallengeID) (cd.Challenge, error) {
	for _, challenge := range r.challenges {
		if challenge.ID == id {
			return challenge, nil
		}
	}
	return cd.Challenge{}, cd.ErrNotFound
}

func (r *memoryRepo) CreateChallenge(context.Context, cd.Challenge) (cd.Challenge, error) {
	return cd.Challenge{}, nil
}

func (r *memoryRepo) UpdateChallenge(context.Context, cd.Challenge) (cd.Challenge, error) {
	return cd.Challenge{}, nil
}

func (r *memoryRepo) SetStatus(context.Context, string, uuid.UUID, cd.ContentStatus, cd.UserID) error {
	return nil
}
