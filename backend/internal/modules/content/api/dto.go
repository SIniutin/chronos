package api

import (
	"encoding/json"
)

type ListSectionsInput struct {
	CourseID string
}

type ListUnitsInput struct {
	SectionID string
}

type ListSkillsInput struct {
	UnitID string
}

type ListChallengesInput struct {
	SkillID string
}

type CourseWriteInput struct {
	ID         string `json:"-"`
	ActorID    string `json:"-"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Title      string `json:"title"`
}

type SectionWriteInput struct {
	ID          string `json:"-"`
	ActorID     string `json:"-"`
	CourseID    string `json:"course_id"`
	Theme       string `json:"theme"`
	Description string `json:"description"`
	Position    int    `json:"position"`
}

type UnitWriteInput struct {
	ID        string `json:"-"`
	ActorID   string `json:"-"`
	SectionID string `json:"section_id"`
	Title     string `json:"title"`
	Position  int    `json:"position"`
}

type SkillWriteInput struct {
	ID       string `json:"-"`
	ActorID  string `json:"-"`
	UnitID   string `json:"unit_id"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Position int    `json:"position"`
}

type ChallengeWriteInput struct {
	ID          string          `json:"-"`
	ActorID     string          `json:"-"`
	SkillID     string          `json:"skill_id"`
	Type        string          `json:"type"`
	Difficulty  string          `json:"difficulty"`
	Tags        json.RawMessage `json:"tags"`
	Level       int             `json:"level"`
	LessonCount int             `json:"lesson_count"`
	Prompt      string          `json:"prompt"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Options     json.RawMessage `json:"options"`
	Answers     json.RawMessage `json:"answers"`
	Explanation string          `json:"explanation"`
	Position    int             `json:"position"`
}

type StatusTransitionInput struct {
	Entity  string
	ID      string
	ActorID string
}

type Course struct {
	ID         string `json:"id"`
	SourceLang string `json:"source_lang"`
	TargetLang string `json:"target_lang"`
	Title      string `json:"title"`
	Status     string `json:"status"`
}

type Section struct {
	ID          string `json:"id"`
	CourseID    string `json:"course_id"`
	Theme       string `json:"theme"`
	Description string `json:"description"`
	Position    int    `json:"position"`
	Status      string `json:"status"`
}

type Unit struct {
	ID        string `json:"id"`
	SectionID string `json:"section_id"`
	Title     string `json:"title"`
	Position  int    `json:"position"`
	Status    string `json:"status"`
}

type Skill struct {
	ID       string `json:"id"`
	UnitID   string `json:"unit_id"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Position int    `json:"position"`
	Status   string `json:"status"`
}

type Challenge struct {
	ID          string          `json:"id"`
	SkillID     string          `json:"skill_id"`
	Type        string          `json:"type"`
	Difficulty  string          `json:"difficulty"`
	Tags        json.RawMessage `json:"tags"`
	Level       int             `json:"level"`
	LessonCount int             `json:"lesson_count"`
	Prompt      string          `json:"prompt"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Options     json.RawMessage `json:"options"`
	Explanation string          `json:"explanation"`
	Position    int             `json:"position"`
	Status      string          `json:"status"`
}

type AuthoringChallenge struct {
	ID          string          `json:"id"`
	SkillID     string          `json:"skill_id"`
	Type        string          `json:"type"`
	Difficulty  string          `json:"difficulty"`
	Tags        json.RawMessage `json:"tags"`
	Level       int             `json:"level"`
	LessonCount int             `json:"lesson_count"`
	Prompt      string          `json:"prompt"`
	Body        string          `json:"body"`
	Payload     json.RawMessage `json:"payload"`
	Options     json.RawMessage `json:"options"`
	Answers     json.RawMessage `json:"answers"`
	Explanation string          `json:"explanation"`
	Position    int             `json:"position"`
	Status      string          `json:"status"`
}
