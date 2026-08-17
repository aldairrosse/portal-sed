-- +goose Down
-- +goose StatementBegin
ALTER TABLE goals ADD CONSTRAINT goals_target_value_check CHECK (target_value > 0);
-- +goose StatementEnd
