## MODIFIED Requirements

### Requirement: Per-node aggregated metrics panel

Upon selecting any area/subarea node, the system SHALL compute and display per-phase metrics from the backend API. **Metrics are fetched from `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}`**. The frontend SHALL NOT compute metrics locally from fixtures.

#### Scenario: Mid-year metrics from API

- GIVEN `cyclePhase = 'medio-anio'`, RRHH selects node with 4 direct employees
- WHEN detail panel renders
- THEN frontend calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}`
- AND response includes:
  ```json
  {
    "nodeId": "uuid",
    "employeeCount": 4,
    "employeesWithGoals": 3,
    "avgProgress": 67.5,
    "completedGoals": 8,
    "pendingGoals": 12,
    "avgRating": null,
    "ratingsCount": 0,
    "employees": [...]
  }
  ```
- AND `avgProgress` renders as percentage (67.5%)
- AND `completedGoals` and `pendingGoals` render as counts
- AND metrics section shows goal progress, not rating

#### Scenario: End-of-year metrics from API

- GIVEN `cyclePhase = 'fin-anio'`, RRHH selects a subarea with 5 direct employees
- WHEN detail panel renders
- THEN frontend calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}`
- AND response includes:
  ```json
  {
    "nodeId": "uuid",
    "employeeCount": 5,
    "employeesWithGoals": 0,
    "avgProgress": null,
    "completedGoals": 0,
    "pendingGoals": 0,
    "avgRating": 4.2,
    "ratingsCount": 5,
    "employees": [...]
  }
  ```
- AND `avgRating` renders as decimal (4.2)
- AND `ratingsCount` shows how many employees have ratings
- AND metrics section shows rating, not goal progress

#### Scenario: Mixed metrics

- WHEN `cyclePhase` is not specified or the cycle includes both goals and ratings
- THEN response may include both `avgProgress` and `avgRating` non-null
- AND both metrics sections render

#### Scenario: No employees in node

- WHEN node has zero direct employees
- THEN response includes:
  ```json
  {
    "nodeId": "uuid",
    "employeeCount": 0,
    "employeesWithGoals": 0,
    "avgProgress": null,
    "completedGoals": 0,
    "pendingGoals": 0,
    "avgRating": null,
    "ratingsCount": 0,
    "employees": []
  }
  ```
- AND metrics panel shows "Sin empleados en esta área"
- AND employee list section is empty

#### Scenario: API failure fallback

- GIVEN area-metrics endpoint returns error
- WHEN detail panel renders
- THEN metrics section shows EmptyState "Error al cargar métricas"
- AND employee list section still renders (from store data)
- AND retry button is available

### Requirement: rhHierarchyStore consumes API

El `rhHierarchyStore` SHALL consumir `area-metrics` del backend en vez de calcular localmente desde fixtures. Las funciones de cómputo local (`computeAreaProgress`, `computeAreaRating`, `buildEmployeeList`) SE ELIMINARÁN.

#### Scenario: Store fetches metrics from API

- GIVEN RRHH selects a node in the hierarchy
- WHEN `rhHierarchyStore.selectNode(nodeId)` is called
- THEN store calls `GET /org-nodes/{nodeId}/area-metrics?cycleId={currentCycleId}`
- AND stores the full response including `employees` array
- AND `getMetrics()` returns the API response directly
- AND `getEmployeeList()` returns employees from API response (mapped to `EmployeeRow[]`)
- AND NO local computation is performed

#### Scenario: Store fetches metrics without cycleId

- WHEN `selectNode(nodeId)` is called and no cycle is selected
- THEN store calls `GET /org-nodes/{nodeId}/area-metrics` (without cycleId)
- AND response includes metrics across all available data

#### Scenario: Loading state during fetch

- WHEN metrics fetch is in progress
- THEN `isLoadingMetrics()` returns true
- AND UI shows skeleton in metrics section

#### Scenario: Error state on fetch failure

- WHEN metrics fetch fails
- THEN `getMetricsError()` returns error message
- AND metrics section shows EmptyState
- AND employee list section shows last successful data (or empty)

### Requirement: Employee list from API

The employee list SHALL come from the `employees` array in the `area-metrics` API response, not from local fixtures.

#### Scenario: Employee list from API

- WHEN `rhHierarchyStore.getEmployeeList()` is called after successful fetch
- THEN returns `EmployeeRow[]` mapped from API response:
  ```typescript
  // API response employees[] → EmployeeRow[]
  {
    id: employee.id,
    name: `${employee.firstName} ${employee.lastName}`,
    position: profileLabel[employee.profileId],
    profile: employee.profileId
  }
  ```
- AND employees are sorted A-Z by name
- AND profile IDs are mapped to Spanish labels via `PROFILE_LABELS`

#### Scenario: Empty employee list

- WHEN API returns `employees: []`
- THEN `getEmployeeList()` returns `[]`
- AND table renders "Sin empleados en esta área"

### Requirement: Removal of local computation functions

The following functions SHALL be removed from `rhHierarchyStore`:
- `computeAreaProgress()` — replaced by backend API
- `computeAreaRating()` — replaced by backend API
- `buildEmployeeList()` — replaced by API employees array
- `collectNodesByIds()` — no longer needed

Imports of fixture data SHALL be removed:
- `goalsData` from `$lib/fixtures/goals/goals.json`
- `assignmentsData` from `$lib/fixtures/goals/assignments.json`
- `rhEvaluationsData` from `$lib/fixtures/evaluations/rh-evaluations.json`

#### Scenario: No fixture imports in store

- WHEN `rhHierarchyStore.svelte.ts` is compiled
- THEN no imports from `$lib/fixtures/` exist
- AND no local aggregation functions exist
- AND store only depends on `orgHierarchyStore` for scope data and `client` for API calls

### Requirement: Metrics panel phase detection

The metrics panel SHALL detect the current cycle phase and display appropriate metrics:

#### Scenario: Mid-year phase

- WHEN `cycleStore.getCurrentPhase()` returns `'medio-anio'`
- THEN metrics panel shows goal progress section
- AND hides rating section
- AND shows: promedio avance, metas completadas, metas pendientes, empleados con metas

#### Scenario: End-of-year phase

- WHEN `cycleStore.getCurrentPhase()` returns `'fin-anio'`
- THEN metrics panel shows rating section
- AND hides goal progress section
- AND Shows: promedio calificación, empleados calificados

#### Scenario: Phase detection from API response

- WHEN API response has `avgProgress !== null` and `avgRating === null`
- THEN panel assumes mid-year phase
- AND shows goal metrics

- WHEN API response has `avgRating !== null` and `avgProgress === null`
- THEN panel assumes end-of-year phase
- AND shows rating metrics

- WHEN both are non-null
- THEN panel shows both sections

## Acceptance Criteria

1. `rhHierarchyStore` calls `GET /org-nodes/{nodeId}/area-metrics` on node selection
2. Local computation functions (`computeAreaProgress`, `computeAreaRating`, `buildEmployeeList`) are removed
3. Fixture imports are removed from `rhHierarchyStore`
4. Employee list comes from API response `employees` array
5. Metrics panel shows correct metrics based on phase (mid-year vs end-of-year)
6. API failure shows EmptyState with retry option
7. Loading state shows skeleton during fetch
8. `pnpm run check` passes
9. No fixture data in production bundle for rh-hierarchy route
