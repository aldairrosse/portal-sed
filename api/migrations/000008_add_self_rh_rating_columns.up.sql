-- Add self_rating and rh_rating columns to evaluation_competencies for independent
-- self-evaluation and RH-evaluation averages.
-- The existing `rating` column is retained for backward compatibility during transition.

ALTER TABLE evaluation_competencies
  ADD COLUMN self_rating INTEGER NULL CHECK (self_rating >= 1 AND self_rating <= 5);

ALTER TABLE evaluation_competencies
  ADD COLUMN rh_rating INTEGER NULL CHECK (rh_rating >= 1 AND rh_rating <= 5);

-- Backfill: copy existing rating → self_rating for evaluations where the employee
-- has completed their self-evaluation (self_evaluation_completed_at IS NOT NULL).
UPDATE evaluation_competencies ec
SET self_rating = ec.rating
FROM evaluations ev
WHERE ec.evaluation_id = ev.id
  AND ev.self_evaluation_completed_at IS NOT NULL;
