-- ============================================================================
-- SED Evaluation Platform - Restore evaluator_scopes table
-- Version: 000005 (down)
-- Description: Recreates the evaluator_scopes table referencing the existing
--              scope_type ENUM (which survived DROP CASCADE in the up migration).
-- ============================================================================

-- +goose Down
-- +goose StatementBegin

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

-- +goose StatementEnd
