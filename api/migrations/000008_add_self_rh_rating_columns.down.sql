-- Rollback: remove self_rating and rh_rating columns.
ALTER TABLE evaluation_competencies DROP COLUMN IF EXISTS self_rating;
ALTER TABLE evaluation_competencies DROP COLUMN IF EXISTS rh_rating;
