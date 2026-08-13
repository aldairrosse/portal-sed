-- +goose Up
ALTER TABLE global_goal_rules ADD COLUMN IF NOT EXISTS default_target DOUBLE PRECISION NOT NULL DEFAULT 100 CHECK (default_target > 0);
