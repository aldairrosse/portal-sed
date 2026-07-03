-- ============================================================================
-- SED Evaluation Platform - Add version columns and aux tables
-- Version: 000002 (down)
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_evaluations_state;
DROP INDEX IF EXISTS idx_employees_name_search;
DROP INDEX IF EXISTS idx_goals_category_weight;
DROP INDEX IF EXISTS idx_phase_transitions_from_phase;
DROP INDEX IF EXISTS idx_org_nodes_path;

ALTER TABLE org_nodes DROP COLUMN IF EXISTS path;
DROP EXTENSION IF EXISTS ltree;

DROP TABLE IF EXISTS ninebox_entry_versions;
DROP TABLE IF EXISTS evaluation_versions;
DROP TABLE IF EXISTS cycle_phase_history;

ALTER TABLE scale_criterions DROP COLUMN IF EXISTS version;
ALTER TABLE competencies DROP COLUMN IF EXISTS version;
ALTER TABLE pillars DROP COLUMN IF EXISTS version;
ALTER TABLE nine_box_entries DROP COLUMN IF EXISTS version;
ALTER TABLE evaluations DROP COLUMN IF EXISTS version;
ALTER TABLE org_nodes DROP COLUMN IF EXISTS version;
ALTER TABLE goals DROP COLUMN IF EXISTS version;
ALTER TABLE cycles DROP COLUMN IF EXISTS version;

-- +goose StatementEnd
