-- ============================================================================
-- SED Evaluation Platform - Add target_value to KPIs
-- Version: 000017 (down)
-- Description: Reverts target_value column from kp_is
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

-- Remove target_value column from kp_is
ALTER TABLE kp_is DROP COLUMN IF EXISTS target_value;

-- Note: PostgreSQL does not support removing values from an enum.
-- The 'binario' value will remain in goal_unit but be unused.

-- +goose StatementEnd
