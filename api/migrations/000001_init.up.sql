-- ============================================================================
-- SED Evaluation Platform - Initial Schema Migration
-- Version: 000001 (up)
-- Description: Creates all entity tables, enums, constraints, and indices
--              for the core data model (24 entities).
-- ============================================================================

-- +goose Up
-- +goose StatementBegin

-- --------------------------------------------------------------------------
-- Custom enum types (must be created before any table referencing them)
-- --------------------------------------------------------------------------
CREATE TYPE org_node_type AS ENUM ('corporate', 'retail');

CREATE TYPE scope_type AS ENUM ('department', 'team', 'individual');

CREATE TYPE phase AS ENUM ('asignacion', 'avance', 'cierre');

CREATE TYPE trigger_type AS ENUM ('auto', 'manual_rh');

CREATE TYPE goal_unit AS ENUM ('porcentaje', 'moneda', 'numero');

CREATE TYPE goal_state AS ENUM ('borrador', 'fijada', 'en_seguimiento', 'evaluada', 'cerrada');

CREATE TYPE axis AS ENUM ('performance', 'potential');

CREATE TYPE evaluation_state AS ENUM (
    'pendiente_asignacion',
    'pendiente_avance',
    'pendiente_evaluacion_final',
    'completada'
);

-- --------------------------------------------------------------------------
-- 1. Organizations (root tenant entity — no FK dependencies)
-- --------------------------------------------------------------------------
CREATE TABLE organizations (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    name       TEXT        NOT NULL,
    slug       TEXT        NOT NULL
);

CREATE UNIQUE INDEX idx_organizations_slug ON organizations (slug);

-- --------------------------------------------------------------------------
-- 2. Evaluation Profiles (no FK dependencies)
-- --------------------------------------------------------------------------
CREATE TABLE evaluation_profiles (
    id          UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT  NOT NULL,
    description TEXT  NULL
);

CREATE UNIQUE INDEX idx_evaluation_profiles_name ON evaluation_profiles (name);

