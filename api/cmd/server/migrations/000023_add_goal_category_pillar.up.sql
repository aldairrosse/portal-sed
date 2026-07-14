-- +goose Up
-- +goose StatementBegin
ALTER TABLE goal_categories
    ADD COLUMN IF NOT EXISTS pillar_id UUID NULL;
ALTER TABLE goal_categories
    ADD CONSTRAINT fk_goal_categories_pillar
    FOREIGN KEY (pillar_id) REFERENCES pillars(id)
    ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_goal_categories_pillar
    ON goal_categories (pillar_id) WHERE pillar_id IS NOT NULL;
-- +goose StatementEnd
