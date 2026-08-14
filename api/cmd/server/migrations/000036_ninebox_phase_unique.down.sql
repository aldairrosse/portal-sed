-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_nine_box_matrixes_cycle_eval_phase;
-- +goose StatementEnd
