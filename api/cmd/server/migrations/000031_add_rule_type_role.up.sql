-- +goose Up
-- +goose StatementBegin
-- Add 'role' rule type to global_goal_rules. ALTER TYPE ... ADD VALUE cannot be
-- used inside a transaction that also USES the new value, so this migration is
-- statement-only (no DML referencing 'role') — safe to run in a transaction.
DO $$ BEGIN
    ALTER TYPE rule_type ADD VALUE 'role';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

ALTER TABLE global_goal_rules ADD COLUMN IF NOT EXISTS profile_id UUID REFERENCES evaluation_profiles(id);
CREATE INDEX IF NOT EXISTS idx_global_goal_rules_profile ON global_goal_rules (profile_id);
-- +goose StatementEnd
