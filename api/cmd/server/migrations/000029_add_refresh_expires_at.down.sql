-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS refresh_expires_at;
-- +goose StatementEnd