# Dev Bar — Impersonation (development only)

## Goal

Add a minimal development-only impersonation bar to speed up manual QA across roles/profiles. Visible only when `ENV=development` (or `APP_ENV=development`). Fully isolated — removable by deleting `api/internal/auth/dev/` and the conditional mount in `api/cmd/server/main.go` plus the frontend mount.

## Non-goals

- Production impersonation or admin UI.
- RBAC bypass in production.
- Persisting impersonation state beyond session cookie.

## Scope

| Layer | Change |
|-------|--------|
| API `api/internal/auth/dev/` | New handlers: `GET /api/v1/dev/status`, `GET /api/v1/dev/employees`, `POST /api/v1/dev/impersonate`; guard `ENV`/`APP_ENV != development → 404` |
| API mount | Conditional `r.Mount("/api/v1/dev", dev.Routes(...))` in `main.go` only when dev enabled |
| DB | Read-only query on `employees` + `evaluation_profiles`: `id, first_name, last_name, profile_name, email` |
| Session | Reuse `auth.SessionStore.Create` + `session_token` httpOnly cookie; reuse `ProfileNameToRole` |
| Frontend `DevBar.svelte` | Mounted only if `GET /api/v1/dev/status` returns 200; selector → `POST /api/v1/dev/impersonate` → `ensureSession()` + reload |

## Isolation

- All dev logic lives in `api/internal/auth/dev/` (handler, routes, service already there). No changes to `internal/handler/auth` or middleware.
- Prod: handlers return 404; router not mounted if guard false.
- Frontend: conditional mount after status probe; no import when disabled.

## Validation

- Production env returns 404 for all `/api/v1/dev/*`.
- Selector lists real DB employees (active).
- Impersonate sets valid `session_token` cookie that passes `RequireAuth`.

## Precondition

- `employees` table seeded; `auth.SessionStore` available.
