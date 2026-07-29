-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS acr          TEXT;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS requires_2fa BOOLEAN NOT NULL DEFAULT false;
-- +goose StatementEnd
