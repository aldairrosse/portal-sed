# mis-evaluados-ui Specification

## Purpose

Rewrite `mis-evaluados/+page.svelte` following the `rh/evaluaciones/+page.svelte` pattern: server-side pagination, search, skeleton on initial load only, manager-mode evaluation table.

**File**: `web/src/routes/mis-evaluados/+page.svelte`

## Requirements

### Requirement: Page header with phase description

Page SHALL display "Mis Evaluados" title with `Users` icon (from lucide-svelte) and phase-dependent subtitle matching the existing descriptions ("Evaluación formal de competencias y cierre de metas de tu equipo" for fin-anio, etc.).

### Requirement: Search bar with result count

Search input (`input-bordered input-sm`, max-w-sm) SHALL be always enabled (no `disabled={loading}`). Right of the search bar, SHALL display:

| Condition | Text |
|-----------|------|
| No search, has results | "Viendo {items.length} de {totalCount} empleado(s)" |
| Active search, has results | "Viendo {items.length} resultado(s)" |
| No results | Hidden |

#### Scenario: Search updates counter

- GIVEN 120 total evaluatees, 50 per page, no search
- WHEN page loads
- THEN bar shows "Viendo 50 de 120 empleados"

### Requirement: Previous / Next pagination

Buttons SHALL use `btn-outline btn-xs` with `ChevronLeft`/`ChevronRight` icons. Disabled states: `!hasPrev || loading` for Previous, `!hasMore || loading` for Next. Page number SHALL display as "Pág. {currentPage + 1}".

### Requirement: Skeleton only on initial load

Skeleton (`PageSkeleton variant="table" rows={5}`) SHALL render ONLY when `loading && items.length === 0`. During subsequent page loads, SHALL keep existing rows visible and allow pagination controls to show disabled state.

### Requirement: Error state with retry

`ErrorState` component SHALL display `storeError` message with a retry callback calling `load()`. SHALL show when `storeError` is truthy.

### Requirement: Empty state

When `items.length === 0 && !loading`, SHALL display centered italic message in Spanish: "Sin evaluados para mostrar".

### Requirement: Manager-mode evaluation table

`EmployeeEvaluationTable` SHALL render with `mode="manager"`, receiving `rows={items}`. Table SHALL pass `selectedEmployeeId` and `onSelect` for drill-down into `EmployeeEvaluationDetail` with `viewerMode="manager"` and `showBreadcrumb={true}`. Row selection SHALL be disabled except during `fin-anio` phase.

### Requirement: titleCase profile labels

Profile names from `items[].profileName` (e.g., "gerente de tienda") SHALL be rendered via `titleCase()` from `$lib/utils/text` in the table's profile column.

### Requirement: Store integration

Page SHALL import from `misEvaluadosStore.svelte.ts` using the same getter pattern as `rh/evaluaciones` (`getItems`, `isLoading`, `getError`, `hasMoreItems`, `hasPrevItems`, `getCurrentPage`, `getTotalCount`, `load`, `search`, `next`, `prev`). `onMount` SHALL call `init(employeeId)` then `load()`. Employee ID SHALL come from `devContext.getProfile()`.
