-- +goose Up
-- +goose StatementBegin
-- Gap 1 (sed-medio-cierre-9box): track which phase a goal comment belongs to.
-- Historical rows predate phase tracking; backfill to 'cierre'.
ALTER TABLE goal_comments
    ADD COLUMN IF NOT EXISTS phase TEXT NOT NULL DEFAULT 'cierre'
    CONSTRAINT chk_goal_comments_phase CHECK (phase IN ('asignacion', 'avance', 'cierre'));

UPDATE goal_comments SET phase = 'cierre' WHERE phase IS NULL OR phase NOT IN ('asignacion', 'avance', 'cierre');

CREATE INDEX IF NOT EXISTS idx_goal_comments_goal_phase ON goal_comments (goal_id, phase);
-- +goose StatementEnd
