-- +goose Up
-- +goose StatementBegin

-- Seed phase_definitions for cycles that lack them.
INSERT INTO phase_definitions (id, created_at, updated_at, phase, label, "order", cycle_id)
SELECT gen_random_uuid(), NOW(), NOW(), 'asignacion', 'Asignación', 1, c.id
FROM cycles c
WHERE NOT EXISTS (SELECT 1 FROM phase_definitions pd WHERE pd.cycle_id = c.id AND pd.phase = 'asignacion');

INSERT INTO phase_definitions (id, created_at, updated_at, phase, label, "order", cycle_id)
SELECT gen_random_uuid(), NOW(), NOW(), 'avance', 'Avance', 2, c.id
FROM cycles c
WHERE NOT EXISTS (SELECT 1 FROM phase_definitions pd WHERE pd.cycle_id = c.id AND pd.phase = 'avance');

INSERT INTO phase_definitions (id, created_at, updated_at, phase, label, "order", cycle_id)
SELECT gen_random_uuid(), NOW(), NOW(), 'cierre', 'Cierre', 3, c.id
FROM cycles c
WHERE NOT EXISTS (SELECT 1 FROM phase_definitions pd WHERE pd.cycle_id = c.id AND pd.phase = 'cierre');

-- Seed phase_transitions: asignacion → avance
INSERT INTO phase_transitions (id, from_phase, to_phase, trigger, created_at, cycle_id, from_phase_id, to_phase_id)
SELECT gen_random_uuid(), 'asignacion', 'avance', 'manual_rh', NOW(), c.id, pd_from.id, pd_to.id
FROM cycles c
JOIN phase_definitions pd_from ON pd_from.cycle_id = c.id AND pd_from.phase = 'asignacion'
JOIN phase_definitions pd_to   ON pd_to.cycle_id   = c.id AND pd_to.phase   = 'avance'
WHERE NOT EXISTS (
    SELECT 1 FROM phase_transitions pt
    WHERE pt.cycle_id = c.id AND pt.from_phase = 'asignacion' AND pt.to_phase = 'avance'
);

-- Seed phase_transitions: avance → cierre
INSERT INTO phase_transitions (id, from_phase, to_phase, trigger, created_at, cycle_id, from_phase_id, to_phase_id)
SELECT gen_random_uuid(), 'avance', 'cierre', 'manual_rh', NOW(), c.id, pd_from.id, pd_to.id
FROM cycles c
JOIN phase_definitions pd_from ON pd_from.cycle_id = c.id AND pd_from.phase = 'avance'
JOIN phase_definitions pd_to   ON pd_to.cycle_id   = c.id AND pd_to.phase   = 'cierre'
WHERE NOT EXISTS (
    SELECT 1 FROM phase_transitions pt
    WHERE pt.cycle_id = c.id AND pt.from_phase = 'avance' AND pt.to_phase = 'cierre'
);

-- +goose StatementEnd
