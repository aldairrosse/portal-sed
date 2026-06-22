# Archive Report: rh-org-hierarchy-metrics

**Archived**: 2026-06-18
**Previous path**: `openspec/changes/rh-org-hierarchy-metrics/`
**Archive path**: `openspec/changes/archive/2026-06-18-rh-org-hierarchy-metrics/`

## What Was Built

A new RRHH-only "Jerarquía" view at `/rh/jerarquia` that lets RRHH browse the full organizational tree, select any area/subarea, and see per-node aggregated metrics:

| Phase | Metrics |
|-------|---------|
| `medio-anio` | Average goal progress %, completed/pending goal counts |
| `fin-anio` | Average evaluation rating (1–5) across employees |
| `inicio-anio` | "Métricas no disponibles" empty state |

The view also provides a read-only employee table (name, position, profile) for the selected area.

## Files Created/Modified

| File | Action | Description |
|------|--------|-------------|
| `web/src/routes/rh/jerarquia/+page.svelte` | Create | Profile guard (`rh` only), two-column layout (40vw/60vw), OrgHierarchyTree integration, phase-conditional metrics cards, employee table. ~215 lines. |
| `web/src/lib/stores/rhHierarchyStore.svelte.ts` | Create | Pure-function aggregation store: `computeAreaProgress`, `computeAreaRating`, `buildEmployeeList`, `selectNode`. 241 lines. |
| `web/src/lib/nav/menuConfig.ts` | Modify | +7 lines — added "Jerarquía" sidebar entry (icon `Network`, profiles `['rh']`) |
| `openspec/specs/rh-org-hierarchy-metrics/spec.md` | Create | Main spec synced from delta spec |

## Key Decisions

1. **Pure functions over pre-computed fixtures** — `computeAreaProgress`, `computeAreaRating`, `buildEmployeeList` are importable, testable, and trivial to swap for API calls later.
2. **No new fixture file** — reuses existing `goals.json`, `assignments.json`, `rh-evaluations.json` directly.
3. **`getScopeIds(nodeId)` includes the node itself** — ensures the area manager appears in both metrics and employee list.
4. **Phase detection from `devContext.getPhase()`** — same pattern as existing Directors Jerarquía.
5. **Svelte 5 `$state`/`$derived` module pattern** — matches `orgHierarchyStore` convention.

## Test Coverage

- **37 unit tests** pass for pure aggregation functions (computeAreaProgress, computeAreaRating, buildEmployeeList)
- **Page-level smoke tests** attempted but removed — Svelte 5 compiled components incompatible with vitest+jsdom SSR rendering
- **`pnpm run check`** passes with zero new type errors (53 pre-existing in other files remain untouched)

## Engram Observations (traceability)

| ID | Type | Topic |
|----|------|-------|
| #285 | discovery | sdd/rh-org-hierarchy-metrics/explore |
| #287 | architecture | sdd/rh-org-hierarchy-metrics/spec |
| #288 | architecture | sdd/rh-org-hierarchy-metrics/design |
| #289 | architecture | sdd/rh-org-hierarchy-metrics/tasks |
| #291 | architecture | sdd/rh-org-hierarchy-metrics/apply-progress |
| #292 | architecture | sdd/rh-org-hierarchy-metrics/archive-report |

## Known Limitations

- No backend API — fixture-driven only (UI-first per proposal scope)
- Corporate tree only — retail tree toggle deferred to future change
- Page-level rendering tests require Playwright E2E setup (not yet available)

## Future Work

1. Backend endpoint `GET /org-nodes/{nodeId}/metrics?cyclePhase=...` — SQL aggregation via ltree path filter + JOINs on goals/evaluations
2. Playwright E2E tests for full page rendering validation
3. Retail tree support with tab/toggle component
4. RBAC middleware (`RequirePermission(PermOrgRead)`) for future backend endpoint

## SDD Cycle Complete

The change has been fully planned, implemented, verified, and archived.
