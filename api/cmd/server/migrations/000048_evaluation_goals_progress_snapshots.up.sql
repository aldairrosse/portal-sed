-- +goose Up
-- +goose StatementBegin
-- F1 sed-snapshots-metas-9box: per-phase direct-value snapshots on evaluation_goals.
-- Sync rule: goals.current_value = cierre_progress ?? avance_progress.
ALTER TABLE evaluation_goals
    ADD COLUMN IF NOT EXISTS avance_progress DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS cierre_progress DOUBLE PRECISION NULL;

-- Backfill: phase-conditional snapshot from the live goal value.
-- avance phase -> avance_progress, cierre phase -> cierre_progress.
-- COALESCE + IS NULL guards keep each branch idempotent on re-run.
UPDATE evaluation_goals eg
SET avance_progress = COALESCE(eg.avance_progress, g.current_value)
FROM goals g, evaluations e, cycles c
WHERE eg.goal_id = g.id
  AND eg.evaluation_id = e.id
  AND c.id = e.cycle_id
  AND c.current_phase = 'avance'
  AND g.current_value IS NOT NULL
  AND g.current_value <> 0
  AND eg.avance_progress IS NULL;

UPDATE evaluation_goals eg
SET cierre_progress = COALESCE(eg.cierre_progress, g.current_value)
FROM goals g, evaluations e, cycles c
WHERE eg.goal_id = g.id
  AND eg.evaluation_id = e.id
  AND c.id = e.cycle_id
  AND c.current_phase = 'cierre'
  AND g.current_value IS NOT NULL
  AND g.current_value <> 0
  AND eg.cierre_progress IS NULL;
-- +goose StatementEnd
