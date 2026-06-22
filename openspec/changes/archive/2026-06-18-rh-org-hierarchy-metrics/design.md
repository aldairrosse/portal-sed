# Design: rh-org-hierarchy-metrics — RRHH Hierarchy Metrics View

## Technical Approach

Reuse existing `OrgHierarchyTree` + `TreeNode` components and `orgHierarchyStore` traversal helpers. Add a new `/rh/jerarquia` route with RRHH profile guard, a `rhHierarchyStore` that computes per-node aggregates from existing raw fixtures (`goals.json`, `assignments.json`, `rh-evaluations.json`) via pure functions, and a detail panel with phase-conditional metrics cards + employee table.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Metrics computation | Pure functions over raw fixtures (not pre-computed `metrics.json`) | Spec requires pure, testable, importable functions. Raw fixtures already exist; pre-computing adds a sync burden. `metrics.json` shape documents the API response contract, not a file. |
| Store pattern | Module-level `$state` + exported functions (matches `orgHierarchyStore`) | Codebase convention. No class needed — state is simple (selectedNodeId + derived metrics). |
| Data source for employees in subtree | `orgHierarchyStore.getDescendants(nodeId)` → extract IDs → include the node itself | `getDescendants` returns children only; spec requires the area manager included. Use `[nodeId, ...descendants.map(n => n.id)]` (same as `getScopeIds`). |
| Goal-to-employee mapping | `assignments.json` → `goalIds` → join with `goals.json` for `progress`/`targetValue` | Assignments link employees to goal IDs. Goals have `progress` and `targetValue`. |
| Phase detection | `devContext.getPhase()` → `$derived` metric type | Existing pattern from Directors Jerarquía. `CyclePhase` = `inicio-anio` | `medio-anio` | `fin-anio`. |
| Backend endpoint (future) | `GET /org-nodes/{nodeId}/metrics?cyclePhase=...` | Spec stub. SQL aggregation via `ltree` path filter + JOINs on goals/evaluations. RBAC: `PermOrgRead` (already exists for RH). |

## Data Flow

```
User clicks OrgNode in tree
        │
        ▼
rhHierarchyStore.selectNode(nodeId)
        │
        ├── orgHierarchyStore.getScopeIds(nodeId) → employeeIds[]
        │
        ├── computeAreaMetrics(employeeIds, assignments, goals, rhEvaluations, phase)
        │     ├── filter assignments by employeeIds → relevant goalIds
        │     ├── join goals → progress/targetValue per goal
        │     ├── compute: avgProgress, completedCount, pendingCount (medio-anio)
        │     ├── compute: avgRating from rhEvaluations (fin-anio)
        │     └── collect employee list (id, name, profileId) from tree nodes
        │
        ▼
$derived reactive bind → UI cards + table
```

## Store Design: `rhHierarchyStore.svelte.ts`

```ts
// State
let selectedNodeId = $state<string>('');
let metrics = $state<AreaMetrics | null>(null);
let employeeList = $state<EmployeeRow[]>([]);

// Pure functions (exported, testable)
export function computeAreaProgress(
  employeeIds: string[],
  assignments: EmployeeAssignment[],
  goals: Goal[]
): { avgProgress: number; completed: number; pending: number }

export function computeAreaRating(
  employeeIds: string[],
  rhEvaluations: RhEvaluation[]
): { avgRating: number | null; ratingsCount: number }

export function buildEmployeeList(
  employeeIds: string[],
  nodes: OrgNode[]
): EmployeeRow[]

// Reactive selectors
export function selectNode(nodeId: string): void
export function getMetrics(): AreaMetrics | null
export function getEmployeeList(): EmployeeRow[]
export function getSelectedNodeId(): string
```

**State shape — `AreaMetrics`:**
```ts
interface AreaMetrics {
  nodeId: string;
  employeeCount: number;
  employeesWithGoals: number;
  avgProgress: number | null;    // medio-anio only
  completedGoals: number;
  pendingGoals: number;
  avgRating: number | null;      // fin-anio only
  ratingsCount: number;
}

interface EmployeeRow {
  id: string;
  name: string;
  position: string;   // profileId → PROFILE_LABELS
  profile: string;
}
```

## Fixture Design

No new fixture file. The store imports existing fixtures directly:
- `$lib/fixtures/goals/goals.json` → `Goal[]`
- `$lib/fixtures/goals/assignments.json` → `EmployeeAssignment[]`
- `$lib/fixtures/evaluations/rh-evaluations.json` → `RhEvaluation[]`
- `orgHierarchyStore` → tree nodes for employee names/profiles

**`metrics.json` shape** (documents future API response, not a file):
```json
{
  "nodeId": "emp-director-01",
  "cyclePhase": "medio-anio",
  "employeeCount": 12,
  "employeesWithGoals": 8,
  "avgGoalProgress": 67.5,
  "completedGoals": 14,
  "pendingGoals": 6,
  "avgEvaluationRating": null,
  "ratingsCount": 0,
  "employees": [
    { "id": "emp-jefe-01", "name": "Carlos Rodríguez Pérez", "position": "Jefe", "profile": "jefe" }
  ]
}
```

