-- +goose Down
-- +goose StatementBegin

-- Remove seed phase_transitions for default transitions
DELETE FROM phase_transitions
WHERE (from_phase = 'asignacion' AND to_phase = 'avance' AND trigger = 'manual_rh')
   OR (from_phase = 'avance' AND to_phase = 'cierre' AND trigger = 'manual_rh');

-- Remove seed phase_definitions
DELETE FROM phase_definitions
WHERE phase IN ('asignacion', 'avance', 'cierre')
  AND label IN ('Asignación', 'Avance', 'Cierre');

-- +goose StatementEnd
