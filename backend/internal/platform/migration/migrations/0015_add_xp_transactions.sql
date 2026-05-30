-- +goose Up
CREATE TABLE xp_transactions (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount integer NOT NULL,
    reason varchar(64) NOT NULL,
    source_type varchar(64) NOT NULL,
    source_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT xp_transactions_amount_check CHECK (amount > 0),
    CONSTRAINT xp_transactions_reason_check CHECK (reason IN ('session_completed', 'correct_answer', 'perfect_session', 'achievement', 'daily_goal')),
    CONSTRAINT xp_transactions_source_unique UNIQUE (user_id, reason, source_type, source_id)
);

CREATE INDEX xp_transactions_user_id_created_at_idx ON xp_transactions(user_id, created_at DESC);

-- +goose Down
DROP TABLE xp_transactions;
