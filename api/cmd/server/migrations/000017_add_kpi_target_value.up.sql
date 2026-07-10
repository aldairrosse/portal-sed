-- ============================================================================
-- SED Evaluation Platform - Add target_value to KPIs
-- Version: 000017 (up)
-- Description: Adds 'binario' to goal_unit enum and target_value column to kp_is
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

-- Add 'binario' to the goal_unit enum
ALTER TYPE goal_unit ADD VALUE IF NOT EXISTS 'binario';

-- Add target_value column to kp_is
ALTER TABLE kp_is ADD COLUMN target_value DOUBLE PRECISION NULL;

-- +goose StatementEnd
