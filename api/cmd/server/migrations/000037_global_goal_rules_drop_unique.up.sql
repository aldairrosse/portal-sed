-- +goose Up
-- +goose StatementBegin
-- A goal may hold multiple rules of the same rule_type (e.g. one department rule
-- per department). Drop the UNIQUE (goal_id, rule_type) constraint added in 000030.
ALTER TABLE global_goal_rules DROP CONSTRAINT IF EXISTS global_goal_rules_goal_id_rule_type_key;
-- +goose StatementEnd
