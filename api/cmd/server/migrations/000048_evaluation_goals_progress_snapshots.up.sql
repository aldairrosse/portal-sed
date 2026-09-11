-- +goose Up
-- +goose StatementBegin
-- F1 sed-snapshots-metas-9box: per-phase direct-value snapshots on evaluation_goals.
-- Sync rule: goals.current_value = cierre_progress ?? avance_progress.
ALTER TABLE evaluation_goals
    ADD COLUMN IF NOT EXISTS avance_progress DOUBLE PRECISION NULL,
    ADD COLUMN IF NOT EXISTS cierre_progress DOUBLE PRECISION NULL;

-- Backfill: avance snapshot from the live goal value for rows that carry progress.
-- COALESCE keeps the form idempotent on re-run.
UPDATE evaluation_goals eg
SET avance_progress = COALESCE(eg.avance_progress, g.current_value)
FROM goals g
WHERE eg.goal_id = g.id
  AND g.current_value <> 0
  AND eg.avance_progress IS NULL;
-- +goose StatementEnd
