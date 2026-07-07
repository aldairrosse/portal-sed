-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS change_requests (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    entity_type   TEXT        NOT NULL,
    entity_id     TEXT        NOT NULL,
    requested_by  TEXT        NOT NULL,
    status        TEXT        NOT NULL DEFAULT 'pending',
    approved_by   TEXT        NULL,
    approved_at   TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS idx_change_requests_entity ON change_requests (entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_change_requests_status ON change_requests (status);
-- +goose StatementEnd
