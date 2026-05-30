package postgre

import (
	"context"
	"errors"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repoImpl struct {
	pool *pgxpool.Pool
}

func NewPostgreRepo(pool *pgxpool.Pool) *repoImpl {
	return &repoImpl{pool: pool}
}

type scanner interface {
	Scan(dest ...any) error
}

func (r *repoImpl) ListPublishedSections(ctx context.Context, courseID cd.CourseID) ([]cd.Section, error) {
	const query = `
		SELECT id, course_id, theme, description, position, status
		FROM sections
		WHERE course_id = $1 AND status = $2
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(courseID).String(), cd.ContentStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sections []cd.Section
	for rows.Next() {
		section, err := scanSection(rows)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *repoImpl) ListAllSections(ctx context.Context, courseID cd.CourseID) ([]cd.Section, error) {
	const query = `SELECT id, course_id, theme, description, position, status FROM sections WHERE course_id = $1 ORDER BY position`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(courseID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sections []cd.Section
	for rows.Next() {
		section, err := scanSection(rows)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}
	return sections, rows.Err()
}

func (r *repoImpl) GetSection(ctx context.Context, id cd.SectionID) (cd.Section, error) {
	const query = `SELECT id, course_id, theme, description, position, status FROM sections WHERE id = $1`
	return scanSection(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
}

func (r *repoImpl) ListPublishedUnits(ctx context.Context, sectionID cd.SectionID) ([]cd.Unit, error) {
	const query = `
		SELECT id, section_id, title, position, status
		FROM units
		WHERE section_id = $1 AND status = $2
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(sectionID).String(), cd.ContentStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var units []cd.Unit
	for rows.Next() {
		unit, err := scanUnit(rows)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func (r *repoImpl) ListAllUnits(ctx context.Context, sectionID cd.SectionID) ([]cd.Unit, error) {
	const query = `SELECT id, section_id, title, position, status FROM units WHERE section_id = $1 ORDER BY position`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(sectionID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var units []cd.Unit
	for rows.Next() {
		unit, err := scanUnit(rows)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func (r *repoImpl) GetUnit(ctx context.Context, id cd.UnitID) (cd.Unit, error) {
	const query = `SELECT id, section_id, title, position, status FROM units WHERE id = $1`
	return scanUnit(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
}

func (r *repoImpl) ListPublishedSkills(ctx context.Context, unitID cd.UnitID) ([]cd.Skill, error) {
	const query = `
		SELECT id, unit_id, title, icon, position, status
		FROM skills
		WHERE unit_id = $1 AND status = $2
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(unitID).String(), cd.ContentStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []cd.Skill
	for rows.Next() {
		skill, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (r *repoImpl) ListAllSkills(ctx context.Context, unitID cd.UnitID) ([]cd.Skill, error) {
	const query = `SELECT id, unit_id, title, icon, position, status FROM skills WHERE unit_id = $1 ORDER BY position`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(unitID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skills []cd.Skill
	for rows.Next() {
		skill, err := scanSkill(rows)
		if err != nil {
			return nil, err
		}
		skills = append(skills, skill)
	}
	return skills, rows.Err()
}

func (r *repoImpl) GetSkill(ctx context.Context, id cd.SkillID) (cd.Skill, error) {
	const query = `SELECT id, unit_id, title, icon, position, status FROM skills WHERE id = $1`
	return scanSkill(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
}

func (r *repoImpl) CreateSection(ctx context.Context, s cd.Section) (cd.Section, error) {
	const query = `
		INSERT INTO sections (id, course_id, theme, description, position, status, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, course_id, theme, description, position, status
	`
	return scanSection(r.pool.QueryRow(ctx, query, uuid.UUID(s.ID).String(), uuid.UUID(s.CourseID).String(), s.Theme, s.Description, s.Position, s.Status, uuidOrNil(s.Audit.CreatedBy), uuidOrNil(s.Audit.UpdatedBy), s.Audit.CreatedAt, s.Audit.UpdatedAt))
}

func (r *repoImpl) UpdateSection(ctx context.Context, s cd.Section) (cd.Section, error) {
	const query = `
		UPDATE sections SET course_id = $2, theme = $3, description = $4, position = $5,
		    status = CASE WHEN status = 'published' THEN 'updating' ELSE status END,
		    updated_by = $6, updated_at = $7
		WHERE id = $1
		RETURNING id, course_id, theme, description, position, status
	`
	return scanSection(r.pool.QueryRow(ctx, query, uuid.UUID(s.ID).String(), uuid.UUID(s.CourseID).String(), s.Theme, s.Description, s.Position, uuidOrNil(s.Audit.UpdatedBy), s.Audit.UpdatedAt))
}

func (r *repoImpl) CreateUnit(ctx context.Context, u cd.Unit) (cd.Unit, error) {
	const query = `
		INSERT INTO units (id, section_id, title, position, status, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, section_id, title, position, status
	`
	return scanUnit(r.pool.QueryRow(ctx, query, uuid.UUID(u.ID).String(), uuid.UUID(u.SectionID).String(), u.Title, u.Position, u.Status, uuidOrNil(u.Audit.CreatedBy), uuidOrNil(u.Audit.UpdatedBy), u.Audit.CreatedAt, u.Audit.UpdatedAt))
}

func (r *repoImpl) UpdateUnit(ctx context.Context, u cd.Unit) (cd.Unit, error) {
	const query = `
		UPDATE units SET section_id = $2, title = $3, position = $4,
		    status = CASE WHEN status = 'published' THEN 'updating' ELSE status END,
		    updated_by = $5, updated_at = $6
		WHERE id = $1
		RETURNING id, section_id, title, position, status
	`
	return scanUnit(r.pool.QueryRow(ctx, query, uuid.UUID(u.ID).String(), uuid.UUID(u.SectionID).String(), u.Title, u.Position, uuidOrNil(u.Audit.UpdatedBy), u.Audit.UpdatedAt))
}

func (r *repoImpl) CreateSkill(ctx context.Context, s cd.Skill) (cd.Skill, error) {
	const query = `
		INSERT INTO skills (id, unit_id, title, icon, position, status, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, unit_id, title, icon, position, status
	`
	return scanSkill(r.pool.QueryRow(ctx, query, uuid.UUID(s.ID).String(), uuid.UUID(s.UnitID).String(), s.Title, s.Icon, s.Position, s.Status, uuidOrNil(s.Audit.CreatedBy), uuidOrNil(s.Audit.UpdatedBy), s.Audit.CreatedAt, s.Audit.UpdatedAt))
}

func (r *repoImpl) UpdateSkill(ctx context.Context, s cd.Skill) (cd.Skill, error) {
	const query = `
		UPDATE skills SET unit_id = $2, title = $3, icon = $4, position = $5,
		    status = CASE WHEN status = 'published' THEN 'updating' ELSE status END,
		    updated_by = $6, updated_at = $7
		WHERE id = $1
		RETURNING id, unit_id, title, icon, position, status
	`
	return scanSkill(r.pool.QueryRow(ctx, query, uuid.UUID(s.ID).String(), uuid.UUID(s.UnitID).String(), s.Title, s.Icon, s.Position, uuidOrNil(s.Audit.UpdatedBy), s.Audit.UpdatedAt))
}

func (r *repoImpl) SetStatus(ctx context.Context, entity string, id uuid.UUID, status cd.ContentStatus, actorID cd.UserID) error {
	table, err := tableName(entity)
	if err != nil {
		return err
	}
	query := "UPDATE " + table + " SET status = $2, updated_by = $3, updated_at = now() WHERE id = $1"
	if status == cd.ContentStatusPublished {
		query += " AND status IN ('draft', 'updating')"
	}
	tag, err := r.pool.Exec(ctx, query, id.String(), status, uuid.UUID(actorID).String())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return cd.ErrNotFound
	}
	return nil
}

func scanCourse(row scanner) (cd.Course, error) {
	var course cd.Course
	var idRaw string
	if err := row.Scan(&idRaw, &course.SourceLang, &course.TargetLang, &course.Title, &course.Status); err != nil {
		return cd.Course{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return cd.Course{}, err
	}
	course.ID = cd.CourseID(id)
	return course, nil
}

func scanSection(row scanner) (cd.Section, error) {
	var section cd.Section
	var idRaw, courseIDRaw string
	if err := row.Scan(&idRaw, &courseIDRaw, &section.Theme, &section.Description, &section.Position, &section.Status); err != nil {
		return cd.Section{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return cd.Section{}, err
	}
	courseID, err := uuid.Parse(courseIDRaw)
	if err != nil {
		return cd.Section{}, err
	}
	section.ID = cd.SectionID(id)
	section.CourseID = cd.CourseID(courseID)
	return section, nil
}

func scanUnit(row scanner) (cd.Unit, error) {
	var unit cd.Unit
	var idRaw, sectionIDRaw string
	if err := row.Scan(&idRaw, &sectionIDRaw, &unit.Title, &unit.Position, &unit.Status); err != nil {
		return cd.Unit{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return cd.Unit{}, err
	}
	sectionID, err := uuid.Parse(sectionIDRaw)
	if err != nil {
		return cd.Unit{}, err
	}
	unit.ID = cd.UnitID(id)
	unit.SectionID = cd.SectionID(sectionID)
	return unit, nil
}

func scanSkill(row scanner) (cd.Skill, error) {
	var skill cd.Skill
	var idRaw, unitIDRaw string
	if err := row.Scan(&idRaw, &unitIDRaw, &skill.Title, &skill.Icon, &skill.Position, &skill.Status); err != nil {
		return cd.Skill{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return cd.Skill{}, err
	}
	unitID, err := uuid.Parse(unitIDRaw)
	if err != nil {
		return cd.Skill{}, err
	}
	skill.ID = cd.SkillID(id)
	skill.UnitID = cd.UnitID(unitID)
	return skill, nil
}

func mapPgError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return cd.ErrNotFound
	}
	return err
}

func uuidOrNil(id cd.UserID) any {
	if uuid.UUID(id) == uuid.Nil {
		return nil
	}
	return uuid.UUID(id).String()
}

func tableName(entity string) (string, error) {
	switch entity {
	case "courses", "course":
		return "courses", nil
	case "sections", "section":
		return "sections", nil
	case "units", "unit":
		return "units", nil
	case "skills", "skill":
		return "skills", nil
	case "challenges", "challenge":
		return "challenges", nil
	default:
		return "", cd.ErrInvalidInput
	}
}
