-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_pillars_type;
ALTER TABLE pillars DROP COLUMN IF EXISTS type;
-- +goose StatementEnd
