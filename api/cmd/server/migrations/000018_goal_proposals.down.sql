-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS goal_proposal_kpi_links;
DROP TABLE IF EXISTS goal_proposals;
-- +goose StatementEnd
