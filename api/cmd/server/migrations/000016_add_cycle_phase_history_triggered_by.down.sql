-- ============================================================================
-- SED Evaluation Platform - Revert cycle_phase_history column fix
-- Version: 000016 (down)
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

ALTER TABLE cycle_phase_history DROP COLUMN IF EXISTS triggered_by;
ALTER TABLE cycle_phase_history ADD COLUMN trigger trigger_type NOT NULL DEFAULT 'manual_rh';
ALTER TABLE cycle_phase_history ADD COLUMN actor_id UUID NULL;

-- +goose StatementEnd
