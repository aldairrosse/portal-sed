-- +goose Up
-- +goose StatementBegin
ALTER TABLE nine_box_quadrants ADD COLUMN IF NOT EXISTS title TEXT;
ALTER TABLE nine_box_quadrants ADD COLUMN IF NOT EXISTS color_hex TEXT;
-- +goose StatementEnd
