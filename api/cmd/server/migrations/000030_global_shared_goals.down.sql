-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS shared_goal_members;
DROP TABLE IF EXISTS shared_goal_groups;
DROP TABLE IF EXISTS global_goal_rules;
DROP TABLE IF EXISTS global_goal_assignments;
DROP TABLE IF EXISTS goal_template_kpi_links;
DROP TABLE IF EXISTS goal_templates;
ALTER TABLE goals DROP COLUMN IF EXISTS goal_kind;
ALTER TABLE goals DROP COLUMN IF EXISTS "type";
DROP TYPE IF EXISTS rule_type;
DROP TYPE IF EXISTS goal_kind_type;
DROP TYPE IF EXISTS goal_type;
-- +goose StatementEnd
