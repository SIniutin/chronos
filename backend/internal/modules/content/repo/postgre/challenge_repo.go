package postgre

import (
	"context"

	cd "github.com/SIniutin/history-app-backend/internal/modules/content/domain"
	"github.com/google/uuid"
)

type ChallengeRepository interface {
	ListPublishedChallenges(ctx context.Context, skillID cd.SkillID) ([]cd.Challenge, error)
	ListAllChallenges(ctx context.Context, skillID cd.SkillID) ([]cd.Challenge, error)
	GetChallenge(ctx context.Context, id cd.ChallengeID) (cd.Challenge, error)
	CreateChallenge(ctx context.Context, challenge cd.Challenge) (cd.Challenge, error)
	UpdateChallenge(ctx context.Context, challenge cd.Challenge) (cd.Challenge, error)
	SetStatus(ctx context.Context, id cd.ChallengeID, status cd.ContentStatus, actorID cd.UserID) error
}

func (r *repoImpl) ListPublishedChallenges(ctx context.Context, skillID cd.SkillID) ([]cd.Challenge, error) {
	const query = `
		SELECT id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status
		FROM challenges
		WHERE skill_id = $1 AND status = $2
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(skillID).String(), cd.ContentStatusPublished)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var challenges []cd.Challenge
	for rows.Next() {
		challenge, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func (r *repoImpl) ListAllChallenges(ctx context.Context, skillID cd.SkillID) ([]cd.Challenge, error) {
	const query = `
		SELECT id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status
		FROM challenges
		WHERE skill_id = $1
		ORDER BY position
	`
	rows, err := r.pool.Query(ctx, query, uuid.UUID(skillID).String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var challenges []cd.Challenge
	for rows.Next() {
		challenge, err := scanChallenge(rows)
		if err != nil {
			return nil, err
		}
		challenges = append(challenges, challenge)
	}
	return challenges, rows.Err()
}

func (r *repoImpl) GetChallenge(ctx context.Context, id cd.ChallengeID) (cd.Challenge, error) {
	const query = `
		SELECT id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status
		FROM challenges
		WHERE id = $1
	`
	return scanChallenge(r.pool.QueryRow(ctx, query, uuid.UUID(id).String()))
}

func (r *repoImpl) CreateChallenge(ctx context.Context, c cd.Challenge) (cd.Challenge, error) {
	const query = `
		INSERT INTO challenges (id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status, created_by, updated_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
		RETURNING id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status
	`
	return scanChallenge(r.pool.QueryRow(ctx, query, uuid.UUID(c.ID).String(), uuid.UUID(c.SkillID).String(), c.Type, c.Difficulty, c.Tags, c.Level, c.LessonCount, c.Prompt, c.Body, c.Payload, c.Options, c.Answers, c.Explanation, c.Position, c.Status, uuidOrNil(c.Audit.CreatedBy), uuidOrNil(c.Audit.UpdatedBy), c.Audit.CreatedAt, c.Audit.UpdatedAt))
}

func (r *repoImpl) UpdateChallenge(ctx context.Context, c cd.Challenge) (cd.Challenge, error) {
	const query = `
		UPDATE challenges SET skill_id = $2, challenge_type = $3, difficulty = $4, tags = $5, level = $6, lesson_count = $7,
		    prompt = $8, body = $9, payload = $10, options = $11, answers = $12, explanation = $13, position = $14,
		    status = CASE WHEN status = 'published' THEN 'updating' ELSE status END,
		    updated_by = $15, updated_at = $16
		WHERE id = $1
		RETURNING id, skill_id, challenge_type, difficulty, tags, level, lesson_count, prompt, body, payload, options, answers, explanation, position, status
	`
	return scanChallenge(r.pool.QueryRow(ctx, query, uuid.UUID(c.ID).String(), uuid.UUID(c.SkillID).String(), c.Type, c.Difficulty, c.Tags, c.Level, c.LessonCount, c.Prompt, c.Body, c.Payload, c.Options, c.Answers, c.Explanation, c.Position, uuidOrNil(c.Audit.UpdatedBy), c.Audit.UpdatedAt))
}

func scanChallenge(row scanner) (cd.Challenge, error) {
	var challenge cd.Challenge
	var idRaw, skillIDRaw string
	if err := row.Scan(
		&idRaw,
		&skillIDRaw,
		&challenge.Type,
		&challenge.Difficulty,
		&challenge.Tags,
		&challenge.Level,
		&challenge.LessonCount,
		&challenge.Prompt,
		&challenge.Body,
		&challenge.Payload,
		&challenge.Options,
		&challenge.Answers,
		&challenge.Explanation,
		&challenge.Position,
		&challenge.Status,
	); err != nil {
		return cd.Challenge{}, mapPgError(err)
	}
	id, err := uuid.Parse(idRaw)
	if err != nil {
		return cd.Challenge{}, err
	}
	skillID, err := uuid.Parse(skillIDRaw)
	if err != nil {
		return cd.Challenge{}, err
	}
	challenge.ID = cd.ChallengeID(id)
	challenge.SkillID = cd.SkillID(skillID)
	return challenge, nil
}
