-- +goose Up
CREATE TABLE user_xp (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_xp integer NOT NULL DEFAULT 0,
    level integer NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_xp_non_negative_check CHECK (total_xp >= 0 AND level >= 1)
);

-- +goose Down
DROP TABLE user_xp;
