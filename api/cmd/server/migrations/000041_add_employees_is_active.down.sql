-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_employees_is_active;
DROP INDEX IF EXISTS idx_employees_manager_is_active;
ALTER TABLE employees DROP COLUMN IF EXISTS is_active;
-- +goose StatementEnd
