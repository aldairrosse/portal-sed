-- +goose Up
-- +goose StatementBegin

DO $$ BEGIN
  ALTER TABLE organizations
    ADD COLUMN root_node_id UUID REFERENCES org_nodes(id) ON DELETE SET NULL;
EXCEPTION WHEN duplicate_column THEN NULL; END $$;

CREATE INDEX IF NOT EXISTS idx_organizations_root_node_id ON organizations(root_node_id);

-- +goose StatementEnd
