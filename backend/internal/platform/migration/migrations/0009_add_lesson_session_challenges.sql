-- +goose Up
CREATE TABLE lesson_session_challenges (
    id uuid PRIMARY KEY,
    session_id uuid NOT NULL REFERENCES lesson_sessions(id) ON DELETE CASCADE,
    challenge_id uuid NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    position integer NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'pending',
    CONSTRAINT lesson_session_challenges_status_check CHECK (status IN ('pending', 'answered')),
    CONSTRAINT lesson_session_challenges_session_position_unique UNIQUE (session_id, position)
);

CREATE INDEX lesson_session_challenges_session_status_position_idx ON lesson_session_challenges(session_id, status, position);
CREATE INDEX lesson_session_challenges_challenge_id_idx ON lesson_session_challenges(challenge_id);

-- +goose Down
DROP TABLE lesson_session_challenges;
