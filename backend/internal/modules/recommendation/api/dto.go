package api

import "context"

type Recommendation struct {
	Type     string  `json:"type"`
	CourseID string  `json:"course_id"`
	UnitID   *string `json:"unit_id,omitempty"`
	SkillID  *string `json:"skill_id,omitempty"`
	Reason   string  `json:"reason"`
}

type Service interface {
	GetNextSkill(ctx context.Context, userID, courseID string) (*Recommendation, error)
}
