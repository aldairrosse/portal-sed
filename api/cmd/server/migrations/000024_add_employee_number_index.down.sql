-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_employees_employee_number;
-- +goose StatementEnd
