-- +goose Up
CREATE TABLE user_streaks (
    user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_days integer NOT NULL DEFAULT 0,
    longest_days integer NOT NULL DEFAULT 0,
    last_activity_date date NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT user_streaks_days_check CHECK (current_days >= 0 AND longest_days >= 0)
);

-- +goose Down
DROP TABLE user_streaks;
