-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_goal_categories_pillar;
ALTER TABLE goal_categories DROP CONSTRAINT IF EXISTS fk_goal_categories_pillar;
ALTER TABLE goal_categories DROP COLUMN IF EXISTS pillar_id;
-- +goose StatementEnd
