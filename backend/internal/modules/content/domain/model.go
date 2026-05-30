package domain

import (
	"encoding/json"
	"errors"
	"time"

	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	"github.com/google/uuid"
)

type UserID = users_domain.UserID

type CourseID uuid.UUID
type SectionID uuid.UUID
type UnitID uuid.UUID
type SkillID uuid.UUID
type ChallengeID uuid.UUID
type SectionDescriptionID uuid.UUID

type Audit struct {
	CreatedBy  UserID
	UpdatedBy  UserID
	ReviewedBy *UserID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ReviewedAt *time.Time
}

type Course struct {
	ID         CourseID
	SourceLang string
	TargetLang string
	Title      string
	Status     ContentStatus
	Audit      Audit
}

type Section struct {
	ID          SectionID
	CourseID    CourseID
	Theme       string
	Description string
	Position    int
	Status      ContentStatus
	Audit       Audit
}

type SectionDescription struct {
	ID        SectionDescriptionID
	SectionID SectionID
	Title     string
	Text      string
	Position  int
	Audit     Audit
}

type Unit struct {
	ID        UnitID
	SectionID SectionID
	Title     string
	Position  int
	Status    ContentStatus
	Audit     Audit
}

type Skill struct {
	ID       SkillID
	UnitID   UnitID
	Title    string
	Icon     string
	Position int
	Status   ContentStatus
	Audit    Audit
}

type Challenge struct {
	ID          ChallengeID
	SkillID     SkillID
	Type        ChallengeType
	Difficulty  Difficulty
	Tags        json.RawMessage
	Level       int
	LessonCount int
	Prompt      string
	Body        string
	Payload     json.RawMessage
	Options     json.RawMessage
	Answers     json.RawMessage
	Explanation string
	Position    int
	Status      ContentStatus
	Audit       Audit
}

func ValidateTags(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var tags []string
	if err := json.Unmarshal(raw, &tags); err != nil {
		return errors.Join(ErrInvalidInput, err)
	}
	return nil
}

func ParseCourseID(raw string) (CourseID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return CourseID{}, errors.Join(ErrInvalidInput, err)
	}
	return CourseID(id), nil
}

func ParseSectionID(raw string) (SectionID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return SectionID{}, errors.Join(ErrInvalidInput, err)
	}
	return SectionID(id), nil
}

func ParseUnitID(raw string) (UnitID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return UnitID{}, errors.Join(ErrInvalidInput, err)
	}
	return UnitID(id), nil
}

func ParseSkillID(raw string) (SkillID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return SkillID{}, errors.Join(ErrInvalidInput, err)
	}
	return SkillID(id), nil
}
