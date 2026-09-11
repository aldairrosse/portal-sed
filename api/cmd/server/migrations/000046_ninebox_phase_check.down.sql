-- +goose Down
-- +goose StatementBegin
-- Data migration is not reversible (deleted asignacion matrices and merged
-- duplicate avance phase_definitions cannot be reconstructed).
SELECT 1;
-- +goose StatementEnd
