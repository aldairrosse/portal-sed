-- +goose Up
-- +goose StatementBegin

ALTER TABLE sessions ADD COLUMN IF NOT EXISTS id_token      TEXT;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS access_token  TEXT;
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS refresh_token TEXT;

-- +goose StatementEnd
