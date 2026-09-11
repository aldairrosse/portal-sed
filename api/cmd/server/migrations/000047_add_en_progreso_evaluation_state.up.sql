-- +goose Up
-- +goose StatementBegin
-- Add 'en_progreso' to evaluation_state. Used by the cierre state machine
-- (pendiente_evaluacion_final -> en_progreso -> completada, see
-- internal/pkg/state/machine.go) and written by SubmitSelf/SubmitRH flows.
-- ALTER TYPE ... ADD VALUE cannot use IF NOT EXISTS on all PG versions,
-- so guard with DO ... EXCEPTION (same as 000031, 000042).
-- Statement-only (no DML referencing the new value) — safe in a transaction.
DO $$ BEGIN
    ALTER TYPE evaluation_state ADD VALUE 'en_progreso';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
-- +goose StatementEnd
