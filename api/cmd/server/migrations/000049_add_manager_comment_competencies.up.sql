-- +goose Up
-- +goose StatementBegin
-- Diferencia comentarios de competencias por rol: jefe escribe manager_comment,
-- RH escribe comments; rating 1-5 compartido en columna rating (fila source=rh).
ALTER TABLE evaluation_competencies
    ADD COLUMN IF NOT EXISTS manager_comment TEXT;
-- +goose StatementEnd
