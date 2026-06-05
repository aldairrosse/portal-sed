-- ============================================================================
-- SED Evaluation Platform - Session Tables (rollback)
-- Version: 000003 (down)
-- Description: Drops session management tables created in the up migration.
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_sessions_expires;
DROP INDEX IF EXISTS idx_sessions_employee;
DROP INDEX IF EXISTS idx_sessions_token_hash;

DROP TABLE IF EXISTS sessions CASCADE;

-- +goose StatementEnd
