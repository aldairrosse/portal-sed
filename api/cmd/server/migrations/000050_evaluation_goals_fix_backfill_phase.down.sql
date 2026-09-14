-- +goose Down
-- +goose StatementBegin
-- NOOP: corrective data move from 000050 cannot be reliably reversed
-- (source column was nulled). Columns themselves belong to 000048.
SELECT 1;
-- +goose StatementEnd
