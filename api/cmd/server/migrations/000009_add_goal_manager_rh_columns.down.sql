-- Revert: drop manager_comment and rh_assessment columns from evaluation_goals.

ALTER TABLE evaluation_goals DROP COLUMN IF EXISTS manager_comment;
ALTER TABLE evaluation_goals DROP COLUMN IF EXISTS rh_assessment;
