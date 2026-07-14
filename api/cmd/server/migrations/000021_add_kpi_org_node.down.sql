-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_kpis_org_node;
ALTER TABLE kp_is DROP COLUMN IF EXISTS org_node_id;

-- +goose StatementEnd
