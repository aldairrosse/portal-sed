-- +goose Up
-- +goose StatementBegin
ALTER TABLE goal_comments ADD COLUMN category_id UUID NULL;
ALTER TABLE goal_comments DROP CONSTRAINT fk_goal_comments_goal;
ALTER TABLE goal_comments ADD CONSTRAINT fk_goal_comments_goal
    FOREIGN KEY (goal_id) REFERENCES goals(id) ON DELETE CASCADE;
ALTER TABLE goal_comments ADD CONSTRAINT fk_goal_comments_category
    FOREIGN KEY (category_id) REFERENCES goal_categories(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_goal_comments_category ON goal_comments (category_id) WHERE category_id IS NOT NULL;
-- +goose StatementEnd
