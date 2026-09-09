-- +goose Down
-- +goose StatementBegin
-- Data consolidation is not reversible (original medio-anio rows are indistinguishable from avance).
SELECT 1;
-- +goose StatementEnd
