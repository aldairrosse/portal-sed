# Delta Spec: rh-org-hierarchy-metrics — RRHH Hierarchy Metrics View

> Delta for: new domain. No existing `rh-org-hierarchy-metrics` spec to modify.

## ADDED Requirements

### Requirement: RRHH Jerarquía — route and role guard

The system SHALL serve a new route `/rh/jerarquia` accessible exclusively when the active dev profile is `rh`. Non-RRHH profiles SHALL see `EmptyState` with "Sin acceso" and a link back to `/`.

#### Scenario: RRHH profile sees Jerarquía

- GIVEN `profile = 'rh'`
- WHEN RRHH navigates to `/rh/jerarquia`
- THEN the full organizational tree (corporate root) renders
- AND the detail panel prompts "Selecciona un área para ver métricas"

#### Scenario: Non-RRHH profile blocked

- GIVEN `profile !== 'rh'`
- WHEN any user navigates to `/rh/jerarquia`
- THEN `EmptyState` renders with title "Sin acceso" and message "Solo el área de RRHH puede ver las métricas de la organización"

### Requirement: Dual-tree org hierarchy for RRHH

The system SHALL render the complete corporate org tree and SHALL support loading the retail tree via a tab or toggle. RRHH SHALL browse all areas and subareas in both trees. The tree SHALL reuse `OrgHierarchyTree` + `TreeNode` unchanged.

#### Scenario: Corporate tree loads by default

- GIVEN RRHH profile active
- WHEN `/rh/jerarquia` loads
- THEN the corporate tree renders with `OrgHierarchyTree` using `getRoot()` from `orgHierarchyStore`
- AND all nodes are selectable regardless of profileId

#### Scenario: Leaf node is selectable

- GIVEN RRHH expands the tree to a `colaborador` or `vendedor` leaf node
- WHEN RRHH clicks a leaf node
- THEN the detail panel shows metrics and employee list for that node's direct employees (single employee if leaf)
- AND the node is highlighted in the tree

### Requirement: Per-node aggregated metrics panel

Upon selecting any area/subarea node, the system SHALL compute and display per-phase metrics from fixture data. **Metrics are scoped to the selected node's direct employees only** (employees whose `org_node_id` matches the selected node), including the area manager. Sub-nodes and their employees are NOT included in the aggregation.

| Phase (`CyclePhase`) | Metric | Computation | Visual |
|---|---|---|---|
| `medio-anio` | Avg goal progress (%) | `mean(goals.progress)` of direct employees in node (only employees with ≥1 goal) | `ProgressIndicator` + badge |
| `medio-anio` | Completed goals | Count of `goals` where `progress ≥ targetValue` for direct employees | Badge-success with count |
| `medio-anio` | Pending goals | Count of `goals` where `progress < targetValue` or null for direct employees | Badge-warning with count |
| `fin-anio` | Avg evaluation rating (1–5) | `mean(rhRating)` across all `CompetencyRating` for direct employees in node | Numeric badge + star |
| `inicio-anio` | — | Metrics not applicable | EmptyState "Métricas no disponibles en inicio de año" |

Employees with zero goals or zero competency ratings SHALL be excluded from means. Empty nodes SHALL show "—" (not zero).

#### Scenario: Mid-year metrics for an area with goals

- GIVEN `cyclePhase = 'medio-anio'`, RRHH selects node with 4 direct employees (3 with goals)
- WHEN detail panel renders
- THEN avg progress = mean of those 3 employees' goal progress values
- AND completed count = goals meeting target, pending count = goals not meeting target
- AND the 1 employee with no goals is excluded from the mean
- AND employees in sub-nodes are NOT included

#### Scenario: End-of-year metrics for an area

- GIVEN `cyclePhase = 'fin-anio'`, RRHH selects a subarea with 5 direct employees
- WHEN detail panel renders
- THEN avg rating = mean of all `rhRating` values across all competencies for those 5 direct employees
- AND employees with no ratings are excluded; if none have ratings, shows "—"
- AND employees in sub-nodes are NOT included

#### Scenario: Phase renders correct metric type (inicio-anio)

- GIVEN `cyclePhase = 'inicio-anio'`
- WHEN RRHH selects any node
- THEN the metrics section shows EmptyState "Métricas no disponibles en inicio de año"
- AND the employee list section still renders

### Requirement: Employee list per area/subarea

The system SHALL render a read-only table of all employees belonging to the selected node's direct level (employees whose `org_node_id` matches the selected node). Columns: name, position (from `OrgNode`), profile (label in Spanish). The list SHALL be sorted alphabetically by name. Rows SHALL NOT be clickable (no drill-down).

#### Scenario: Employee list for a director area

- GIVEN RRHH selects Director A node
- WHEN the detail panel renders the employee list
- THEN all descendants (jefes, colaboradores, vendedores) appear in a table
- AND the area manager (Director A) is included in the list
- AND the table is sorted by name A-Z

#### Scenario: Employee list for a leaf node

- GIVEN RRHH selects a `colaborador` leaf node
- WHEN the detail panel renders
- THEN the table shows exactly 1 row (the selected employee)

