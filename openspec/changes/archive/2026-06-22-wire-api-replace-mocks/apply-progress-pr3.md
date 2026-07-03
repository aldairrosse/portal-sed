# Apply Progress: wire-api-replace-mocks — PR3 (Component loading/error states)

## Status: In Progress

## Completed Tasks

### T3.4 ✅ — Nine-box components loading/error states

- Added `isLoading()` and `getError()` accessors to `nineBoxStore.svelte.ts`
- `NineBoxMatrix.svelte`: wrapped grid with `PageSkeleton variant="card"` (loading) and `ErrorState` (error with `reload` callback)
- `NineBoxSliders.svelte`: auto-disables sliders when `isLoading()` is true via `isDisabled` derived
- `9x9/+page.svelte`: added `loading`/`error` gates before existing `EmptyState` conditions

## Files Changed (T3.4)

| File | Action |
|------|--------|
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Modified — added `isLoading()`, `getError()` |
| `web/src/lib/components/nine-box/NineBoxMatrix.svelte` | Modified — added loading/error wrappers |
| `web/src/lib/components/nine-box/NineBoxSliders.svelte` | Modified — auto-disabled on loading |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Modified — loading/error gates |
| `openspec/changes/wire-api-replace-mocks/tasks.md` | Modified — marked T3.4 complete |

## Deviations from Design

- Added loading/error at both component level (NineBoxMatrix) and page level (9x9/+page) for layered defense: component handles its own store state, page prevents rendering before data is available

### T3.6 ✅ — Replace getPhase() from devContext with getActivePhase() from cycle store

- Replaced `getPhase()` from `$lib/stores/devContext.svelte` with `getActivePhase()` from `$lib/api/cycle.svelte` in all 6 component/route files
- Used `?? 'inicio-anio'` fallback to handle null return from `getActivePhase()`
- DevToolbar kept its `setProfile`/`setPhase`/`getProfile` devContext imports (dev tool requirements)
- `pnpm check` passes with no new type errors

## Files Changed (T3.6)

| File | Action |
|------|--------|
| `web/src/lib/components/DevToolbar.svelte` | Modified — replaced `getPhase` import, removed from devContext import line |
| `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte` | Modified — replaced `getPhase` with `getActivePhase` |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modified — replaced `getPhase` with `getActivePhase` |
| `web/src/routes/+page.svelte` | Modified — replaced `getPhase` with `getActivePhase` |
| `web/src/routes/mis-evaluados/+page.svelte` | Modified — replaced `getPhase` with `getActivePhase` |
| `web/src/routes/rh/evaluaciones/+page.svelte` | Modified — replaced `getPhase` with `getActivePhase` |
| `openspec/changes/wire-api-replace-mocks/tasks.md` | Modified — marked T3.6 complete |

## Remaining in PR3

- T3.1: evaluation components loading/error states
- T3.2: competency components loading/error states
- T3.3: goals components loading/error states
- T3.5: org-hierarchy components loading/error states
