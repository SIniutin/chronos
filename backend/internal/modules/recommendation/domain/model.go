package domain

import (
	content_domain "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
)

type UserID = users_domain.UserID
type CourseID = content_domain.CourseID
type UnitID = content_domain.UnitID
type SkillID = content_domain.SkillID
type ChallengeID = content_domain.ChallengeID

type RecommendationType string

const (
	RecommendationTypeContinue RecommendationType = "continue"
	RecommendationTypeReview   RecommendationType = "review"
	RecommendationTypeNewSkill RecommendationType = "new_skill"
)

type Recommendation struct {
	Type     RecommendationType
	CourseID CourseID
	UnitID   *UnitID
	SkillID  *SkillID
	Reason   string
}
