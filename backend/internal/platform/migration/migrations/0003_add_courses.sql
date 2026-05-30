-- +goose Up
CREATE TABLE courses (
    id uuid PRIMARY KEY,
    source_lang varchar(16) NOT NULL,
    target_lang varchar(16) NOT NULL,
    title varchar(255) NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'draft',
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL,
    CONSTRAINT courses_status_check CHECK (status IN ('draft', 'updating', 'published', 'archived'))
);

-- +goose Down
DROP TABLE courses;
