-- +goose Down
-- +goose StatementBegin
ALTER TABLE evaluation_competencies
    DROP COLUMN IF EXISTS manager_comment;
-- +goose StatementEnd
