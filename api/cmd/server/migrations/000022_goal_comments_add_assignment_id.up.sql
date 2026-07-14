-- +goose Up
-- +goose StatementBegin
ALTER TABLE goal_comments ADD COLUMN assignment_id UUID NULL;
ALTER TABLE goal_comments ADD CONSTRAINT fk_goal_comments_assignment
    FOREIGN KEY (assignment_id) REFERENCES goal_assignments(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_goal_comments_assignment
    ON goal_comments (assignment_id) WHERE assignment_id IS NOT NULL;
-- +goose StatementEnd
