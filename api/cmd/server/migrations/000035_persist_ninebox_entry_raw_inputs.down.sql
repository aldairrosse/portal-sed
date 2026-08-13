-- +goose Down
-- +goose StatementBegin
ALTER TABLE nine_box_entries
    DROP COLUMN IF EXISTS goal_progress_percent,
    DROP COLUMN IF EXISTS self_rating,
    DROP COLUMN IF EXISTS hr_rating;
-- +goose StatementEnd
