DROP INDEX IF EXISTS idx_organizations_root_node_id;
ALTER TABLE organizations DROP COLUMN IF EXISTS root_node_id;
