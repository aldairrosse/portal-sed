-- +goose Up
-- +goose StatementBegin
-- Gap 2 (sed-medio-cierre-9box): allow one rating per (evaluation, competency, source)
-- so self and rh ratings coexist. Source column already added in 000033.
DROP INDEX IF EXISTS idx_eval_comp_eval_comp;
CREATE UNIQUE INDEX IF NOT EXISTS idx_eval_comp_eval_comp_source
    ON evaluation_competencies (evaluation_id, competency_id, source);
-- +goose StatementEnd
