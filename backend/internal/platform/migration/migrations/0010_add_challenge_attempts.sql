-- +goose Up
CREATE TABLE challenge_attempts (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES lesson_sessions(id) ON DELETE CASCADE,
    session_challenge_id uuid NOT NULL REFERENCES lesson_session_challenges(id) ON DELETE CASCADE,
    challenge_id uuid NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    user_answer jsonb NOT NULL DEFAULT 'null'::jsonb,
    is_correct boolean NOT NULL DEFAULT false,
    mistakes text[] NOT NULL DEFAULT ARRAY[]::text[],
    answered_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX challenge_attempts_session_id_answered_at_idx ON challenge_attempts(session_id, answered_at);
CREATE INDEX challenge_attempts_session_challenge_id_idx ON challenge_attempts(session_challenge_id);

-- +goose Down
DROP TABLE challenge_attempts;
