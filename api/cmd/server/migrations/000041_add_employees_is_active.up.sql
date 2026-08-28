-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT true;
CREATE INDEX IF NOT EXISTS idx_employees_is_active ON employees(is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_employees_manager_is_active ON employees(manager_id, is_active);
-- +goose StatementEnd
