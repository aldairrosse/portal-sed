-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uniq_active_per_org;
ALTER TABLE cycles DROP COLUMN IF EXISTS is_active;
-- +goose StatementEnd
