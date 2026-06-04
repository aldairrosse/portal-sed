-- ============================================================================
-- SED Evaluation Platform - Initial Schema Migration (rollback)
-- Version: 000001 (down)
-- Description: Drops all entities, tables, and enums created in the up migration.
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse dependency order (children before parents)
DROP TABLE IF EXISTS evaluation_goals         CASCADE;
DROP TABLE IF EXISTS evaluation_competencies  CASCADE;
DROP TABLE IF EXISTS nine_box_entries         CASCADE;
DROP TABLE IF EXISTS goal_kpi_links           CASCADE;
DROP TABLE IF EXISTS evaluations              CASCADE;
DROP TABLE IF EXISTS nine_box_matrixes        CASCADE;
DROP TABLE IF EXISTS goals                    CASCADE;
DROP TABLE IF EXISTS competency_acceptance_levels CASCADE;
DROP TABLE IF EXISTS scale_criterions         CASCADE;
DROP TABLE IF EXISTS phase_transitions        CASCADE;
DROP TABLE IF EXISTS goal_assignments         CASCADE;
DROP TABLE IF EXISTS goal_categories          CASCADE;
DROP TABLE IF EXISTS evaluator_scopes         CASCADE;
DROP TABLE IF EXISTS kp_is                    CASCADE;
DROP TABLE IF EXISTS nine_box_scales          CASCADE;
DROP TABLE IF EXISTS nine_box_quadrants       CASCADE;
DROP TABLE IF EXISTS level_definitions        CASCADE;
DROP TABLE IF EXISTS competencies             CASCADE;
DROP TABLE IF EXISTS phase_definitions        CASCADE;
DROP TABLE IF EXISTS pillars                  CASCADE;
DROP TABLE IF EXISTS employees                CASCADE;
DROP TABLE IF EXISTS cycles                   CASCADE;
DROP TABLE IF EXISTS org_nodes                CASCADE;
DROP TABLE IF EXISTS evaluation_profiles      CASCADE;
DROP TABLE IF EXISTS organizations            CASCADE;

-- Drop custom enum types
DROP TYPE IF EXISTS evaluation_state;
DROP TYPE IF EXISTS axis;
DROP TYPE IF EXISTS goal_state;
DROP TYPE IF EXISTS goal_unit;
DROP TYPE IF EXISTS trigger_type;
DROP TYPE IF EXISTS phase;
DROP TYPE IF EXISTS scope_type;
DROP TYPE IF EXISTS org_node_type;

-- +goose StatementEnd
