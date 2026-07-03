# Apply Progress — PR4 (T4.1–T4.8)

**Change:** wire-api-replace-mocks
**Date:** 2026-06-11
**Status:** ✅ Complete

## Completed Tasks

### T4.1 — goal/routes.go ✅
- Changed `NewRouter(handler)` → `NewRouter(handler, authSvc *authsvc.AuthService)`
- Replaced `AuthPlaceholder` → `RequireAuth(authSvc)`
- Added `RequirePermission(auth.PermGoalRead)` on GET groups
- Added `RequireAnyPermission(goal write perms...)` on POST/PUT/DELETE/PATCH groups
- Updated `NewSubRouter` signature

### T4.2 — evaluation/routes.go ✅
- Changed `NewRouter(handler)` → `NewRouter(handler, authSvc)`
- Changed `RegisterRoutes(r, handler)` → `RegisterRoutes(r, handler, authSvc)`
- Added `RequireAuth(authSvc)` at router level
- Added `RequirePermission(auth.PermEvalRH)` on RH evaluation endpoints
- Added `RequirePermission(auth.PermEval9x9)` on nine-box write endpoints

### T4.3 — competency/routes.go ✅
- Added `AuthSvc *authsvc.AuthService` to `Dependencies` struct
- Replaced `AuthPlaceholder` → `RequireAuth(deps.AuthSvc)`
- Added `RequirePermission(auth.PermCompetencyWrite)` on write endpoints
- READ endpoints keep only `RequireAuth`

### T4.4 — cycle/routes.go ✅
- Changed `NewRouter(handler)` → `NewRouter(handler, authSvc)`
- Changed `RegisterRoutes(r, handler)` → `RegisterRoutes(r, handler, authSvc)`
- Added `RequireAuth(authSvc)` at router level

### T4.5 — org/routes.go ✅
- Changed `NewRouter(handler)` → `NewRouter(handler, authSvc)`
- Changed `RegisterRoutes(r, handler)` → `RegisterRoutes(r, handler, authSvc)`
- Added `RequireAuth(authSvc)` at router level

### T4.6 — main.go ✅
- Removed `apiV1.Use(middleware.AuthPlaceholder)`
- Removed unused `middleware` import
- Passed `authSvc` to all `NewRouter` and `RegisterRoutes` calls
- Added `AuthSvc: authSvc` to competency `Dependencies`

### T4.7 — Dev auth service ✅
- Created `api/internal/auth/dev/service.go`
- Preset users: dev-rh, dev-jefe, dev-colaborador with appropriate roles/permissions
- `GetUserByEmail()` for lookup
- `CreateDevSession()` creates real DB sessions via `SessionStore.Create()`

### T4.8 — SSOAdapter interface ✅
- Created `api/internal/auth/sso/adapter.go`
- `SSOAdapter` interface with `ValidateToken`, `GetEndSessionURL`, `GetUserFromToken`
- `SSOUser` struct for external user data
- `noopAdapter` placeholder implementation
- Documentation comments for OIDC/SAML integration guide

## Files Changed

| File | Action |
|------|--------|
| `api/internal/handler/goal/routes.go` | Modified — new signature, RequireAuth, per-endpoint permissions |
| `api/internal/handler/evaluation/routes.go` | Modified — new signature, RequireAuth, RH/9x9 permissions |
| `api/internal/handler/competency/routes.go` | Modified — new dependency, RequireAuth, write permissions |
| `api/internal/handler/cycle/routes.go` | Modified — new signature, RequireAuth |
| `api/internal/handler/org/routes.go` | Modified — new signature, RequireAuth |
| `api/cmd/server/main.go` | Modified — pass authSvc, remove AuthPlaceholder |
| `api/internal/middleware/auth.go` | Modified — nil-safe RequireAuth for test mode |
| `api/internal/auth/dev/service.go` | **New** — dev auth service with preset users |
| `api/internal/auth/sso/adapter.go` | **New** — SSOAdapter interface |
| `api/internal/handler/cycle/cycle_handler_test.go` | Modified — pass nil authSvc for test mode |
| `api/integration/server_test.go` | Modified — pass authSvc through, remove AuthPlaceholder |
| `api/integration/routes_test.go` | Modified — comment update |

## Build & Test Results
- `go build ./...` — ✅ passes
- `go vet ./...` — ✅ passes  
- All handler tests (6 packages) — ✅ pass
- All service/repo tests — ✅ pass

## Remaining for this PR
None — PR4 is complete.

## Dependency notes
- PR4 is independent of PR2/PR3 (no shared files)
- Can merge in parallel with PR2 and PR3
- Integration test requires real PostgreSQL (skip without DATABASE_URL)
