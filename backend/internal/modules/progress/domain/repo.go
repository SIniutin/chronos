package domain

import "context"

type Repository interface {
	GetCourseProgress(ctx context.Context, userID UserID, courseID CourseID) (*CourseProgress, error)
	SaveCourseProgress(ctx context.Context, progress CourseProgress) error
	GetUnitProgress(ctx context.Context, userID UserID, unitID UnitID) (*UnitProgress, error)
	SaveUnitProgress(ctx context.Context, progress UnitProgress) error
	GetSkillProgress(ctx context.Context, userID UserID, skillID SkillID) (*SkillProgress, error)
	ListSkillProgressByUser(ctx context.Context, userID UserID) ([]SkillProgress, error)
	SaveSkillProgress(ctx context.Context, progress SkillProgress) error
}
