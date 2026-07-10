-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS goal_proposals (
    id              UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ     NOT NULL DEFAULT now(),
    goal_id         UUID            NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    requested_by    UUID            NOT NULL,
    name            TEXT            NOT NULL,
    description     TEXT            NULL,
    unit            goal_unit       NOT NULL,
    direction       TEXT            NOT NULL CHECK (direction IN ('ascendente', 'descendente')),
    weight          DOUBLE PRECISION NOT NULL CHECK (weight >= 0 AND weight <= 100),
    target_value    DOUBLE PRECISION NOT NULL CHECK (target_value > 0),
    baseline_value  DOUBLE PRECISION NULL,
    status          TEXT            NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending', 'accepted', 'rejected')),
    reviewed_by     UUID            NULL,
    reviewed_at     TIMESTAMPTZ     NULL
);

CREATE INDEX IF NOT EXISTS idx_goal_proposals_goal_id ON goal_proposals (goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_proposals_goal_status ON goal_proposals (goal_id, status);

CREATE TABLE IF NOT EXISTS goal_proposal_kpi_links (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    proposal_id UUID        NOT NULL REFERENCES goal_proposals(id) ON DELETE CASCADE,
    kpi_id      UUID        NOT NULL REFERENCES kp_is(id) ON DELETE CASCADE,
    UNIQUE(proposal_id, kpi_id)
);

CREATE INDEX IF NOT EXISTS idx_gpkl_proposal_id ON goal_proposal_kpi_links (proposal_id);
-- +goose StatementEnd
