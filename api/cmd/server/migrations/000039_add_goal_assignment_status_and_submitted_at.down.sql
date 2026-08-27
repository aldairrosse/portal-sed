-- +goose Down
-- +goose StatementBegin
ALTER TABLE goal_assignments
    DROP COLUMN IF EXISTS submitted_at,
    DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS goal_assignment_status;
-- +goose StatementEnd
