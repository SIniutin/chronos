-- +goose Up
CREATE TABLE course_progress (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    status varchar(32) NOT NULL DEFAULT 'locked',
    started_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz DEFAULT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, course_id),
    CONSTRAINT course_progress_status_check CHECK (status IN ('locked', 'available', 'in_progress', 'completed'))
);

CREATE INDEX course_progress_user_id_status_idx ON course_progress(user_id, status);

-- +goose Down
DROP TABLE course_progress;
