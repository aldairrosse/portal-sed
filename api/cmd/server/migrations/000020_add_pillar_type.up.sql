-- +goose Up
-- +goose StatementBegin
ALTER TABLE pillars
    ADD COLUMN IF NOT EXISTS type TEXT NOT NULL DEFAULT 'competencias'
    CHECK (type IN ('competencias', 'metas'));
CREATE INDEX IF NOT EXISTS idx_pillars_type ON pillars (type);
-- +goose StatementEnd
