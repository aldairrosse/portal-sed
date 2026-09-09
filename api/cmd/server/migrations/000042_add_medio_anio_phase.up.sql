-- +goose Up
-- +goose StatementBegin
-- Add 'medio-anio' phase. ALTER TYPE ... ADD VALUE cannot use IF NOT EXISTS
-- on all PG versions, so guard with DO ... EXCEPTION (same as 000031).
DO $$ BEGIN
    ALTER TYPE phase ADD VALUE 'medio-anio';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DROP INDEX IF EXISTS idx_evaluations_emp_cycle;
CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluations_emp_cycle_phase ON evaluations (employee_id, cycle_id, phase);
CREATE INDEX IF NOT EXISTS idx_evaluations_cycle_phase ON evaluations (cycle_id, phase);
-- +goose StatementEnd
