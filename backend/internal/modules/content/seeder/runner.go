package seeder

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

type Runner struct {
	content ContentService
}

type Result struct {
	Skipped    bool
	Courses    int
	Sections   int
	Units      int
	Skills     int
	Challenges int
}

func NewRunner(content ContentService) *Runner {
	return &Runner{content: content}
}

func (r *Runner) Run(ctx context.Context, seed SeedFile, actorID string) (Result, error) {
	if err := validateSeed(seed); err != nil {
		return Result{}, err
	}
	exists, err := r.courseExists(ctx, seed.Course.Title)
	if err != nil {
		return Result{}, err
	}
	if exists {
		return Result{Skipped: true}, nil
	}

	course, err := r.content.CreateCourse(ctx, content_api.CourseWriteInput{
		ActorID:    actorID,
		SourceLang: seed.Course.SourceLang,
		TargetLang: seed.Course.TargetLang,
		Title:      seed.Course.Title,
	})
	if err != nil {
		return Result{}, fmt.Errorf("create course: %w", err)
	}
	result := Result{Courses: 1}
	if err := r.publish(ctx, actorID, "courses", uuid.UUID(course.ID), content_domain.ContentStatusPublished); err != nil {
		return Result{}, err
	}

	for sectionIndex, seedSection := range seed.Sections {
		section, err := r.content.CreateSection(ctx, content_api.SectionWriteInput{
			ActorID:     actorID,
			CourseID:    uuid.UUID(course.ID).String(),
			Theme:       seedSection.Theme,
			Description: seedSection.Description,
			Position:    sectionIndex + 1,
		})
		if err != nil {
			return Result{}, fmt.Errorf("create section %q: %w", seedSection.Theme, err)
		}
		result.Sections++
		if err := r.publish(ctx, actorID, "sections", uuid.UUID(section.ID), content_domain.ContentStatusPublished); err != nil {
			return Result{}, err
		}

		for unitIndex, seedUnit := range seedSection.Units {
			unit, err := r.content.CreateUnit(ctx, content_api.UnitWriteInput{
				ActorID:   actorID,
				SectionID: uuid.UUID(section.ID).String(),
				Title:     seedUnit.Title,
				Position:  unitIndex + 1,
			})
			if err != nil {
				return Result{}, fmt.Errorf("create unit %q: %w", seedUnit.Title, err)
			}
			result.Units++
			if err := r.publish(ctx, actorID, "units", uuid.UUID(unit.ID), content_domain.ContentStatusPublished); err != nil {
				return Result{}, err
			}

			for skillIndex, seedSkill := range seedUnit.Skills {
				skill, err := r.content.CreateSkill(ctx, content_api.SkillWriteInput{
					ActorID:  actorID,
					UnitID:   uuid.UUID(unit.ID).String(),
					Title:    seedSkill.Title,
					Icon:     seedSkill.Icon,
					Position: skillIndex + 1,
				})
				if err != nil {
					return Result{}, fmt.Errorf("create skill %q: %w", seedSkill.Title, err)
				}
				result.Skills++
				if err := r.publish(ctx, actorID, "skills", uuid.UUID(skill.ID), content_domain.ContentStatusPublished); err != nil {
					return Result{}, err
				}

				for challengeIndex, seedChallenge := range seedSkill.Challenges {
					position := seedChallenge.Position
					if position <= 0 {
						position = challengeIndex + 1
					}
					challenge, err := r.content.CreateChallenge(ctx, content_api.ChallengeWriteInput{
						ActorID:     actorID,
						SkillID:     uuid.UUID(skill.ID).String(),
						Type:        defaultString(seedChallenge.Type, "theory"),
						Difficulty:  defaultString(seedChallenge.Difficulty, "easy"),
						Tags:        jsonDefault(seedChallenge.Tags, "[]"),
						Level:       defaultInt(seedChallenge.Level, 1),
						LessonCount: defaultInt(seedChallenge.LessonCount, 1),
						Prompt:      seedChallenge.Prompt,
						Body:        seedChallenge.Body,
						Payload:     jsonDefault(seedChallenge.Payload, "{}"),
						Options:     jsonDefault(seedChallenge.Options, "[]"),
						Answers:     jsonDefault(seedChallenge.Answers, "[]"),
						Explanation: seedChallenge.Explanation,
						Position:    position,
					})
					if err != nil {
						return Result{}, fmt.Errorf("create challenge %q: %w", seedChallenge.Prompt, err)
					}
					result.Challenges++
					if err := r.publish(ctx, actorID, "challenges", uuid.UUID(challenge.ID), content_domain.ContentStatus(seedChallenge.Status)); err != nil {
						return Result{}, err
					}
				}
			}
		}
	}
	return result, nil
}

func (r *Runner) courseExists(ctx context.Context, title string) (bool, error) {
	courses, err := r.content.ListAllCourses(ctx)
	if err != nil {
		return false, fmt.Errorf("list courses: %w", err)
	}
	for _, course := range courses {
		if strings.EqualFold(strings.TrimSpace(course.Title), strings.TrimSpace(title)) {
			return true, nil
		}
	}
	return false, nil
}

func (r *Runner) publish(ctx context.Context, actorID, entity string, id uuid.UUID, status content_domain.ContentStatus) error {
	if status != content_domain.ContentStatusPublished {
		return nil
	}
	if err := r.content.Publish(ctx, content_api.StatusTransitionInput{
		Entity:  entity,
		ID:      id.String(),
		ActorID: actorID,
	}); err != nil {
		return fmt.Errorf("publish %s %s: %w", entity, id, err)
	}
	return nil
}

func validateSeed(seed SeedFile) error {
	if strings.TrimSpace(seed.Course.Title) == "" {
		return fmt.Errorf("seed course title is empty")
	}
	if strings.TrimSpace(seed.Course.SourceLang) == "" || strings.TrimSpace(seed.Course.TargetLang) == "" {
		return fmt.Errorf("seed course languages are empty")
	}
	return nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}

func jsonDefault(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 {
		return json.RawMessage(fallback)
	}
	return raw
}
