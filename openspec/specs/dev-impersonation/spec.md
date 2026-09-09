# dev-impersonation Specification

## Purpose
Gate the dev-only impersonation bar and its endpoints to development environments.

## Requirements

### Requirement: Dev-only impersonation bar SHALL be gated by development environment

The system SHALL expose `GET /api/v1/dev/status`, `GET /api/v1/dev/employees` and `POST /api/v1/dev/impersonate` only when `ENV`/`APP_ENV` is `development`, and SHALL return 404 otherwise. The frontend `DevBar.svelte` SHALL render only after the status probe returns 200.

#### Scenario: Production guard returns 404

- WHEN `ENV` is `production` and a client calls `GET /api/v1/dev/status`
- THEN the API returns 404 and the dev bar renders nothing
