-- ============================================================================
-- SED Evaluation Platform - Session Tables
-- Version: 000003 (up)
-- Description: Creates session management tables for authentication (C7).
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

-- --------------------------------------------------------------------------
-- Sessions
-- Stores authenticated user sessions with token hashing.
-- Tokens are SHA-256 hashed before storage; only the hash is persisted.
-- --------------------------------------------------------------------------

CREATE TABLE sessions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    employee_id     UUID        NOT NULL,
    token_hash      TEXT        NOT NULL,
    ip_address      INET        NULL,
    user_agent      TEXT        NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_active_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_revoked      BOOLEAN     NOT NULL DEFAULT false,

    CONSTRAINT fk_sessions_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

-- Token hash lookup (login validation is the most frequent session query)
CREATE UNIQUE INDEX idx_sessions_token_hash ON sessions (token_hash);

-- List sessions for an employee (logout all, audit)
CREATE INDEX idx_sessions_employee ON sessions (employee_id);

-- Filter active, non-expired sessions (cleanup jobs, expiry checks)
CREATE INDEX idx_sessions_expires ON sessions (expires_at) WHERE NOT is_revoked;

-- +goose StatementEnd
