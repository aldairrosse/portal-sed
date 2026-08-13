-- +goose Up
-- +goose StatementBegin
ALTER TABLE nine_box_entries
    ADD COLUMN IF NOT EXISTS goal_progress_percent DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS self_rating DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS hr_rating DOUBLE PRECISION;
-- +goose StatementEnd
