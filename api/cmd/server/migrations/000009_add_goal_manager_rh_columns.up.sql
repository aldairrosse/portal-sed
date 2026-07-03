-- Add manager_comment and rh_assessment columns to evaluation_goals.
-- These are used by the per-goal state endpoints (PUT /evaluations/{id}/goal-state)
-- and manager comment endpoint (PUT /evaluations/{id}/goal-comments).
-- The existing final_rating and final_comments columns are retained.

ALTER TABLE evaluation_goals
  ADD COLUMN manager_comment TEXT NULL;

ALTER TABLE evaluation_goals
  ADD COLUMN rh_assessment TEXT NULL;
