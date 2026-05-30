-- +goose Up
CREATE TABLE lesson_sessions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    skill_id uuid NOT NULL REFERENCES skills(id) ON DELETE CASCADE,
    status varchar(32) NOT NULL DEFAULT 'active',
    started_at timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz DEFAULT NULL,
    CONSTRAINT lesson_sessions_status_check CHECK (status IN ('active', 'finished'))
);

CREATE INDEX lesson_sessions_user_id_started_at_idx ON lesson_sessions(user_id, started_at DESC);
CREATE INDEX lesson_sessions_skill_id_idx ON lesson_sessions(skill_id);

-- +goose Down
DROP TABLE lesson_sessions;
