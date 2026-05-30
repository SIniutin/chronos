-- +goose Up
CREATE TABLE skill_progress (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    status varchar(32) NOT NULL DEFAULT 'locked',
    level integer NOT NULL DEFAULT 0,
    mastery double precision NOT NULL DEFAULT 0,
    correct_answers integer NOT NULL DEFAULT 0,
    wrong_answers integer NOT NULL DEFAULT 0,
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz DEFAULT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, skill_id),
    CONSTRAINT skill_progress_status_check CHECK (status IN ('locked', 'available', 'in_progress', 'completed')),
    CONSTRAINT skill_progress_level_check CHECK (level >= 0 AND level <= 5),
    CONSTRAINT skill_progress_mastery_check CHECK (mastery >= 0 AND mastery <= 1),
    CONSTRAINT skill_progress_answers_check CHECK (correct_answers >= 0 AND wrong_answers >= 0)
);

CREATE INDEX skill_progress_user_id_status_idx ON skill_progress(user_id, status);

-- +goose Down
DROP TABLE skill_progress;
