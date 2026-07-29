-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN IF EXISTS acr;
ALTER TABLE sessions DROP COLUMN IF EXISTS requires_2fa;
-- +goose StatementEnd
