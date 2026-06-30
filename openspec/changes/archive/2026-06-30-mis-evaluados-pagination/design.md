# Design: Mis Evaluados — Server-Side Pagination

## Technical Approach

Replicate the established OFFSET/LIMIT pagination pattern from "Evaluaciones RH" (`rhEvaluadosStore` + `rh/evaluaciones`) into "Mis Evaluados". The change flows through three layers: new paginated repo methods → service method with `hasMore` computation → handler query-param parsing on the existing route. Frontend mirrors the reference: a new store module with `init(empId)` + `load()`/`next()`/`prev()`/`search()`, and a page rewrite consuming the store's getters.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Repo location | Add `ListByManagerPaginated` + `CountByManager` to `EmployeeRepo` | `ListByManager` exists there; paginated variant follows `ListWithProfiles`/`CountWithProfiles` pattern in same file |
| Profile JOIN in repo | Use `LEFT JOIN evaluation_profiles` + `scanEmployeeRowsWithProfile` | Existing helper already scans profile fields; avoids new scan function |
| Self-exclusion | `AND e.id != $managerID` in SQL WHERE | Single-condition filter in SQL is cheaper and simpler than post-fetch filtering |
| Service method | Add `GetMyEvaluateesPaginated` to `EvaluateeService` interface | Keeps existing `GetMyEvaluatees` intact for any other callers; new method has distinct signature |
| Handler approach | Modify existing `GetMyEvaluatees` handler to always parse `offset`/`limit`/`q` | Route path stays identical; defaults to `offset=0, limit=50` when params absent — backward compatible |
| Store `init(empId)` | Store `employeeId` at module level, set once from page `onMount` | `rhEvaluadosStore` calls `/employees` (no path param); this store needs `{empId}` in URL — must capture it before first `load()` |
| Employee ID source | `getSession().user?.employeeId` from `session.svelte.ts` | `devContext.getProfile()` returns `profileId` (role), NOT `employeeId` (UUID). Session has the correct field |

## Data Flow

```
Browser                          API                              DB
  │                               │                                │
  │ GET /employees/{empId}/       │                                │
  │   evaluatees?offset=0&        │                                │
  │   limit=50&q=mar              │                                │
  │──────────────────────────────>│                                │
  │                               │ Handler: parse empId,          │
  │                               │   offset, limit, q             │
  │                               │   → evaluateeSvc               │
  │                               │     .GetMyEvaluateesPaginated  │
  │                               │───────────┐                    │
  │                               │           │ empRepo            │
  │                               │           │.ListByManager      │
  │                               │           │  Paginated(ctx,    │
  │                               │           │  mgrID, q, 0, 50)  │
  │                               │           │───────────────────>│
  │                               │           │                    │
  │                               │           │ empRepo            │
  │                               │           │.CountByManager     │
  │                               │           │  (ctx, mgrID, q)   │
  │                               │           │───────────────────>│
  │                               │           │<───────────────────│
  │                               │           │<───────────────────│
  │                               │<──────────┘                    │
  │                               │ hasMore = offset+len < total   │
  │                               │ → EmployeeListResponse         │
  │<──────────────────────────────│                                │
  │ { data: [...], meta:          │                                │
  │   { hasMore, total, ... } }   │                                │
```

## File Changes

### PR 1 — Backend

| File | Action | Description |
|------|--------|-------------|
| `api/internal/repository/org/employee_repo.go` | Modify | Add `ListByManagerPaginated(ctx, managerID, query, offset, limit)` and `CountByManager(ctx, managerID, query)` |
| `api/internal/service/org/evaluatee_service.go` | Modify | Add `GetMyEvaluateesPaginated` to interface + implementation |
| `api/internal/handler/org/org_handler.go` | Modify | Update `GetMyEvaluatees` to parse `offset`, `limit`, `q` query params |

**Repo — SQL for `ListByManagerPaginated`:**

```sql
SELECT e.id, e.created_at, e.updated_at, e.first_name, e.last_name, e.email,
       e.employee_number, e.is_active, e.org_node_id, e.manager_id, e.profile_id,
       COALESCE(ep.name, '') as profile_name,
       COALESCE(ep.description, '') as profile_description, e.job_title
FROM employees e
LEFT JOIN evaluation_profiles ep ON e.profile_id = ep.id
WHERE e.manager_id = $1
  AND e.id != $1          -- exclude self
  AND e.is_active = true
  AND (e.first_name ILIKE $2 OR e.last_name ILIKE $2
       OR e.email ILIKE $2 OR e.employee_number ILIKE $2)  -- when query != ""
ORDER BY e.last_name, e.first_name, e.id
LIMIT $N OFFSET $N+1
```

Parameters: `$1=managerID`, `$2=%query%` (omit ILIKE clause when query is empty), then `$LIMIT`, `$OFFSET`.

**Repo — SQL for `CountByManager`:**

```sql
SELECT COUNT(*)
FROM employees e
WHERE e.manager_id = $1
  AND e.id != $1
  AND e.is_active = true
  [AND (e.first_name ILIKE $2 OR ...)]  -- same conditional ILIKE
```

