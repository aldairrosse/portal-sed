-- +goose Down
-- +goose StatementBegin
-- NOOP: keep updated_at column/data; dropping would lose audit info.
-- Column predates Ent expectation on some DBs, so do not drop on rollback.
SELECT 1;
-- +goose StatementEnd
