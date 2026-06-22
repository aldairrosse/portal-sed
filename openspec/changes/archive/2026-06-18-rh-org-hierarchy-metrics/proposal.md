# Proposal: rh-org-hierarchy-metrics — RRHH Hierarchy Metrics View

## 1. Intent

Add a **"Jerarquía"** section at `/rh/jerarquia` visible exclusively to users with the **RRHH** role. The view lets RRHH browse the full organizational tree (areas and subareas), select any node, and see per-area aggregated metrics:

- **Mid-year (phase `avance`)**: average goal progress of all employees in the selected area (including the area manager).
- **End-of-year (phase `cierre`)**: average evaluation rating across employees in the area.
- **Extra**: completed vs. pending goal counts for the selected area.
- **Employee list**: read-only list of all employees belonging to the selected area/subarea.

The goal is to give RRHH a transversal, data-driven view of organizational performance without needing to drill into each individual's evaluation manually.

## 2. Scope

### In Scope
- New route `/rh/jerarquia` — accessible only to RRHH profile; phase-conditional rendering of metric type.
- Reuse `OrgHierarchyTree` + `TreeNode` components (no new tree implementation).
- Area/subarea selection triggers detail panel with aggregated metrics.
- Metrics computation from goals fixture (mid-year progress avg; end-of-year rating avg) scoped to the selected node's subtree employees.
- Completed vs. pending goal count per area.
- Read-only employee list (name, position) for selected area.
- RBAC guard on the route (RRHH only); redirect/empty state for other profiles.

### Out of Scope
- Backend API or persistence (UI-first; fixtures only).
- Exporting metrics or bulk actions on employees.
- Historical trend charts or period comparisons.
- Editing any employee, goal, or evaluation from this view.
- Modifying the existing Directors Jerarquía view at `/evaluacion/9x9/jerarquia`.

## 3. Approach

### Route & Components
```
web/src/routes/rh/jerarquia/+page.svelte   ← new
web/src/lib/stores/rhHierarchyStore.svelte.ts  ← new (aggregation logic)
web/src/lib/fixtures/rh-hierarchy/metrics.json  ← new (mock area aggregates)
```

### Metric Computation (fixture-driven)
| Cycle Phase | Metric | Source |
|---|---|---|
| `avance` | Avg goal progress % | `goals.json` filtered by employeeIds in subtree |
| `cierre` | Avg evaluation rating (1–5) | `rh-evaluations.json` filtered by employeeIds |
| `avance` | Completed goals count | Goals with `progress >= targetValue` |
| `avance` | Pending goals count | Goals with `progress < targetValue` or null |

### Aggregation Flow
1. User selects an `OrgNode` in the tree → `rhHierarchyStore` receives `nodeId`.
2. Store calls `getDescendants(nodeId)` from `orgHierarchyStore` → collects all employee IDs in subtree.
3. Filters `goals.json` / `rh-evaluations.json` by those employee IDs.
4. Computes averages and counts → exposes as reactive `$state`.

### Reuse Strategy
- `OrgHierarchyTree` and `TreeNode`: used as-is (existing Directors Jerarquía components).
- `orgHierarchyStore`: reused for tree traversal (`getDescendants`, `getNodeById`).
- `devContext.svelte.ts`: reused for `profile` and `cyclePhase` detection.
- `PROFILE_LABELS`, DaisyUI components (`card`, `badge`, `table`): reused.

### Phase-Conditional Rendering
```ts
const metricType = $derived(
  cyclePhase === 'avance'  ? 'progress' :
  cyclePhase === 'cierre' ? 'rating'   : 'unavailable'
);
```

## 4. Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Aggregation logic produces misleading averages (e.g., mixing employees with 0 goals) | Medium | Filter to only employees with at least one goal; show "—" for empty areas |
| Fixture data not representative of real aggregation behavior | Low | Keep aggregation functions pure and documented; easy to swap for real API later |
| Scope creep: RRHH asks for drill-down into individual evaluations from this view | Medium | Explicit out-of-scope; link to existing employee evaluation routes instead |

## 5. Dependencies

| Dependency | Change / Spec | Notes |
|---|---|---|
| `org-hierarchy` spec | `openspec/specs/org-hierarchy/spec.md` | OrgNode type, tree traversal functions |
| `mid-year-progress-ui` spec | `openspec/specs/mid-year-progress-ui/spec.md` | `progress` field on Goal |
| `annual-evaluation-ui` spec | `openspec/specs/annual-evaluation-ui/spec.md` | Evaluation rating data shape |
| `goals-and-weighting` spec | `openspec/specs/goals-and-weighting/spec.md` | Goal data model (weight, unit, targetValue) |
| `OrgHierarchyTree` component | `web/src/lib/components/org-hierarchy/OrgHierarchyTree.svelte` | Reused as-is |
| `orgHierarchyStore` | `web/src/lib/stores/orgHierarchyStore.svelte.ts` | Tree traversal helpers |
| `devContext` store | `web/src/lib/stores/devContext.svelte.ts` | Profile + cycle phase |

## 6. Success Criteria

- [ ] `/rh/jerarquia` renders a full org tree (all corporate + retail nodes) when logged in as RRHH.
- [ ] Selecting any area/subarea shows correct average goal progress (mid-year) or evaluation rating (end-of-year) based on `cyclePhase`.
- [ ] Completed vs. pending goal counts are correct for the selected area.
- [ ] Employee list shows all employees in the selected area/subarea (including the area manager) with name and position.
- [ ] Non-RRHH profiles see `EmptyState` with "Sin acceso" message.
- [ ] Component reuses `OrgHierarchyTree`, `TreeNode`, `orgHierarchyStore` without modifications.
- [ ] All metric computations are pure functions in `rhHierarchyStore`; easy to swap fixture → API later.
- [ ] `pnpm run check` passes with no type errors.
- [ ] Phase guard: when `cyclePhase` is `asignacion`, a descriptive empty state is shown (metrics not yet applicable).

## Non-goals

- No backend API; fixtures only.
- No authentication changes; RRHH access assumed from existing RBAC.
- No modification of the Directors Jerarquía view.
- No export, bulk actions, or drill-down into individual evaluations.
