-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goal_progress_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id     UUID        NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    value       FLOAT       NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID        NULL REFERENCES employees(id)
);

CREATE INDEX IF NOT EXISTS idx_goal_progress_logs_goal_time
    ON goal_progress_logs (goal_id, recorded_at DESC);
-- +goose StatementEnd
