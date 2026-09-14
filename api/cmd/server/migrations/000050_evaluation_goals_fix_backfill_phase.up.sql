-- +goose Up
-- +goose StatementBegin
-- Fix 000048 unconditional backfill: copy misplaced avance snapshot to cierre
-- where cycle phase is cierre. Preserves avance snapshot (no NULL wipe).
-- Idempotent: only touches rows with
-- avance_progress set and cierre_progress still NULL.
UPDATE evaluation_goals eg
SET cierre_progress = eg.avance_progress
FROM evaluations e
JOIN cycles c ON c.id = e.cycle_id
WHERE eg.evaluation_id = e.id
  AND c.current_phase = 'cierre'
  AND eg.avance_progress IS NOT NULL
  AND eg.cierre_progress IS NULL;

-- Backfill any cierre row still without snapshot from the live goal value
-- (same guards as 000048, safe on re-run).
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
