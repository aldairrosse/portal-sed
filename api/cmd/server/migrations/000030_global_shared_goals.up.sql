-- +goose Up
-- +goose StatementBegin
-- New enum types for goal type and kind
DO $$ BEGIN
    CREATE TYPE goal_type AS ENUM ('personal', 'global', 'shared');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE goal_kind_type AS ENUM ('qualitative', 'quantitative');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

DO $$ BEGIN
    CREATE TYPE rule_type AS ENUM ('department', 'min_direct_reports');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- Extend goals table with type and goal_kind
ALTER TABLE goals ADD COLUMN IF NOT EXISTS "type" goal_type NOT NULL DEFAULT 'personal';
ALTER TABLE goals ADD COLUMN IF NOT EXISTS goal_kind goal_kind_type NULL;

-- Goal templates (reusable forms)
CREATE TABLE IF NOT EXISTS goal_templates (
    id UUID PRIMARY KEY,
    name VARCHAR NOT NULL,
    description TEXT,
    unit goal_unit NOT NULL,
    direction VARCHAR NOT NULL DEFAULT 'ascendente',
    target_value DOUBLE PRECISION NOT NULL,
    goal_kind goal_kind_type NOT NULL,
    is_public BOOLEAN NOT NULL DEFAULT false,
    created_by UUID NOT NULL REFERENCES employees(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Template-KPI links (N:M)
CREATE TABLE IF NOT EXISTS goal_template_kpi_links (
    id UUID PRIMARY KEY,
    template_id UUID NOT NULL REFERENCES goal_templates(id) ON DELETE CASCADE,
    kpi_id UUID NOT NULL REFERENCES kp_is(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, kpi_id)
);

-- Global goal assignments (mass assignment to employees)
CREATE TABLE IF NOT EXISTS global_goal_assignments (
    id UUID PRIMARY KEY,
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    weight DOUBLE PRECISION NOT NULL CHECK (weight >= 0 AND weight <= 100),
    target_value DOUBLE PRECISION NOT NULL,
    baseline_value DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (goal_id, employee_id)
);

-- Global goal rules (department / min direct reports)
CREATE TABLE IF NOT EXISTS global_goal_rules (
    id UUID PRIMARY KEY,
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    rule_type rule_type NOT NULL,
    department_id UUID REFERENCES org_nodes(id),
    min_direct_reports INT,
    default_weight DOUBLE PRECISION NOT NULL CHECK (default_weight >= 0 AND default_weight <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (goal_id, rule_type)
);

-- Shared goal groups (boss/director group goals)
CREATE TABLE IF NOT EXISTS shared_goal_groups (
    id UUID PRIMARY KEY,
    goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES employees(id),
    name VARCHAR NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (goal_id)
);

-- Shared goal members (group members with variable weight)
CREATE TABLE IF NOT EXISTS shared_goal_members (
    id UUID PRIMARY KEY,
    group_id UUID NOT NULL REFERENCES shared_goal_groups(id) ON DELETE CASCADE,
    employee_id UUID NOT NULL REFERENCES employees(id),
    weight DOUBLE PRECISION NOT NULL CHECK (weight >= 0 AND weight <= 100),
    target_value DOUBLE PRECISION NOT NULL,
    baseline_value DOUBLE PRECISION,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (group_id, employee_id)
);

-- Indexes for new tables
CREATE INDEX IF NOT EXISTS idx_global_goal_assignments_employee ON global_goal_assignments (employee_id);
CREATE INDEX IF NOT EXISTS idx_global_goal_rules_goal ON global_goal_rules (goal_id);
CREATE INDEX IF NOT EXISTS idx_shared_goal_members_employee ON shared_goal_members (employee_id);
CREATE INDEX IF NOT EXISTS idx_shared_goal_groups_creator ON shared_goal_groups (created_by);
CREATE INDEX IF NOT EXISTS idx_goal_templates_creator ON goal_templates (created_by);
CREATE INDEX IF NOT EXISTS idx_goal_templates_public ON goal_templates (is_public);
-- +goose StatementEnd
