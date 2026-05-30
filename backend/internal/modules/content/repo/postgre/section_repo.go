package postgre

import (
	"context"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
)

type SectionsRepository interface {
	ListPublishedSections(ctx context.Context, courseID cd.CourseID) ([]cd.Section, error)
	ListAllSections(ctx context.Context, courseID cd.CourseID) ([]cd.Section, error)
	CreateSection(ctx context.Context, section cd.Section) (cd.Section, error)
	UpdateSection(ctx context.Context, section cd.Section) (cd.Section, error)
	SetStatus(ctx context.Context, id cd.SectionID, status cd.ContentStatus, actorID cd.UserID) error
}
