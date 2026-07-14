-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_goal_comments_assignment;
ALTER TABLE goal_comments DROP CONSTRAINT IF EXISTS fk_goal_comments_assignment;
ALTER TABLE goal_comments DROP COLUMN IF EXISTS assignment_id;
-- +goose StatementEnd
