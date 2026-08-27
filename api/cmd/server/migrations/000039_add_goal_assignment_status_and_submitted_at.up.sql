-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'goal_assignment_status') THEN
        CREATE TYPE goal_assignment_status AS ENUM ('borrador', 'enviada');
    END IF;
END $$;

ALTER TABLE goal_assignments
    ADD COLUMN IF NOT EXISTS status goal_assignment_status NOT NULL DEFAULT 'borrador',
    ADD COLUMN IF NOT EXISTS submitted_at TIMESTAMPTZ;
-- +goose StatementEnd
