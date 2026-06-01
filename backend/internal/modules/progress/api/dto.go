package api

import "context"

type CatalogProgress struct {
	CourseID         string          `json:"course_id"`
	CourseStatus     string          `json:"course_status"`
	TotalLessons     int             `json:"total_lessons"`
	AvailableLessons int             `json:"available_lessons"`
	CompletedLessons int             `json:"completed_lessons"`
	Units            []UnitProgress  `json:"units"`
	Skills           []SkillProgress `json:"skills"`
}

type UnitProgress struct {
	UnitID string `json:"unit_id"`
	Status string `json:"status"`
}

type SkillProgress struct {
	SkillID        string  `json:"skill_id"`
	UnitID         string  `json:"unit_id"`
	Status         string  `json:"status"`
	Level          int     `json:"level"`
	Mastery        float64 `json:"mastery"`
	CorrectAnswers int     `json:"correct_answers"`
	WrongAnswers   int     `json:"wrong_answers"`
}

type CatalogService interface {
	GetCatalogProgress(ctx context.Context, userID string, courseID string) (CatalogProgress, error)
	CompleteAllForUser(ctx context.Context, userID string) (CatalogProgress, error)
}
