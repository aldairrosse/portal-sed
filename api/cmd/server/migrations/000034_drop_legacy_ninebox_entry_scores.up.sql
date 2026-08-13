-- +goose Up
-- +goose StatementBegin
ALTER TABLE nine_box_entries
    DROP COLUMN IF EXISTS performance_score,
    DROP COLUMN IF EXISTS potential_score;
-- +goose StatementEnd
