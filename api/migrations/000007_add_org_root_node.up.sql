-- Add root_node_id to organizations so the tree root is explicit
ALTER TABLE organizations
  ADD COLUMN root_node_id UUID REFERENCES org_nodes(id) ON DELETE SET NULL;

CREATE INDEX idx_organizations_root_node_id ON organizations(root_node_id);
