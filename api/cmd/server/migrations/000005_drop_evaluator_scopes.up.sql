-- ============================================================================
-- SED Evaluation Platform - Drop evaluator_scopes table
-- Version: 000005 (up)
-- Description: Drops the redundant evaluator_scopes table after verifying it
--              is empty. Scope is now derived from the org tree + employee's
--              org_node_id at query time.
--
-- PRECHECK: The DO block below RAISES EXCEPTION if any rows exist in the
--           table. If this migration blocks in production, verify that no
--           application code writes to evaluator_scopes (none does — see
--           explore obs #356), then empty the table manually and retry.
--
-- OPERATOR ACTION REQUIRED before running in prod: verify backups exist
-- and are restorable. See runbook: <link if known, else note in PR description>
--
-- NOTE: The scope_type ENUM created in 000001_init.up.sql survives DROP CASCADE.
--       It is not dropped here because ENUMs are schema-level objects, not
--       table-level. The down migration references the existing ENUM by name.
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

DO $$
BEGIN
    IF (SELECT COUNT(*) FROM evaluator_scopes) > 0 THEN
        RAISE EXCEPTION 'evaluator_scopes is not empty (% rows). Blocking migration.',
            (SELECT COUNT(*) FROM evaluator_scopes);
    END IF;

    DROP TABLE IF EXISTS evaluator_scopes CASCADE;
END $$;

-- +goose StatementEnd