### Requirement: Non-functional — pure aggregation functions

All metric computations SHALL be pure functions in `rhHierarchyStore` accepting fixture data as input and returning computed aggregates. No side effects, no direct DOM access, no mutation of `orgHierarchyStore` data. Rationale: trivial to replace with API call later — swap `fetchMetrics(fixtures, employeeIds)` for `client.GET('/org-nodes/{nodeId}/metrics')`.

#### Scenario: Pure function testability

- GIVEN a known set of employee IDs and mock goals fixture
- WHEN `computeAreaProgress(employeeIds, goalsFixture)` is called
- THEN it returns the correct mean and counts
- AND the function is importable and testable in isolation

---

## OpenAPI Contract (Stub — Future Backend)

The exploration report identified the need for a backend aggregation endpoint. The contract below is a **stub** for this change (fixture-driven; no `api/` code).

### GET /org-nodes/{nodeId}/metrics

**Query params**: `cyclePhase` (`inicio-anio` | `medio-anio` | `fin-anio`)

**Response** `200`:
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

**Notes**: `avgEvaluationRating` is null when `cyclePhase !== 'fin-anio'`. `avgGoalProgress` is null when `cyclePhase !== 'medio-anio'`.

---

## UI Spec

| Aspect | Decision |
|--------|----------|
| **Route** | `/rh/jerarquia` — new `+page.svelte` |
| **Menu entry** | Sidebar: "Jerarquía" under RRHH group when `profile === 'rh'` |
| **Layout** | Two-column: tree (left, ~40vw) + detail panel (right, ~60vw). Responsive: stack on mobile. |
| **Tree** | Reuse `OrgHierarchyTree` + `TreeNode` as-is. Root = `getRoot()` from `orgHierarchyStore`. |
| **Detail panel** | Three sections: header (node name + profile label), metrics cards (2×2 grid), employee table. |
| **Store** | New `web/src/lib/stores/rhHierarchyStore.svelte.ts` — aggregation logic. Reads from `orgHierarchyStore` + fixtures. |
| **Fixtures** | New `web/src/lib/fixtures/rh-hierarchy/metrics.json` — mock aggregated data for known node IDs. |
| **Components reused** | `OrgHierarchyTree`, `TreeNode`, `EmptyState`, `PageSkeleton`, `ProgressIndicator` |
| **Role guard** | `$derived`: `profile === 'rh'` — blocks rendering for non-RRHH. |
| **Phase guard** | `$derived`: checks `cyclePhase` to render `progress` vs `rating` vs `unavailable` metrics. |

---

## Data Model / Query Notes

### Aggregation flow (fixture-driven)

```
User selects OrgNode (nodeId)
      │
      ▼
orgHierarchyStore.getDescendants(nodeId) → employeeIds[]
      │
      ▼
filter goals.json / rh-evaluations.json by employeeIds
      │
      ▼
computeAreaProgress(employeeIds, goals) → { avgProgress, completedCount, pendingCount }
computeAreaRating(employeeIds, ratings) → { avgRating, ratingsCount }
      │
      ▼
Reactively bind to UI via $derived / $state
```

### Metrics JSON fixture shape

```jsonc
{
  "emp-director-01": { "avgProgress": 65, "completedGoals": 7, "pendingGoals": 5, "avgRating": 3.8, "employeeCount": 7 },
  "emp-jefe-01": { "avgProgress": 71, "completedGoals": 3, "pendingGoals": 2, "avgRating": 4.1, "employeeCount": 3 }
}
```

---

## Out of Scope

- Backend API or persistence (fixtures only).
- Exporting metrics (CSV/PDF).
- Historical trend charts or period comparisons.
- Drill-down into individual employee evaluations from this view.
- Editing any employee, goal, or evaluation data.
- Modifying the existing Directors Jerarquía at `/evaluacion/9x9/jerarquia`.
- Retail tree fixture (corporate tree only in this change; retail toggle is a future addition).
- Authentication changes; RBAC is assumed from existing dev persona devContext.

---

## Acceptance Criteria

- [ ] `/rh/jerarquia` renders full corporate tree with `OrgHierarchyTree` when `profile === 'rh'`.
- [ ] Non-RRHH profiles see `EmptyState` "Sin acceso".
- [ ] Selecting any area/subarea shows per-phase aggregated metrics from fixture.
- [ ] Mid-year (`medio-anio`): avg progress, completed count, pending count render correctly.
- [ ] End-of-year (`fin-anio`): avg evaluation rating renders correctly.
- [ ] Start-of-year (`inicio-anio`): "Métricas no disponibles" empty state.
- [ ] Employee table shows all direct employees with name, position, profile.
- [ ] Employee table includes the area manager.
- [ ] `OrgHierarchyTree` and `TreeNode` are reused without modification.
- [ ] Aggregation functions are pure and importable from `rhHierarchyStore`.
- [ ] `pnpm run check` passes with no type errors.
- [ ] Sidebar shows "Jerarquía" entry only for RRHH profile.
