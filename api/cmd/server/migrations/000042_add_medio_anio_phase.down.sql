-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_evaluations_emp_cycle_phase;
CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluations_emp_cycle ON evaluations (employee_id, cycle_id);
-- NOTE: PostgreSQL does not support ALTER TYPE ... DROP VALUE, so the
-- 'medio-anio' enum value is intentionally left in place on downgrade.
-- +goose StatementEnd
