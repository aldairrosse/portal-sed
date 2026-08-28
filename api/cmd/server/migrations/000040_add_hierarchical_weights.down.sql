-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_weight_configs;
DROP TABLE IF EXISTS cycle_configs;
-- +goose StatementEnd
