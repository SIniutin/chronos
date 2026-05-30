-- +goose Up
CREATE TABLE challenges (
    id uuid PRIMARY KEY,
    skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    challenge_type varchar(64) NOT NULL,
    difficulty varchar(16) NOT NULL DEFAULT 'easy',
    tags jsonb NOT NULL DEFAULT '[]'::jsonb,
    level integer NOT NULL,
    lesson_count integer NOT NULL,
    prompt text NOT NULL,
    body text NOT NULL,
    payload jsonb NOT NULL DEFAULT '{}'::jsonb,
    options jsonb NOT NULL DEFAULT '[]'::jsonb,
    answers jsonb NOT NULL DEFAULT '[]'::jsonb,
    explanation text NOT NULL,
    position integer NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'draft',
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL,
    CONSTRAINT challenges_type_check CHECK (challenge_type IN ('theory', 'single_choice', 'multiple_choice', 'timeline', 'match_pairs', 'image_question', 'match_image', 'quote_question', 'true_false', 'fill_in_blank')),
    CONSTRAINT challenges_difficulty_check CHECK (difficulty IN ('easy', 'medium', 'hard', 'undefined')),
    CONSTRAINT challenges_status_check CHECK (status IN ('draft', 'updating', 'published', 'archived'))
);

CREATE INDEX challenges_skill_id_position_idx ON challenges(skill_id, position);

-- +goose Down
DROP TABLE challenges;
