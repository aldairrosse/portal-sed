-- +goose Up
-- +goose StatementBegin

-- Seed backward phase_transitions: avance → asignacion
INSERT INTO phase_transitions (id, from_phase, to_phase, trigger, created_at, cycle_id, from_phase_id, to_phase_id)
SELECT gen_random_uuid(), 'avance', 'asignacion', 'manual_rh', NOW(), c.id, pd_from.id, pd_to.id
FROM cycles c
JOIN phase_definitions pd_from ON pd_from.cycle_id = c.id AND pd_from.phase = 'avance'
JOIN phase_definitions pd_to   ON pd_to.cycle_id   = c.id AND pd_to.phase   = 'asignacion'
WHERE NOT EXISTS (
    SELECT 1 FROM phase_transitions pt
    WHERE pt.cycle_id = c.id AND pt.from_phase = 'avance' AND pt.to_phase = 'asignacion'
);

-- Seed backward phase_transitions: cierre → avance
INSERT INTO phase_transitions (id, from_phase, to_phase, trigger, created_at, cycle_id, from_phase_id, to_phase_id)
SELECT gen_random_uuid(), 'cierre', 'avance', 'manual_rh', NOW(), c.id, pd_from.id, pd_to.id
FROM cycles c
JOIN phase_definitions pd_from ON pd_from.cycle_id = c.id AND pd_from.phase = 'cierre'
JOIN phase_definitions pd_to   ON pd_to.cycle_id   = c.id AND pd_to.phase   = 'avance'
WHERE NOT EXISTS (
    SELECT 1 FROM phase_transitions pt
    WHERE pt.cycle_id = c.id AND pt.from_phase = 'cierre' AND pt.to_phase = 'avance'
);

-- +goose StatementEnd
