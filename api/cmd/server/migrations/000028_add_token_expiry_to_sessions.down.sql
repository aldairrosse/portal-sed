-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS token_expires_at;
-- +goose StatementEnd
