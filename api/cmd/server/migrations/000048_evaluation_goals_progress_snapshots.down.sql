-- +goose Down
-- +goose StatementBegin
ALTER TABLE evaluation_goals
    DROP COLUMN IF EXISTS cierre_progress,
    DROP COLUMN IF EXISTS avance_progress;
-- +goose StatementEnd
