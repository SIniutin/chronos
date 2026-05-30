-- +goose Up
CREATE TABLE user_achievements (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    achievement_id uuid NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    unlocked_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, achievement_id)
);

CREATE INDEX user_achievements_user_id_unlocked_at_idx ON user_achievements(user_id, unlocked_at DESC);

-- +goose Down
DROP TABLE user_achievements;