**Service signature:**

```go
GetMyEvaluateesPaginated(ctx context.Context, evaluatorID string, query string, offset, limit int) (*org.EmployeeListResponse, error)
```

Logic: parse UUID → verify evaluator exists → call `ListByManagerPaginated` + `CountByManager` → set `hasMore = offset + len(rows) < total`.

**Handler changes to `GetMyEvaluatees`:**

```go
// Parse query params (same pattern as ListEmployees)
q := r.URL.Query().Get("q")
limit := 50   // default
// parse + clamp 1-200
offset := 0   // default
// parse + clamp >= 0

result, err := h.evaluateeSvc.GetMyEvaluateesPaginated(r.Context(), empID, q, offset, limit)
```

### PR 2 — Frontend Store

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/stores/misEvaluadosStore.svelte.ts` | Create | New paginated store mirroring `rhEvaluadosStore` |

**Module structure** (exact mirror of `rhEvaluadosStore.svelte.ts`):

```ts
// $state fields
let items = $state<EmployeeListItemExtended[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let hasMore = $state(false);
let hasPrev = $state(false);
let apiTotal = $state(0);
let currentPage = $state(0);
let currentQ: string | undefined;
let employeeId = '';  // set by init()

const PAGE_SIZE = 50;

// Exports (same names as rhEvaluadosStore)
export function getItems() / isLoading() / getError() / hasMoreItems()
export function hasPrevItems() / getCurrentPage() / getTotalCount() / getSearchQuery()
export function init(empId: string)  // ← NEW: captures manager's employeeId
export async function load()         // GET /employees/{employeeId}/evaluatees
export async function next() / prev()
export function search(query: string)  // 300ms debounce
```

**Key difference from `rhEvaluadosStore`:** `load()` calls `client.GET('/employees/{empId}/evaluatees', { params: { path: { empId: employeeId }, query: { q: currentQ, offset, limit: PAGE_SIZE } } })` instead of `GET /employees`.

### PR 3 — Frontend UI

| File | Action | Description |
|------|--------|-------------|
| `web/src/routes/mis-evaluados/+page.svelte` | Modify | Full rewrite following `rh/evaluaciones/+page.svelte` pattern |

**Key differences from `rh/evaluaciones/+page.svelte`:**

| Aspect | rh/evaluaciones | mis-evaluados |
|--------|-----------------|---------------|
| Store import | `rhEvaluadosStore.svelte` | `misEvaluadosStore.svelte` |
| `onMount` | `load()` | `init(employeeId); load()` |
| Employee ID | Not needed (global `/employees`) | `getSession().user?.employeeId` |
| Table mode | `mode="rh"` | `mode="manager"` |
| Detail viewerMode | `viewerMode="rh"` | `viewerMode="manager"` |
| Phase subtitle | "...de todos los empleados" | "...de tu equipo" |
| Empty state text | "Sin empleados para mostrar" | "Sin evaluados para mostrar" |
| Title icon | `ClipboardList` | `Users` |
| Removed imports | N/A | Remove `orgHierarchyStore`, `goalsStore`, `getAssignmentsByProfile` |

**Component structure** (identical layout to reference):

```
<div class="flex flex-col gap-6">
  Header: Users icon + "Mis evaluados" + phaseDescription
  Search bar: input + "Viendo X de Y" counter + Prev/Next buttons
  {#if loading && items.length === 0}  → PageSkeleton
  {:else if storeError}                → ErrorState
  {:else if items.length === 0}        → "Sin evaluados para mostrar"
  {:else}                              → EmployeeEvaluationTable mode="manager"
    {#snippet detail()}                → EmployeeEvaluationDetail viewerMode="manager"
{/if}
```

## Integration Points

1. **Store → Backend**: `load()` calls `GET /employees/{employeeId}/evaluatees?offset=N&limit=50&q=...` via `openapi-fetch` client
2. **Page → Store**: `onMount` calls `init(getSession().user?.employeeId)` then `load()`. All reactive bindings use `$derived(getter())` pattern
3. **Team scoping (self-exclusion)**: Handled server-side via `AND e.id != $managerID` in SQL — no client-side filtering needed
4. **Search**: `input oninput` → `store.search(value)` → 300ms debounce → resets `currentPage=0` → `load()`

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | Repo `ListByManagerPaginated` / `CountByManager` | Table-driven tests with test DB — verify self-exclusion, ILIKE search, pagination bounds |
| Unit (Go) | Service `GetMyEvaluateesPaginated` | Mock repo — verify `hasMore` calculation, UUID validation, evaluator-not-found error |
| Unit (Go) | Handler param parsing | Existing `org_handler_test.go` pattern — verify limit clamping, offset default, invalid UUID → 400 |
| E2E | Full pagination flow | Navigate to mis-evaluados, verify skeleton → table, search filters results, next/prev paginate |

## Open Questions

None — the pattern is fully established by the reference implementation.
