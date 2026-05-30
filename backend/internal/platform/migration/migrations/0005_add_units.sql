-- +goose Up
CREATE TABLE units (
    id uuid PRIMARY KEY,
    section_id uuid NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    title varchar(255) NOT NULL,
    position integer NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'draft',
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL,
    CONSTRAINT units_status_check CHECK (status IN ('draft', 'updating', 'published', 'archived'))
);

CREATE INDEX units_section_id_position_idx ON units(section_id, position);

-- +goose Down
DROP TABLE units;
