-- +goose Down
ALTER TABLE global_goal_rules DROP COLUMN IF EXISTS default_target;
