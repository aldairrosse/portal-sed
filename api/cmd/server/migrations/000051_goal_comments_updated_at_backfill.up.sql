-- +goose Up
-- +goose StatementBegin
-- Backfill goal_comments.updated_at before Ent auto-migrate sets NOT NULL.
-- goal_comments (000010) was created with only created_at; Ent expects
-- updated_at NOT NULL and fails with "column updated_at contains null values".
ALTER TABLE goal_comments ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ;

UPDATE goal_comments
SET updated_at = COALESCE(updated_at, created_at, NOW())
WHERE updated_at IS NULL;
-- +goose StatementEnd
