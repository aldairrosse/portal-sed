-- +goose Down
-- +goose StatementBegin
-- NOTE: PostgreSQL does not support ALTER TYPE ... DROP VALUE, so the
-- 'en_progreso' enum value is intentionally left in place on downgrade.
SELECT 1;
-- +goose StatementEnd
