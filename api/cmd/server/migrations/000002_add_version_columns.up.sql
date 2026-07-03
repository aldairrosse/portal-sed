-- ============================================================================
-- SED Evaluation Platform - Add version columns and aux tables
-- Version: 000002 (up)
-- Description: Adds optimistic locking version columns, audit history tables,
--              and aux tables required by C2–C6 APIs.
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

-- --------------------------------------------------------------------------
-- Version columns for optimistic locking
-- --------------------------------------------------------------------------

ALTER TABLE cycles ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE goals ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE org_nodes ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE evaluations ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE nine_box_entries ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE pillars ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE competencies ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

ALTER TABLE scale_criterions ADD COLUMN version INTEGER NOT NULL DEFAULT 1;

-- --------------------------------------------------------------------------
-- Cycle phase history (audit trail for phase transitions)
-- --------------------------------------------------------------------------

CREATE TABLE cycle_phase_history (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id        UUID        NOT NULL,
    from_phase      phase       NOT NULL,
    to_phase        phase       NOT NULL,
    trigger         trigger_type NOT NULL,
    actor_id        UUID        NULL,
    reason          TEXT        NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_cph_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_cph_cycle ON cycle_phase_history (cycle_id);

-- --------------------------------------------------------------------------
-- Evaluation versions (separate table for optimistic locking)
-- --------------------------------------------------------------------------

CREATE TABLE evaluation_versions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    evaluation_id   UUID        NOT NULL,
    version         INTEGER     NOT NULL DEFAULT 1,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_ev_evaluation
        FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_ev_evaluation UNIQUE (evaluation_id)
);

-- --------------------------------------------------------------------------
-- NineBox entry versions (separate table for optimistic locking)
-- --------------------------------------------------------------------------

CREATE TABLE ninebox_entry_versions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    entry_id        UUID        NOT NULL,
    version         INTEGER     NOT NULL DEFAULT 1,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_nbev_entry
        FOREIGN KEY (entry_id) REFERENCES nine_box_entries(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_nbev_entry UNIQUE (entry_id)
);

-- --------------------------------------------------------------------------
-- Org node ltree path (for efficient tree traversal)
-- --------------------------------------------------------------------------

CREATE EXTENSION IF NOT EXISTS ltree;

ALTER TABLE org_nodes ADD COLUMN path ltree NULL;

CREATE INDEX idx_org_nodes_path ON org_nodes USING gist (path);

-- --------------------------------------------------------------------------
-- Additional indexes for C2–C6 query patterns
-- --------------------------------------------------------------------------

-- C2: cycle transitions by phase
CREATE INDEX idx_phase_transitions_from_phase ON phase_transitions (from_phase);

-- C4: goal weight validation
CREATE INDEX idx_goals_category_weight ON goals (category_id, weight);

-- C5: employee search
CREATE INDEX idx_employees_name_search ON employees USING gin (
    to_tsvector('spanish', first_name || ' ' || last_name || ' ' || email)
);

-- C6: evaluation dashboard
CREATE INDEX idx_evaluations_state ON evaluations (state);

-- +goose StatementEnd

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
