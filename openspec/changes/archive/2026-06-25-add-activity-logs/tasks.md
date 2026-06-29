# Tasks: Activity Logs

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~450–550 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend Foundation (DB + schema + service + endpoint) | PR 1 | Base branch: main. Self-contained; endpoint returns 200 with empty array. |
| 2 | Frontend Integration (/perfil page) | PR 2 | Base: main after PR 1 merge. Replaces fixture with real API. |
| 3 | Wire LogActivity into domain handlers | PR 3 | Base: main after PR1+PR2 merge. Logs 3–5 key operations. |

## PR 1: Backend Foundation

- [x] 1.1 Create `api/migrations/000006_add_activity_logs.up.sql` with DDL (table + index)
- [x] 1.2 Create `api/migrations/000006_add_activity_logs.down.sql`
- [x] 1.3 Create Ent schema `api/internal/schema/activitylog.go` (TimeMixin, fields, edge, index)
- [x] 1.4 Add `edge.To("activity_logs", ActivityLog.Type)` in `api/internal/schema/employee.go`
- [x] 1.5 Regenerate Ent client: `cd api && go run entgo.io/ent/cmd/entc generate --target ./internal ./internal/schema`
- [x] 1.6 Create repository `api/internal/repository/activity/activity_repo.go` (Create + ListByEmployee via Ent queries)
- [x] 1.7 Create service `api/internal/service/activity/activity_service.go` (LogActivity + ListByEmployee with ownership check)
- [x] 1.8 Create OpenAPI spec `api/openapi/activity-logs.yaml` (GET endpoint + ActivityLog schema)
- [x] 1.9 Generate TS types: `cd web && pnpm openapi-typescript ../api/openapi/activity-logs.yaml -o src/lib/api/schemas/activity-logs.d.ts`
- [x] 1.10 Create handler `api/internal/handler/activity/activity_handler.go` (ListActivityLogs)
- [x] 1.11 Create routes `api/internal/handler/activity/routes.go` (RequireAuth + RateLimit)
- [x] 1.12 Wire in `cmd/server/main.go`: repo → service → handler + RegisterRoutes
- [x] 1.13 Build: `cd api && go build ./...` — no errors
- [x] 1.14 **PR 1 Verify**: `go test ./...` — all pass; no tests exist for new packages yet

## PR 2: Frontend Integration

- [x] 2.1 Add `ActivityLogsPaths` import + merge into `AppPaths` in `web/src/lib/api/client.ts`
- [x] 2.2 Create store `web/src/lib/stores/activityLogStore.svelte.ts` ($state logs/loading/error + loadActivityLogs)
- [x] 2.3 Update `web/src/routes/perfil/+page.svelte`: remove fixture import, add store + onMount, `profileId` → `employeeId`
- [x] 2.4 Run `cd web && pnpm run check` — verify no TS errors (0 new errors)
- [x] 2.5 Run `cd web && pnpm run lint` — verify no lint errors (0 new errors)
- [x] 2.6 **PR 2 Verify**: navigate `/perfil`, verify timeline renders (empty if no logs); confirm no fixture import remains

## PR 3: Wire LogActivity into Handlers

- [x] 3.1 Identify 3–5 domain operations to log initially (e.g. evaluation submitted, goal created, KPI updated)
- [x] 3.2 Add `LogActivity()` calls in those handlers after successful operations
- [x] 3.3 Run `cd api && go build ./...` — verify no errors
- [x] 3.4 **PR 3 Verify**: `go test ./... -short` all pass; build clean
