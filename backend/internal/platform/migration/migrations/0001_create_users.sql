-- +goose Up
CREATE TABLE users (
    id uuid PRIMARY KEY,
    email varchar(254) NOT NULL UNIQUE,
    login varchar(32) NOT NULL UNIQUE,
    password_hash text NOT NULL,
    password_hash_algo varchar(32) NOT NULL,
    password_changed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    role varchar(32) NOT NULL DEFAULT 'student',
    CONSTRAINT users_role_check CHECK (role IN ('student', 'content_editor', 'content_reviewer', 'admin'))
);

-- +goose Down
DROP TABLE users;
