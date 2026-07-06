-- Add self_rating and rh_rating columns to evaluation_competencies for independent
-- self-evaluation and RH-evaluation averages.
-- The existing `rating` column is retained for backward compatibility during transition.

-- +goose Up
-- +goose StatementBegin

DO $$ BEGIN ALTER TABLE evaluation_competencies ADD COLUMN self_rating INTEGER NULL CHECK (self_rating >= 1 AND self_rating <= 5); EXCEPTION WHEN duplicate_column THEN NULL; END $$;

DO $$ BEGIN ALTER TABLE evaluation_competencies ADD COLUMN rh_rating INTEGER NULL CHECK (rh_rating >= 1 AND rh_rating <= 5); EXCEPTION WHEN duplicate_column THEN NULL; END $$;

UPDATE evaluation_competencies ec
SET self_rating = ec.rating
FROM evaluations ev
WHERE ec.evaluation_id = ev.id
  AND ev.self_evaluation_completed_at IS NOT NULL
  AND ec.self_rating IS NULL;

-- +goose StatementEnd
