-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_goal_comments_goal_phase;
ALTER TABLE goal_comments DROP CONSTRAINT IF EXISTS chk_goal_comments_phase;
ALTER TABLE goal_comments DROP COLUMN IF EXISTS phase;
-- +goose StatementEnd
