-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    cname text;
BEGIN
    SELECT conname INTO cname
    FROM pg_constraint
    WHERE conrelid = 'goals'::regclass AND contype = 'c'
      AND pg_get_constraintdef(oid) ILIKE '%target_value%';
    IF cname IS NOT NULL THEN
        EXECUTE format('ALTER TABLE goals DROP CONSTRAINT %I', cname);
    END IF;
END $$;
-- +goose StatementEnd
