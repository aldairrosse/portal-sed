-- +goose Down
-- +goose StatementBegin

DELETE FROM phase_transitions
WHERE from_phase = 'avance' AND to_phase = 'asignacion'
  AND trigger = 'manual_rh';

DELETE FROM phase_transitions
WHERE from_phase = 'cierre' AND to_phase = 'avance'
  AND trigger = 'manual_rh';

-- +goose StatementEnd
