# Archive Report: Add Activity Logs

**Change**: add-activity-logs
**Archived**: 2026-06-25
**Previous location**: `openspec/changes/add-activity-logs/`
**Archive location**: `openspec/changes/archive/2026-06-25-add-activity-logs/`
**Verification**: PASS WITH WARNINGS — READY TO ARCHIVE

## Change Summary

Reemplazó el fixture `activity-logs.json` en `/perfil` por registros reales de actividad persistidos en PostgreSQL. Se agregó tabla `activity_logs` con migración `000006`, esquema Ent, utility `LogActivity()` como único punto de inserción, endpoint REST `GET /api/v1/employees/{employeeId}/activity-logs`, y consumo frontend desde la vista de perfil. 5 handlers de dominio (`evaluation`, `competency`, `goal` ×2, `cycle`) invocan `LogActivity()` tras operaciones exitosas.

## Capabilities Added

| Capability | Type | Description |
|------------|------|-------------|
| `activity-logs` | New | Registro, persistencia y consulta de eventos de actividad del empleado: tabla BD, schema Ent, utility de registro, endpoint REST y consumo frontend |
| `ui-shell` | Modified | `/perfil` deja de usar fixture y consume API real para el timeline de actividad |

## Deliverables

### PR 1 — Backend Foundation
Database migration (`000006`), Ent schema, repository, service, OpenAPI spec, handler, routes, wiring.

### PR 2 — Frontend Integration
API client type merge, activity log store (`$state` triplet), perfil page update (remove fixture, call API, profileId → employeeId).

### PR 3 — Wire LogActivity into Handlers
5 domain handlers log activity after successful operations: `evaluation_started`, `evaluation_completed`, `goal_progress`, `kpi_updated`, `cycle_configured`.

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `activity-logs` | Created | Full spec promoted from delta — 6 requirements, 7 acceptance criteria |

Source of truth: `openspec/specs/activity-logs/spec.md`

## Files Created

| File | Lines | Description |
|------|-------|-------------|
| `api/migrations/000006_add_activity_logs.up.sql` | 25 | DDL: tabla `activity_logs` + índice compuesto `(employee_id, created_at DESC)` |
| `api/migrations/000006_add_activity_logs.down.sql` | 10 | DROP TABLE + DROP INDEX |
| `api/internal/schema/activitylog.go` | 52 | Ent schema: TimeMixin, fields, edge to Employee, index |
| `api/internal/repository/activity/activity_repo.go` | 43 | Repository: `Create()`, `ListByEmployee()` con raw SQL |
| `api/internal/service/activity/activity_service.go` | 77 | Service: `LogActivity()`, `ListByEmployee()` con auth check |
| `api/internal/handler/activity/activity_handler.go` | 90 | Handler: `ListActivityLogs()` — GET endpoint |
| `api/internal/handler/activity/routes.go` | 33 | `RegisterRoutes()` con RequireAuth + rate limit |
| `api/openapi/activity-logs.yaml` | 50 | OpenAPI 3.1 spec del endpoint |
| `web/src/lib/api/schemas/activity-logs.d.ts` | 89 | Tipos TS generados desde OpenAPI |
| `web/src/lib/stores/activityLogStore.svelte.ts` | 62 | Store: `$state` logs/loading/error + `loadActivityLogs()` |

**Total new files**: 10 | **Total new lines**: 531

## Files Modified

| File | Description | Location |
|------|-------------|----------|
| `api/cmd/server/main.go` | Wiring: repo → service → handler + `RegisterActivityRoutes` | Backend DI |
| `api/internal/schema/employee.go` | Added `edge.To("activity_logs", ActivityLog.Type)` | Ent edge |
| `web/src/lib/api/client.ts` | Merged `ActivityLogsPaths` into `AppPaths` type | API client |
| `web/src/routes/perfil/+page.svelte` | Removed fixture import, added store + onMount, `profileId` → `employeeId` | Perfil page |

## Task Completion

All 23 tasks verified complete across 3 PRs:
- **PR 1** (Backend Foundation): 13/13 tasks ✅
- **PR 2** (Frontend Integration): 6/6 tasks ✅
- **PR 3** (Wire LogActivity): 4/4 tasks ✅

## Key Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| PK type | UUID (consistencia) | `gen_random_uuid()` como todas las tablas existentes |
| Action field | TEXT (no enum) | Nuevas acciones sin migración |
| Metadata | JSONB NULL | Contexto variable sin schema changes |
| Index | `(employee_id, created_at DESC)` | Cubre el 100% de queries del timeline |
| Authorization | Service layer (ownership check) | Un solo endpoint no justifica middleware dedicado |
| Limit | 50 hardcoded, max 100 via param | Spec dice "últimos 50, sin scroll infinito" |
| Frontend store | Module-level `$state` | Consistencia con stores existentes |

## Verification Notes

- **API build**: ✅ `go build ./...` — clean
- **API tests**: ✅ `go test ./... -short` — all pass
- **Web check**: ✅ 73 errors — all pre-existing (zero new)
- **Web lint**: ✅ 48 errors — all pre-existing (zero new)
- **CRITICAL issues**: None resolved (0 found)
- **Warnings**: 1 (metadata optional but not Nillable — non-blocking)

## Lessons Learned

1. **Migration headers matter**: The initial `down.sql` used `-- +goose Up` instead of `-- +goose Down` — caught and fixed during verification.
2. **`updated_at` in append-only tables**: The `TimeMixin` adds `updated_at` even though activity logs never update. Considered dead weight but acceptable for consistency.
3. **16 action types**: Documented catalog ensures cross-handler consistency. The frontend maps 15 of 16 icons (`goal_overview` uses fallback). Future change can add the missing icon.
