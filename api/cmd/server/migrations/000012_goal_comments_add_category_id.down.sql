-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_goal_comments_category;
ALTER TABLE goal_comments DROP CONSTRAINT IF EXISTS fk_goal_comments_category;
ALTER TABLE goal_comments DROP COLUMN IF EXISTS category_id;
-- +goose StatementEnd
