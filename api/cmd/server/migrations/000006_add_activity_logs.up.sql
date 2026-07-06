-- ============================================================================
-- SED Evaluation Platform - Add Activity Logs Table
-- Version: 000006 (up)
-- Description: Creates the activity_logs table for append-only audit trail
--              of employee actions, indexed by (employee_id, created_at DESC)
--              for the /perfil timeline query.
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS activity_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    employee_id UUID        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    module      TEXT        NOT NULL,
    metadata    JSONB       NULL,

    CONSTRAINT fk_activity_logs_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_emp_created
    ON activity_logs (employee_id, created_at DESC);

-- +goose StatementEnd
