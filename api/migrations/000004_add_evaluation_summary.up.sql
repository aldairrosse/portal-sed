-- ============================================================================
-- SED Evaluation Platform - Evaluation Summary Materialized View
-- Version: 000004 (up)
-- Description: Creates the evaluation_summary materialized view for dashboard
--              aggregation of evaluation counts by cycle and state.
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

CREATE MATERIALIZED VIEW IF NOT EXISTS evaluation_summary AS
SELECT cycle_id, state, COUNT(1) as count
FROM evaluations
GROUP BY cycle_id, state
WITH DATA;

CREATE UNIQUE INDEX IF NOT EXISTS idx_evaluation_summary_cycle_state
ON evaluation_summary (cycle_id, state);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP MATERIALIZED VIEW IF EXISTS evaluation_summary;

-- +goose StatementEnd
