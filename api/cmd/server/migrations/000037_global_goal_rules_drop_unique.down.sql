-- +goose Down
-- +goose StatementBegin
ALTER TABLE global_goal_rules ADD CONSTRAINT global_goal_rules_goal_id_rule_type_key UNIQUE (goal_id, rule_type);
-- +goose StatementEnd
