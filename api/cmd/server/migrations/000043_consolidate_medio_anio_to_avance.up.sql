-- +goose Up
-- +goose StatementBegin
-- Consolidate legacy 'medio-anio' rows into canonical 'avance'.
-- Enum value is kept for read compatibility; writes use asignacion/avance/cierre.
UPDATE cycles SET current_phase = 'avance', updated_at = NOW() WHERE current_phase = 'medio-anio';
UPDATE phase_definitions SET phase = 'avance', updated_at = NOW() WHERE phase = 'medio-anio';
UPDATE phase_transitions SET from_phase = 'avance' WHERE from_phase = 'medio-anio';
UPDATE phase_transitions SET to_phase = 'avance' WHERE to_phase = 'medio-anio';
UPDATE cycle_phase_history SET from_phase = 'avance' WHERE from_phase = 'medio-anio';
UPDATE cycle_phase_history SET to_phase = 'avance' WHERE to_phase = 'medio-anio';
-- +goose StatementEnd
