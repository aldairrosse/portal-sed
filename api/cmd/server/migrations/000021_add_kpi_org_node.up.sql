-- ============================================================================
-- SED Evaluation Platform - Scope KPIs by direction (org_node)
-- Version: 000021 (up)
-- Description: Adds org_node_id to kp_is so KPIs are scoped to the user's
--              main direction (direct child of the org root).
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

ALTER TABLE kp_is
    ADD COLUMN IF NOT EXISTS org_node_id UUID NULL
    REFERENCES org_nodes(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_kpis_org_node ON kp_is (org_node_id);

-- +goose StatementEnd
