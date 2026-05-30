package usecase

import (
	"errors"

	content_api "github.com/SIniutin/history-app-backend/internal/modules/content/api"
	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

func mapDomainError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, cd.ErrInvalidInput):
		return errors.Join(content_api.ErrInvalidInput, err)
	case errors.Is(err, cd.ErrNotFound):
		return content_api.ErrNotFound
	default:
		return err
	}
}

func ToAPICourse(c cd.Course) content_api.Course {
	return content_api.Course{
		ID:         uuid.UUID(c.ID).String(),
		SourceLang: c.SourceLang,
		TargetLang: c.TargetLang,
		Title:      c.Title,
		Status:     string(c.Status),
	}
}

func ToAPISection(s cd.Section) content_api.Section {
	return content_api.Section{
		ID:          uuid.UUID(s.ID).String(),
		CourseID:    uuid.UUID(s.CourseID).String(),
		Theme:       s.Theme,
		Description: s.Description,
		Position:    s.Position,
		Status:      string(s.Status),
	}
}

func ToAPIUnit(u cd.Unit) content_api.Unit {
	return content_api.Unit{
		ID:        uuid.UUID(u.ID).String(),
		SectionID: uuid.UUID(u.SectionID).String(),
		Title:     u.Title,
		Position:  u.Position,
		Status:    string(u.Status),
	}
}

func ToAPISkill(s cd.Skill) content_api.Skill {
	return content_api.Skill{
		ID:       uuid.UUID(s.ID).String(),
		UnitID:   uuid.UUID(s.UnitID).String(),
		Title:    s.Title,
		Icon:     s.Icon,
		Position: s.Position,
		Status:   string(s.Status),
	}
}

func ToAPIChallenge(c cd.Challenge) content_api.Challenge {
	return content_api.Challenge{
		ID:          uuid.UUID(c.ID).String(),
		SkillID:     uuid.UUID(c.SkillID).String(),
		Type:        string(c.Type),
		Difficulty:  string(c.Difficulty),
		Tags:        c.Tags,
		Level:       c.Level,
		LessonCount: c.LessonCount,
		Prompt:      c.Prompt,
		Body:        c.Body,
		Payload:     c.Payload,
		Options:     c.Options,
		Explanation: c.Explanation,
		Position:    c.Position,
		Status:      string(c.Status),
	}
}

func ToAPIAuthoringChallenge(c cd.Challenge) content_api.AuthoringChallenge {
	return content_api.AuthoringChallenge{
		ID:          uuid.UUID(c.ID).String(),
		SkillID:     uuid.UUID(c.SkillID).String(),
		Type:        string(c.Type),
		Difficulty:  string(c.Difficulty),
		Tags:        c.Tags,
		Level:       c.Level,
		LessonCount: c.LessonCount,
		Prompt:      c.Prompt,
		Body:        c.Body,
		Payload:     c.Payload,
		Options:     c.Options,
		Answers:     c.Answers,
		Explanation: c.Explanation,
		Position:    c.Position,
		Status:      string(c.Status),
	}
}
