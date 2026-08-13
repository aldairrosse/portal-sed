-- +goose Down
-- +goose StatementBegin
ALTER TABLE nine_box_entries
    ADD COLUMN performance_score INTEGER NOT NULL CHECK (performance_score >= 1 AND performance_score <= 9),
    ADD COLUMN potential_score   INTEGER NOT NULL CHECK (potential_score >= 1 AND potential_score <= 9);
-- +goose StatementEnd
