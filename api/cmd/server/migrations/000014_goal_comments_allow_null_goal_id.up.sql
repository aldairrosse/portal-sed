-- +goose Up
-- +goose StatementBegin
ALTER TABLE goal_comments ALTER COLUMN goal_id DROP NOT NULL;
-- +goose StatementEnd
