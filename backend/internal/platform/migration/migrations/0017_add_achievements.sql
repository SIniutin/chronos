-- +goose Up
CREATE TABLE achievements (
    id uuid PRIMARY KEY,
    code varchar(64) NOT NULL UNIQUE,
    title varchar(255) NOT NULL,
    description text NOT NULL,
    xp_reward integer NOT NULL DEFAULT 0,
    condition jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO achievements (id, code, title, description, xp_reward, condition) VALUES
('11111111-1111-4111-8111-111111111111', 'first_session', 'Первый урок', 'Завершить первую учебную сессию.', 0, '{"event":"session_completed"}'::jsonb),
('22222222-2222-4222-8222-222222222222', 'perfect_session', 'Без ошибок', 'Завершить сессию со 100% правильных ответов.', 0, '{"event":"perfect_session"}'::jsonb),
('33333333-3333-4333-8333-333333333333', 'streak_3', 'Три дня подряд', 'Заниматься три дня подряд.', 0, '{"streak_days":3}'::jsonb);

-- +goose Down
DROP TABLE achievements;
