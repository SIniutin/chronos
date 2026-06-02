package seeder

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

func TestRunnerCreatesStructuredHistorySeed(t *testing.T) {
	seed, err := LoadFile(filepath.Join("..", "..", "..", "..", "seeds", "history_course_structured.json"))
	if err != nil {
		t.Fatalf("load seed failed: %v", err)
	}
	content := newFakeContentService()

	result, err := NewRunner(content).Run(context.Background(), seed, uuid.NewString())
	if err != nil {
		t.Fatalf("run seed failed: %v", err)
	}
	if result.Skipped {
		t.Fatalf("seed should not be skipped")
	}
	if result.Courses != 1 || result.Sections != 47 || result.Units != 60 || result.Skills != 67 || result.Challenges != 535 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if content.createdChallenges != 535 {
		t.Fatalf("expected 535 created challenges, got %d", content.createdChallenges)
	}
	if content.published["challenges"] != 529 {
		t.Fatalf("expected 529 published challenges, got %d", content.published["challenges"])
	}
}

func TestRunnerSkipsExistingCourse(t *testing.T) {
	seed := SeedFile{
		Course: SeedCourse{Title: "История России начала XX века", SourceLang: "ru", TargetLang: "ru"},
	}
	content := newFakeContentService()
	content.courses = append(content.courses, content_domain.Course{
		ID:    content_domain.CourseID(uuid.New()),
		Title: "История России начала XX века",
	})

	result, err := NewRunner(content).Run(context.Background(), seed, uuid.NewString())
	if err != nil {
		t.Fatalf("run seed failed: %v", err)
	}
	if !result.Skipped {
		t.Fatalf("expected skip result, got %+v", result)
	}
	if content.createdCourses != 0 {
		t.Fatalf("expected no created courses, got %d", content.createdCourses)
	}
}

func TestResolveActorReturnsClearMissingActorError(t *testing.T) {
	_, err := ResolveActor(context.Background(), fakeUsers{}, "admin@example.com")
	if err == nil {
		t.Fatalf("expected missing actor error")
	}
	if !errors.Is(err, users_domain.ErrUserNotFound) {
		t.Fatalf("expected wrapped user not found error, got %v", err)
	}
}

type fakeContentService struct {
	courses           []content_domain.Course
	createdCourses    int
	createdSections   int
	createdUnits      int
	createdSkills     int
	createdChallenges int
	published         map[string]int
}

func newFakeContentService() *fakeContentService {
	return &fakeContentService{published: map[string]int{}}
}

func (s *fakeContentService) ListAllCourses(context.Context) ([]content_domain.Course, error) {
	return s.courses, nil
}

func (s *fakeContentService) CreateCourse(_ context.Context, input content_api.CourseWriteInput) (content_domain.Course, error) {
	s.createdCourses++
	course := content_domain.Course{
		ID:         content_domain.CourseID(uuid.New()),
		SourceLang: input.SourceLang,
		TargetLang: input.TargetLang,
		Title:      input.Title,
	}
	s.courses = append(s.courses, course)
	return course, nil
}

func (s *fakeContentService) CreateSection(_ context.Context, input content_api.SectionWriteInput) (content_domain.Section, error) {
	s.createdSections++
	id, _ := uuid.Parse(input.CourseID)
	return content_domain.Section{
		ID:          content_domain.SectionID(uuid.New()),
		CourseID:    content_domain.CourseID(id),
		Theme:       input.Theme,
		Description: input.Description,
		Position:    input.Position,
	}, nil
}

func (s *fakeContentService) CreateUnit(_ context.Context, input content_api.UnitWriteInput) (content_domain.Unit, error) {
	s.createdUnits++
	id, _ := uuid.Parse(input.SectionID)
	return content_domain.Unit{
		ID:        content_domain.UnitID(uuid.New()),
		SectionID: content_domain.SectionID(id),
		Title:     input.Title,
		Position:  input.Position,
	}, nil
}

func (s *fakeContentService) CreateSkill(_ context.Context, input content_api.SkillWriteInput) (content_domain.Skill, error) {
	s.createdSkills++
	id, _ := uuid.Parse(input.UnitID)
	return content_domain.Skill{
		ID:       content_domain.SkillID(uuid.New()),
		UnitID:   content_domain.UnitID(id),
		Title:    input.Title,
		Icon:     input.Icon,
		Position: input.Position,
	}, nil
}

func (s *fakeContentService) CreateChallenge(_ context.Context, input content_api.ChallengeWriteInput) (content_domain.Challenge, error) {
	s.createdChallenges++
	id, _ := uuid.Parse(input.SkillID)
	if len(input.Tags) == 0 || len(input.Payload) == 0 || len(input.Options) == 0 || len(input.Answers) == 0 {
		return content_domain.Challenge{}, errors.New("challenge JSON defaults must be present")
	}
	if !json.Valid(input.Tags) || !json.Valid(input.Payload) || !json.Valid(input.Options) || !json.Valid(input.Answers) {
		return content_domain.Challenge{}, errors.New("challenge JSON must be valid")
	}
	return content_domain.Challenge{
		ID:          content_domain.ChallengeID(uuid.New()),
		SkillID:     content_domain.SkillID(id),
		Type:        content_domain.ChallengeType(input.Type),
		Difficulty:  content_domain.Difficulty(input.Difficulty),
		Tags:        input.Tags,
		Level:       input.Level,
		LessonCount: input.LessonCount,
		Prompt:      input.Prompt,
		Body:        input.Body,
		Payload:     input.Payload,
		Options:     input.Options,
		Answers:     input.Answers,
		Explanation: input.Explanation,
		Position:    input.Position,
	}, nil
}

func (s *fakeContentService) Publish(_ context.Context, input content_api.StatusTransitionInput) error {
	s.published[input.Entity]++
	return nil
}

type fakeUsers struct{}

func (fakeUsers) GetByEmail(context.Context, users_domain.Email) (users_domain.User, error) {
	return users_domain.User{}, users_domain.ErrUserNotFound
}
