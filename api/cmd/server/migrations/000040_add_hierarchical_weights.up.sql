-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS cycle_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id UUID NOT NULL UNIQUE REFERENCES cycles(id) ON DELETE CASCADE,
    g_weight DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (g_weight >= 0 AND g_weight <= 100),
    p_weight DOUBLE PRECISION NOT NULL DEFAULT 100 CHECK (p_weight >= 0 AND p_weight <= 100),
    CHECK (g_weight + p_weight = 100)
);
CREATE TABLE IF NOT EXISTS team_weight_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cycle_id UUID NOT NULL REFERENCES cycles(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES org_nodes(id) ON DELETE CASCADE,
    j_weight DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (j_weight >= 0 AND j_weight <= 100),
    pj_weight DOUBLE PRECISION NOT NULL DEFAULT 100 CHECK (pj_weight >= 0 AND pj_weight <= 100),
    UNIQUE (cycle_id, team_id),
    CHECK (j_weight + pj_weight = 100)
);
-- +goose StatementEnd
