-- +goose Down
-- +goose StatementBegin
-- The 'role' enum value is left in place (Postgres cannot drop a single enum
-- value); dropping the column removes the only column that could reference it.
DROP INDEX IF EXISTS idx_global_goal_rules_profile;
ALTER TABLE global_goal_rules DROP COLUMN IF EXISTS profile_id;
-- +goose StatementEnd
