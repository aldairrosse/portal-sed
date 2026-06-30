# Tasks: Mis Evaluados — Server-Side Pagination

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~500 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (backend) → PR 2 (store) → PR 3 (UI) |
| Delivery strategy | auto-chain |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Paginated evaluatees endpoint | PR 1 | Base: main; repo+service+handler+tests |
| 2 | misEvaluadosStore | PR 2 | Base: main; depends on PR 1 endpoint |
| 3 | Mis Evaluados page rewrite | PR 3 | Base: main; depends on PR 2 store |

## Phase 1: Backend — Paginated Evaluatees Endpoint (PR 1)

- [ ] 1.1 Add `ListByManagerPaginated(ctx, managerID, query, offset, limit)` to `api/internal/repository/org/employee_repo.go` — SQL with `LEFT JOIN evaluation_profiles`, `WHERE manager_id=$1 AND e.id!=$1 AND is_active=true`, ILIKE search on name/email/employeeNumber, `ORDER BY last_name, first_name, e.id LIMIT OFFSET`
- [ ] 1.2 Add `CountByManager(ctx, managerID, query)` to `api/internal/repository/org/employee_repo.go` — same WHERE + self-exclusion, returns `int`
- [ ] 1.3 Add `GetMyEvaluateesPaginated(ctx, evaluatorID, query string, offset, limit int) (*EmployeeListResponse, error)` to `EvaluateeService` interface and implementation in `api/internal/service/org/evaluatee_service.go` — UUID validation, evaluator-exists check, calls `ListByManagerPaginated`+`CountByManager`, computes `hasMore = offset+len(rows) < total`
- [ ] 1.4 Update `GetMyEvaluatees` handler in `api/internal/handler/org/org_handler.go` — parse `offset`/`limit`/`q` query params (pattern: `ListEmployees` handler), clamp limit 1-200, default 50, delegate to `GetMyEvaluateesPaginated`
- [ ] 1.5 Update `mockEvaluateeService` in `api/internal/handler/org/org_handler_test.go` — add `getMyEvaluateesPaginatedFunc` field + method, update existing `GetMyEvaluatees` mock to match new handler signature
- [ ] 1.6 Add unit tests for repo: `TestListByManagerPaginated` / `TestCountByManager` in `api/internal/repository/org/employee_repo_test.go` — verify self-exclusion, ILIKE search, pagination bounds
- [ ] 1.7 Add unit tests for service: `TestEvaluateeService_GetMyEvaluateesPaginated` in `api/internal/service/org/evaluatee_service_test.go` — verify hasMore=false at last page, unknown UUID error, evaluator-not-found error
- [ ] 1.8 Add handler tests for query param parsing: update `TestGetMyEvaluatees_*` in `api/internal/handler/org/org_handler_test.go` — verify limit clamping, default values, invalid UUID → 400

## Phase 2: Frontend Store (PR 2)

- [x] 2.1 Create `web/src/lib/stores/misEvaluadosStore.svelte.ts` — mirror `rhEvaluadosStore.svelte.ts` structure: `$state` fields (`items`, `loading`, `error`, `hasMore`, `hasPrev`, `apiTotal`, `currentPage`, `currentQ`), `PAGE_SIZE=50`
- [x] 2.2 Export getters: `getItems`, `isLoading`, `getError`, `hasMoreItems`, `hasPrevItems`, `getCurrentPage`, `getTotalCount`, `getSearchQuery` — same names as `rhEvaluadosStore`
- [x] 2.3 Implement `init(employeeId: string)` — stores `employeeId` for all subsequent API calls
- [x] 2.4 Implement `load()` — calls `client.GET('/employees/{empId}/evaluatees', { params: { path: { empId: employeeId }, query: { q: currentQ, offset: currentPage * PAGE_SIZE, limit: PAGE_SIZE } } })`, sets `items`/`hasMore`/`apiTotal`/`error`
- [x] 2.5 Implement `next()`/`prev()` — guard with `hasMore`/`currentPage>0`, increment/decrement `currentPage`, call `load()`
- [x] 2.6 Implement `search(query)` — 300ms debounce, reset `currentPage=0`, set `currentQ`, call `load()`

## Phase 3: Frontend UI — Mis Evaluados Page Rewrite (PR 3)

- [x] 3.1 Rewrite `web/src/routes/mis-evaluados/+page.svelte` — follow `rh/evaluaciones/+page.svelte` structure: header with `Users` icon + "Mis evaluados" + phase-dependent subtitle, search bar with result counter, Prev/Next buttons with page number
- [x] 3.2 Import store + wire reactive bindings — `$derived` getters from `misEvaluadosStore.svelte.ts`, `onMount` calls `init(getSession().user?.employeeId)` then `load()`
- [x] 3.3 Implement search bar — `input-bordered input-sm max-w-sm`, always enabled, `oninput` → `search(val)`, result counter shows "Viendo X de Y empleado(s)" or "Viendo X resultado(s)"
- [x] 3.4 Implement pagination controls — `btn-outline btn-xs` with `ChevronLeft`/`ChevronRight`, disabled when `!hasPrev||loading` / `!hasMore||loading`, display "Pág. {currentPage+1}"
- [x] 3.5 Conditionally render skeleton/error/empty/table — `PageSkeleton` on `loading && items.length===0`, `ErrorState` on `storeError`, "Sin evaluados para mostrar" on empty, `EmployeeEvaluationTable` with `mode="rh"` and `EmployeeEvaluationDetail` with `viewerMode="manager"`
- [x] 3.6 Remove unused imports — `orgHierarchyStore`, `goalsStore`, `getProfile`, `getAssignmentsByProfile`, `getChildren`