## Route & Guard

**New file:** `web/src/routes/rh/jerarquia/+page.svelte`

```ts
// Profile guard
const profile = $derived(getProfile());
const isAuthorized = $derived(profile === 'rh');

// Tree root — full corporate tree for RRHH
const treeRoot = $derived(getRoot());

// Phase-conditional metric type
const phase = $derived(getPhase());
const metricType = $derived(
  phase === 'medio-anio' ? 'progress' :
  phase === 'fin-anio'   ? 'rating'   : 'unavailable'
);
```

**Layout:** Two-column (tree 40vw left, detail 60vw right). Responsive: stack on mobile. Follows exact pattern from `/evaluacion/9x9/jerarquia/+page.svelte`.

**Menu update** (`menuConfig.ts`): Add entry after "Evaluaciones RH":
```ts
{ label: 'Jerarquía', href: '/rh/jerarquia', icon: 'Network', profiles: ['rh'] }
```

## Backend Endpoint Design (Future — Documented for API Migration)

### `GET /org-nodes/{nodeId}/metrics`

| Aspect | Decision |
|--------|----------|
| **Handler** | `api/internal/handler/org/metrics_handler.go` — thin, delegates to service |
| **Service** | `api/internal/service/org/metrics_service.go` — new `MetricsService` interface |
| **SQL approach** | Single query: `SELECT` with `ltree` path filter on `org_nodes`, JOIN `employees` on `org_node_id`, JOIN `goal_assignments` + `goals` for progress, JOIN `rh_evaluations` for ratings. Use `AVG()`, `COUNT()` with `GROUP BY`. No N+1. |
| **RBAC** | Middleware: `RequireAuth` + `RequirePermission(PermOrgRead)`. RH role already has `PermOrgRead`. |
| **Query param** | `cyclePhase` — determines which metrics to compute (skip unnecessary JOINs) |
| **Response** | Matches `AreaMetrics` + `employees[]` shape above |

**SQL sketch (Ent/pgx):**
```sql
SELECT
  n.id AS node_id,
  COUNT(DISTINCT e.id) AS employee_count,
  AVG(CASE WHEN g.progress IS NOT NULL THEN g.progress::float / g.target_value * 100 END) AS avg_progress,
  COUNT(CASE WHEN g.progress IS NOT NULL AND g.progress >= g.target_value THEN 1 END) AS completed,
  COUNT(CASE WHEN g.id IS NOT NULL AND (g.progress IS NULL OR g.progress < g.target_value) THEN 1 END) AS pending
FROM org_nodes n
JOIN employees e ON e.org_node_id = n.id OR e.org_node_id IN (
  SELECT id FROM org_nodes WHERE path <@ n.path
)
LEFT JOIN goal_assignments ga ON ga.employee_id = e.id
LEFT JOIN goals g ON g.id = ga.goal_id
WHERE n.id = $1
GROUP BY n.id;
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/routes/rh/jerarquia/+page.svelte` | Create | RRHH Jerarquía page — tree + detail panel |
| `web/src/lib/stores/rhHierarchyStore.svelte.ts` | Create | Aggregation store with pure functions |
| `web/src/lib/nav/menuConfig.ts` | Modify | Add "Jerarquía" entry for `rh` profile |
| `web/src/lib/stores/rhHierarchyStore.test.ts` | Create | Unit tests for pure aggregation functions |
| `web/src/routes/rh/jerarquia/+page.test.ts` | Create | Component smoke tests |

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| **Unit** | `computeAreaProgress` — correct avg, completed/pending counts, excludes employees with 0 goals | Vitest, known fixture subsets |
| **Unit** | `computeAreaRating` — correct avg rating, null when no ratings, excludes employees with 0 ratings | Vitest |
| **Unit** | `buildEmployeeList` — includes area manager, sorted A-Z, correct profile labels | Vitest |
| **Smoke** | Page renders tree when `profile === 'rh'` | Vitest + Testing Library |
| **Smoke** | Page shows `EmptyState` "Sin acceso" when `profile !== 'rh'` | Vitest + Testing Library |
| **Smoke** | Phase guard: `inicio-anio` shows "Métricas no disponibles" | Vitest |

## Migration / Rollout

No migration required. Fixture-driven; no backend changes in this phase. Future API migration: replace store's fixture imports with `client.GET('/org-nodes/{nodeId}/metrics')` — pure functions become server-side.

## Open Questions

- [ ] **Retail tree**: Spec says "corporate tree only in this change". Confirm retail toggle is explicitly deferred.
- [ ] **Backend RBAC gap**: Current handlers use `RequireAuth` only (no role check). Future endpoint needs `RequirePermission(PermOrgRead)` middleware — confirm this middleware exists or needs creation.
- [ ] **Goal progress %**: `progress / targetValue * 100` assumes `targetValue > 0`. Need to handle `targetValue === 0` edge case (skip or treat as 100%).
