-- +goose Up
CREATE TABLE sections (
    id uuid PRIMARY KEY,
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    theme varchar(255) NOT NULL,
    description text NOT NULL,
    position integer NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'draft',
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL,
    CONSTRAINT sections_status_check CHECK (status IN ('draft', 'updating', 'published', 'archived'))
);

CREATE TABLE section_descriptions (
    id uuid PRIMARY KEY,
    section_id uuid NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    title varchar(255) NOT NULL,
    text text NOT NULL,
    position integer NOT NULL,
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL
);

CREATE INDEX sections_course_id_position_idx ON sections(course_id, position);

-- +goose Down
DROP TABLE section_descriptions;
DROP TABLE sections;
