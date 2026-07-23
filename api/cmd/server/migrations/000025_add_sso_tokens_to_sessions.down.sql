-- +goose Down
-- +goose StatementBegin

ALTER TABLE sessions DROP COLUMN IF EXISTS id_token;
ALTER TABLE sessions DROP COLUMN IF EXISTS access_token;
ALTER TABLE sessions DROP COLUMN IF EXISTS refresh_token;

-- +goose StatementEnd
