-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_eval_comp_eval_comp_source;
CREATE UNIQUE INDEX IF NOT EXISTS idx_eval_comp_eval_comp
    ON evaluation_competencies (evaluation_id, competency_id);
-- +goose StatementEnd
