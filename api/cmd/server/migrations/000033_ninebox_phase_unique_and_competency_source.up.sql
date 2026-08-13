-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_nine_box_matrixes_cycle_eval;

DO $$ BEGIN
    CREATE TYPE evaluation_competency_source AS ENUM ('self', 'rh');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE evaluation_competencies
    ADD COLUMN IF NOT EXISTS source evaluation_competency_source NOT NULL DEFAULT 'rh';

UPDATE evaluation_competencies ec
SET source = 'self'
FROM evaluations ev
WHERE ec.evaluation_id = ev.id
  AND ev.self_evaluation_completed_at IS NOT NULL
  AND (ev.rh_evaluation_completed_at IS NULL
       OR ev.self_evaluation_completed_at <= ev.rh_evaluation_completed_at);
-- +goose StatementEnd
