-- +goose Down
-- +goose StatementBegin
UPDATE goal_comments SET goal_id = '00000000-0000-0000-0000-000000000000' WHERE goal_id IS NULL;
ALTER TABLE goal_comments ALTER COLUMN goal_id SET NOT NULL;
-- +goose StatementEnd
