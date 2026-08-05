-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS refresh_expires_at TIMESTAMPTZ NULL;
-- +goose StatementEnd