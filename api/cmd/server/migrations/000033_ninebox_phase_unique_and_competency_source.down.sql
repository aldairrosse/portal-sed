-- +goose Down
-- +goose StatementBegin
ALTER TABLE evaluation_competencies DROP COLUMN IF EXISTS source;
DROP TYPE IF EXISTS evaluation_competency_source;
CREATE UNIQUE INDEX IF NOT EXISTS idx_nine_box_matrixes_cycle_eval
    ON nine_box_matrixes (cycle_id, evaluator_id);
-- +goose StatementEnd
