-- +goose Up
CREATE TABLE unit_progress (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    unit_id uuid NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    status varchar(32) NOT NULL DEFAULT 'locked',
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz DEFAULT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, unit_id),
    CONSTRAINT unit_progress_status_check CHECK (status IN ('locked', 'available', 'in_progress', 'completed'))
);

CREATE INDEX unit_progress_user_id_status_idx ON unit_progress(user_id, status);

-- +goose Down
DROP TABLE unit_progress;
