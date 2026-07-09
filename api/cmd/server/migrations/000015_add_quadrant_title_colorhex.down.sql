-- +goose Down
-- +goose StatementBegin
ALTER TABLE nine_box_quadrants DROP COLUMN IF EXISTS color_hex;
ALTER TABLE nine_box_quadrants DROP COLUMN IF EXISTS title;
-- +goose StatementEnd
