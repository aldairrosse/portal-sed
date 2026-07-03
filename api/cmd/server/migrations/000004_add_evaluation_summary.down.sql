-- +goose Down
-- +goose StatementBegin

DROP MATERIALIZED VIEW IF EXISTS evaluation_summary;

-- +goose StatementEnd
