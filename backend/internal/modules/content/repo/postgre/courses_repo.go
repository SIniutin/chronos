package postgre

import (
	"context"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

type CoursesRepository interface {
	ListPublishedCourses(ctx context.Context) ([]cd.Course, error)
	ListAllCourses(ctx context.Context) ([]cd.Course, error)
	CreateCourse(ctx context.Context, course cd.Course) (cd.Course, error)
	UpdateCourse(ctx context.Context, course cd.Course) (cd.Course, error)
	SetStatus(ctx context.Context, id cd.CourseID, status cd.ContentStatus, actorID cd.UserID) error
}

func (r *repoImpl) ListPublishedCourses(ctx context.Context) ([]cd.Course, error) {
	const query = `
		SELECT id, source_lang, target_lang, title, status
		FROM courses
		WHERE status = $1
		ORDER BY title
	`
	rows, err := r.pool.Query(ctx, query, cd.ContentStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var courses []cd.Course
	for rows.Next() {
		var course cd.Course
		var idRaw string
		if err := rows.Scan(&idRaw, &course.SourceLang, &course.TargetLang, &course.Title, &course.Status); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(idRaw)
		if err != nil {
			return nil, err
		}
		course.ID = cd.CourseID(id)
		courses = append(courses, course)
	}
	return courses, rows.Err()
}

func (r *repoImpl) ListAllCourses(ctx context.Context) ([]cd.Course, error) {
	const query = `SELECT id, source_lang, target_lang, title, status FROM courses ORDER BY title`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var courses []cd.Course
	for rows.Next() {
		var course cd.Course
		var idRaw string
		if err := rows.Scan(&idRaw, &course.SourceLang, &course.TargetLang, &course.Title, &course.Status); err != nil {
			return nil, err
		}
		id, err := uuid.Parse(idRaw)
		if err != nil {
			return nil, err
		}
		course.ID = cd.CourseID(id)
		courses = append(courses, course)
	}
	return courses, rows.Err()
}

func (r *repoImpl) CreateCourse(ctx context.Context, c cd.Course) (cd.Course, error) {
	const query = `
		INSERT INTO courses (id, source_lang, target_lang, title, status, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, source_lang, target_lang, title, status
	`
	return scanCourse(r.pool.QueryRow(ctx, query, uuid.UUID(c.ID).String(), c.SourceLang, c.TargetLang, c.Title, c.Status, uuidOrNil(c.Audit.CreatedBy), uuidOrNil(c.Audit.UpdatedBy), c.Audit.CreatedAt, c.Audit.UpdatedAt))
}

func (r *repoImpl) UpdateCourse(ctx context.Context, c cd.Course) (cd.Course, error) {
	const query = `
		UPDATE courses
		SET source_lang = $2, target_lang = $3, title = $4,
		    status = CASE WHEN status = 'published' THEN 'updating' ELSE status END,
		    updated_by = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, source_lang, target_lang, title, status
	`
	return scanCourse(r.pool.QueryRow(ctx, query, uuid.UUID(c.ID).String(), c.SourceLang, c.TargetLang, c.Title, uuidOrNil(c.Audit.UpdatedBy), c.Audit.UpdatedAt))
}
