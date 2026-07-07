-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goal_comments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    goal_id     UUID        NOT NULL,
    author_id   TEXT        NOT NULL,
    author_name TEXT        NOT NULL,
    content     TEXT        NOT NULL,
    CONSTRAINT fk_goal_comments_goal
        FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_goal_comments_goal ON goal_comments (goal_id);
-- +goose StatementEnd
