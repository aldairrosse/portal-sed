-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS idx_employees_employee_number ON employees (employee_number);
-- +goose StatementEnd
