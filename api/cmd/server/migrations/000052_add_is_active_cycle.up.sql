-- +goose Up
-- +goose StatementBegin
-- Active-cycle flag: single active cycle per organization (partial unique index).
ALTER TABLE cycles ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT false;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_active_per_org
	ON cycles(organization_id) WHERE is_active;

-- Backfill: latest cycle per org (year DESC, updated_at DESC) becomes active,
-- but only for orgs without an active cycle yet (idempotent, keeps manual flags).
UPDATE cycles c
SET is_active = true
FROM (
	SELECT DISTINCT ON (organization_id) id, organization_id
	FROM cycles
	ORDER BY organization_id, year DESC, updated_at DESC
) latest
WHERE c.id = latest.id
	AND NOT EXISTS (
		SELECT 1 FROM cycles e
		WHERE e.organization_id = latest.organization_id
			AND e.is_active
	);
-- +goose StatementEnd
