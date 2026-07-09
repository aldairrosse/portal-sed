-- ============================================================================
-- SED Evaluation Platform - Fix cycle_phase_history column naming
-- Version: 000016 (up)
-- Description: Replaces `trigger` + `actor_id` with `triggered_by` to match
--              the Go repository code and the original design spec.
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

-- Add the column the code actually uses
ALTER TABLE cycle_phase_history ADD COLUMN triggered_by UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';

-- Remove the columns that were never used by application code
ALTER TABLE cycle_phase_history DROP COLUMN IF EXISTS trigger;
ALTER TABLE cycle_phase_history DROP COLUMN IF EXISTS actor_id;

-- +goose StatementEnd
