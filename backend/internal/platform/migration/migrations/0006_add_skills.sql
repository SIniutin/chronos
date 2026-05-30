-- +goose Up
CREATE TABLE skills (
    id uuid PRIMARY KEY,
    unit_id uuid NOT NULL REFERENCES units(id) ON DELETE CASCADE,
    title varchar(255) NOT NULL,
    icon varchar(64) NOT NULL,
    position integer NOT NULL,
    status varchar(32) NOT NULL DEFAULT 'draft',
    created_by uuid REFERENCES users(id),
    updated_by uuid REFERENCES users(id),
    reviewed_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    reviewed_at timestamptz DEFAULT NULL,
    CONSTRAINT skills_status_check CHECK (status IN ('draft', 'updating', 'published', 'archived'))
);

CREATE INDEX skills_unit_id_position_idx ON skills(unit_id, position);

-- +goose Down
DROP TABLE skills;
