-- ============================================================================
-- SED Evaluation Platform - Add Activity Logs Table
-- Version: 000006 (down)
-- Description: Drops the activity_logs table and its composite index.
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_activity_logs_emp_created;
DROP TABLE IF EXISTS activity_logs;

-- +goose StatementEnd