-- --------------------------------------------------------------------------
-- 3. Org Nodes (FK → organizations; self-ref parent_id)
-- --------------------------------------------------------------------------
CREATE TABLE org_nodes (
    id              UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT now(),
    name            TEXT          NOT NULL,
    type            org_node_type NOT NULL,
    code            TEXT          NOT NULL,
    metadata        JSONB         NULL,
    organization_id UUID          NOT NULL,
    parent_id       UUID          NULL,

    CONSTRAINT fk_org_nodes_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_org_nodes_parent
        FOREIGN KEY (parent_id) REFERENCES org_nodes(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_org_nodes_org_parent ON org_nodes (organization_id, parent_id);
CREATE INDEX idx_org_nodes_org_type   ON org_nodes (organization_id, type);

-- --------------------------------------------------------------------------
-- 4. Cycles (FK → organizations)
-- --------------------------------------------------------------------------
CREATE TABLE cycles (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    year            INTEGER     NOT NULL,
    current_phase   phase       NOT NULL,
    started_at      TIMESTAMPTZ NULL,
    finished_at     TIMESTAMPTZ NULL,
    organization_id UUID        NOT NULL,

    CONSTRAINT fk_cycles_organization
        FOREIGN KEY (organization_id) REFERENCES organizations(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_cycles_org_year ON cycles (organization_id, year);
CREATE INDEX idx_cycles_current_phase ON cycles (current_phase);

-- --------------------------------------------------------------------------
-- 5. Employees (FK → org_nodes, evaluation_profiles; self-ref manager_id)
-- --------------------------------------------------------------------------
CREATE TABLE employees (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by      UUID        NOT NULL,
    updated_by      UUID        NOT NULL,
    first_name      TEXT        NOT NULL,
    last_name       TEXT        NOT NULL,
    employee_number TEXT        NOT NULL,
    email           TEXT        NOT NULL,
    is_active       BOOLEAN     NOT NULL DEFAULT true,
    org_node_id     UUID        NOT NULL,
    manager_id      UUID        NULL,
    profile_id      UUID        NOT NULL,

    CONSTRAINT fk_employees_org_node
        FOREIGN KEY (org_node_id) REFERENCES org_nodes(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_employees_manager
        FOREIGN KEY (manager_id) REFERENCES employees(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_employees_profile
        FOREIGN KEY (profile_id) REFERENCES evaluation_profiles(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_employees_email ON employees (email);
CREATE INDEX idx_employees_org_manager ON employees (org_node_id, manager_id);
CREATE INDEX idx_employees_org_profile ON employees (org_node_id, profile_id);
CREATE INDEX idx_employees_profile    ON employees (profile_id);
CREATE INDEX idx_employees_mgr_active ON employees (manager_id, is_active);

-- --------------------------------------------------------------------------
-- 6. Pillars (no FK dependencies)
-- --------------------------------------------------------------------------
CREATE TABLE pillars (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    name        TEXT        NOT NULL,
    description TEXT        NULL
);

CREATE UNIQUE INDEX idx_pillars_name ON pillars (name);

-- --------------------------------------------------------------------------
-- 7. Phase Definitions (FK → cycles)
-- --------------------------------------------------------------------------
CREATE TABLE phase_definitions (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    phase           phase       NOT NULL,
    label           TEXT        NOT NULL,
    order           INTEGER     NOT NULL CHECK (order >= 1 AND order <= 3),
    allowed_actors  JSONB       NULL,
    allowed_actions JSONB       NULL,
    blocked_actions JSONB       NULL,
    cycle_id        UUID        NOT NULL,

    CONSTRAINT fk_phase_definitions_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE
);

-- --------------------------------------------------------------------------
-- 8. Competencies (FK → pillars)
-- --------------------------------------------------------------------------
CREATE TABLE competencies (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    name        TEXT        NOT NULL,
    description TEXT        NULL,
    pillar_id   UUID        NOT NULL,

    CONSTRAINT fk_competencies_pillar
        FOREIGN KEY (pillar_id) REFERENCES pillars(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_competencies_pillar ON competencies (pillar_id);

-- --------------------------------------------------------------------------
-- 9. Level Definitions (no FK dependencies — catalog table)
-- --------------------------------------------------------------------------
CREATE TABLE level_definitions (
    id          INTEGER PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    level       INTEGER NOT NULL CHECK (level >= 1 AND level <= 5),
    label       TEXT    NOT NULL,
    description TEXT    NULL
);

CREATE UNIQUE INDEX idx_level_definitions_level ON level_definitions (level);

-- --------------------------------------------------------------------------
-- 10. 9×9 Quadrants (no FK dependencies — catalog table)
-- --------------------------------------------------------------------------
CREATE TABLE nine_box_quadrants (
    id                    UUID  PRIMARY KEY DEFAULT gen_random_uuid(),
    quadrant              INTEGER NOT NULL CHECK (quadrant >= 1 AND quadrant <= 9),
    label                 TEXT   NOT NULL,
    description           TEXT   NULL,
    color                 TEXT   NOT NULL,
    action_recommendation TEXT   NULL
);

CREATE UNIQUE INDEX idx_nine_box_quadrants_quadrant ON nine_box_quadrants (quadrant);

-- --------------------------------------------------------------------------
-- 11. 9×9 Scales (no FK dependencies — catalog table)
-- --------------------------------------------------------------------------
CREATE TABLE nine_box_scales (
    id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    axis        axis    NOT NULL,
    level       INTEGER NOT NULL CHECK (level >= 1 AND level <= 9),
    label       TEXT    NOT NULL,
    description TEXT    NULL
);

CREATE UNIQUE INDEX idx_nine_box_scales_axis_level ON nine_box_scales (axis, level);

-- --------------------------------------------------------------------------
-- 12. KPIs (no FK dependencies)
-- --------------------------------------------------------------------------
CREATE TABLE kp_is (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    name        TEXT        NOT NULL,
    unit        goal_unit   NOT NULL,
    description TEXT        NULL
);

CREATE UNIQUE INDEX idx_kp_is_name ON kp_is (name);

-- --------------------------------------------------------------------------
-- 13. Evaluator Scopes (FK → employees, cycles)
-- --------------------------------------------------------------------------
CREATE TABLE evaluator_scopes (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    scope_type   scope_type  NOT NULL,
    scope_data   JSONB       NULL,
    evaluator_id UUID        NOT NULL,
    cycle_id     UUID        NULL,

    CONSTRAINT fk_evaluator_scopes_evaluator
        FOREIGN KEY (evaluator_id) REFERENCES employees(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_evaluator_scopes_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE SET NULL
);

CREATE INDEX idx_evaluator_scopes_eval_cycle ON evaluator_scopes (evaluator_id, cycle_id);

-- --------------------------------------------------------------------------
-- 14. Goal Categories (FK → employees)
-- --------------------------------------------------------------------------
CREATE TABLE goal_categories (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID        NOT NULL,
    updated_by  UUID        NOT NULL,
    name        TEXT        NOT NULL,
    description TEXT        NULL,
    weight      DOUBLE PRECISION NOT NULL CHECK (weight >= 0 AND weight <= 100),
    employee_id UUID        NOT NULL,

    CONSTRAINT fk_goal_categories_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_goal_categories_employee      ON goal_categories (employee_id);
CREATE UNIQUE INDEX idx_goal_categories_emp_name ON goal_categories (employee_id, name);

-- --------------------------------------------------------------------------
-- 15. Goal Assignments (FK → employees, cycles)
-- --------------------------------------------------------------------------
CREATE TABLE goal_assignments (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    employee_id UUID        NOT NULL,
    cycle_id    UUID        NOT NULL,

    CONSTRAINT fk_goal_assignments_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_goal_assignments_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_goal_assignments_emp_cycle ON goal_assignments (employee_id, cycle_id);

-- --------------------------------------------------------------------------
-- 16. Phase Transitions (FK → cycles, phase_definitions)
-- --------------------------------------------------------------------------
CREATE TABLE phase_transitions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    from_phase    phase       NOT NULL,
    to_phase      phase       NOT NULL,
    trigger       trigger_type NOT NULL,
    conditions    JSONB       NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    cycle_id      UUID        NOT NULL,
    from_phase_id UUID        NOT NULL,
    to_phase_id   UUID        NOT NULL,

    CONSTRAINT fk_phase_transitions_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_phase_transitions_from_phase
        FOREIGN KEY (from_phase_id) REFERENCES phase_definitions(id)
        ON DELETE NO ACTION,

    CONSTRAINT fk_phase_transitions_to_phase
        FOREIGN KEY (to_phase_id) REFERENCES phase_definitions(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_phase_transitions_from_to ON phase_transitions (from_phase, to_phase);

-- --------------------------------------------------------------------------
-- 17. Scale Criteria (FK → competencies, pillars)
-- --------------------------------------------------------------------------
CREATE TABLE scale_criterions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    level         INTEGER     NOT NULL CHECK (level >= 1 AND level <= 5),
    description   TEXT        NOT NULL,
    competency_id UUID        NOT NULL,
    pillar_id     UUID        NOT NULL,

    CONSTRAINT fk_scale_criterions_competency
        FOREIGN KEY (competency_id) REFERENCES competencies(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_scale_criterions_pillar
        FOREIGN KEY (pillar_id) REFERENCES pillars(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_scale_criterions_cell ON scale_criterions (competency_id, pillar_id, level);

-- --------------------------------------------------------------------------
-- 18. Competency Acceptance Levels (FK → competencies, evaluation_profiles)
-- --------------------------------------------------------------------------
CREATE TABLE competency_acceptance_levels (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    level         INTEGER     NOT NULL CHECK (level >= 1 AND level <= 5),
    competency_id UUID        NOT NULL,
    profile_id    UUID        NOT NULL,

    CONSTRAINT fk_cal_competency
        FOREIGN KEY (competency_id) REFERENCES competencies(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_cal_profile
        FOREIGN KEY (profile_id) REFERENCES evaluation_profiles(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_cal_comp_profile ON competency_acceptance_levels (competency_id, profile_id);

-- --------------------------------------------------------------------------
-- 19. Goals (FK → goal_categories)
-- --------------------------------------------------------------------------
CREATE TABLE goals (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
    created_by    UUID         NOT NULL,
    updated_by    UUID         NOT NULL,
    name          TEXT         NOT NULL,
    description   TEXT         NULL,
    unit          goal_unit    NOT NULL,
    weight        DOUBLE PRECISION NOT NULL CHECK (weight >= 0 AND weight <= 100),
    target_value  DOUBLE PRECISION NOT NULL CHECK (target_value > 0),
    current_value DOUBLE PRECISION NOT NULL DEFAULT 0,
    state         goal_state   NOT NULL,
    category_id   UUID         NOT NULL,

    CONSTRAINT fk_goals_category
        FOREIGN KEY (category_id) REFERENCES goal_categories(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_goals_category   ON goals (category_id);
CREATE INDEX idx_goals_state      ON goals (state);
CREATE INDEX idx_goals_created_by ON goals (created_by);

-- --------------------------------------------------------------------------
-- 20. 9×9 Matrices (FK → cycles, employees)
-- --------------------------------------------------------------------------
CREATE TABLE nine_box_matrixes (
    id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    cycle_id     UUID        NOT NULL,
    evaluator_id UUID        NOT NULL,

    CONSTRAINT fk_nine_box_matrixes_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_nine_box_matrixes_evaluator
        FOREIGN KEY (evaluator_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_nine_box_matrixes_cycle_eval ON nine_box_matrixes (cycle_id, evaluator_id);
CREATE INDEX idx_nine_box_matrixes_evaluator ON nine_box_matrixes (evaluator_id);

-- --------------------------------------------------------------------------
-- 21. Evaluations (FK → employees, cycles)
-- --------------------------------------------------------------------------
CREATE TABLE evaluations (
    id                          UUID              PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at                  TIMESTAMPTZ       NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ       NOT NULL DEFAULT now(),
    created_by                  UUID              NOT NULL,
    updated_by                  UUID              NOT NULL,
    phase                       phase             NOT NULL,
    state                       evaluation_state  NOT NULL,
    self_evaluation_completed_at TIMESTAMPTZ      NULL,
    rh_evaluation_completed_at  TIMESTAMPTZ       NULL,
    employee_id                 UUID              NOT NULL,
    cycle_id                    UUID              NOT NULL,

    CONSTRAINT fk_evaluations_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_evaluations_cycle
        FOREIGN KEY (cycle_id) REFERENCES cycles(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_evaluations_emp_cycle   ON evaluations (employee_id, cycle_id);
CREATE INDEX idx_evaluations_cycle_state        ON evaluations (cycle_id, state);
CREATE INDEX idx_evaluations_cycle_phase        ON evaluations (cycle_id, phase);

-- --------------------------------------------------------------------------
-- 22. Goal-KPI Links (FK → goals, kp_is)
-- --------------------------------------------------------------------------
CREATE TABLE goal_kpi_links (
    id         INTEGER     PRIMARY KEY GENERATED BY DEFAULT AS IDENTITY,
    goal_id    UUID        NOT NULL,
    kpi_id     UUID        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_goal_kpi_links_goal
        FOREIGN KEY (goal_id) REFERENCES goals(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_goal_kpi_links_kpi
        FOREIGN KEY (kpi_id) REFERENCES kp_is(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_goal_kpi_links_pair ON goal_kpi_links (goal_id, kpi_id);
CREATE INDEX idx_goal_kpi_links_kpi         ON goal_kpi_links (kpi_id);

-- --------------------------------------------------------------------------
-- 23. 9×9 Entries (FK → nine_box_matrixes, employees)
-- --------------------------------------------------------------------------
CREATE TABLE nine_box_entries (
    id                UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by        UUID        NOT NULL,
    updated_by        UUID        NOT NULL,
    performance_score INTEGER     NOT NULL CHECK (performance_score >= 1 AND performance_score <= 9),
    potential_score   INTEGER     NOT NULL CHECK (potential_score >= 1 AND potential_score <= 9),
    quadrant          INTEGER     NOT NULL CHECK (quadrant >= 1 AND quadrant <= 9),
    comments          TEXT        NULL,
    matrix_id         UUID        NOT NULL,
    evaluatee_id      UUID        NOT NULL,

    CONSTRAINT fk_nine_box_entries_matrix
        FOREIGN KEY (matrix_id) REFERENCES nine_box_matrixes(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_nine_box_entries_evaluatee
        FOREIGN KEY (evaluatee_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

CREATE UNIQUE INDEX idx_nine_box_entries_matrix_eval ON nine_box_entries (matrix_id, evaluatee_id);
CREATE INDEX idx_nine_box_entries_evaluatee          ON nine_box_entries (evaluatee_id);

-- --------------------------------------------------------------------------
-- 24. Evaluation Competencies (FK → evaluations, competencies, evaluation_profiles)
-- --------------------------------------------------------------------------
CREATE TABLE evaluation_competencies (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    rating        INTEGER     NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comments      TEXT        NULL,
    evaluation_id UUID        NOT NULL,
    competency_id UUID        NOT NULL,
    profile_id    UUID        NOT NULL,

    CONSTRAINT fk_evaluation_competencies_evaluation
        FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_evaluation_competencies_competency
        FOREIGN KEY (competency_id) REFERENCES competencies(id)
        ON DELETE NO ACTION,

    CONSTRAINT fk_evaluation_competencies_profile
        FOREIGN KEY (profile_id) REFERENCES evaluation_profiles(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_eval_comp_eval_comp ON evaluation_competencies (evaluation_id, competency_id);

-- --------------------------------------------------------------------------
-- 25. Evaluation Goals (FK → evaluations, goals)
-- --------------------------------------------------------------------------
CREATE TABLE evaluation_goals (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    final_rating   INTEGER     NULL CHECK (final_rating >= 1 AND final_rating <= 5),
    final_comments TEXT        NULL,
    evaluation_id  UUID        NOT NULL,
    goal_id        UUID        NOT NULL,

    CONSTRAINT fk_evaluation_goals_evaluation
        FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_evaluation_goals_goal
        FOREIGN KEY (goal_id) REFERENCES goals(id)
        ON DELETE NO ACTION
);

CREATE UNIQUE INDEX idx_eval_goals_eval_goal ON evaluation_goals (evaluation_id, goal_id);

-- +goose StatementEnd
