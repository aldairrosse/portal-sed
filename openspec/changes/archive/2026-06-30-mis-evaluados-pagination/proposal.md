# Proposal: Mis Evaluados — Server-Side Pagination

## Intent

Replicate the OFFSET/LIMIT server-side pagination pattern from "Evaluaciones RH" (`rhEvaluadosStore` + `rh/evaluaciones`) to the "Mis Evaluados" section. Currently, mis evaluados loads ALL assignments client-side and filters via `orgHierarchyStore.getChildren()`. This breaks at scale and provides no search capability.

Additionally, scope the view to **only the manager's direct team** (excluding self), matching the team selector pattern from "Asignación de metas".

## What Changes

### Backend

1. **New endpoint** `GET /employees/{empId}/evaluatees?offset=&limit=&q=` — paginated version of the existing `GetMyEvaluatees`
2. **Repository**: `ListByManagerPaginated(managerID, query, offset, limit)` + `CountByManager(managerID, query)` — mirrors `ListWithProfiles`/`CountWithProfiles` pattern
3. **Service**: `GetMyEvaluateesPaginated(evaluatorID, query, offset, limit)` — computes `hasMore = offset + len(rows) < total`
4. **Handler**: Parse `offset`, `limit`, `q` query params, clamp limit 1-200, default 50

### Frontend Store

New `misEvaluadosStore.svelte.ts` — same structure as `rhEvaluadosStore`:
- Module-level `$state`: `items`, `loading`, `error`, `currentPage`, `apiTotal`, `hasMore`, `hasPrev`, `currentQ`
- `load()` reads `currentPage` and `currentQ` from state, calls `GET /employees/{empId}/evaluatees?offset=&limit=&q=`
- `next()`/`prev()` increment/decrement `currentPage`, call `load()`
- `search(query)` debounced 300ms, resets `currentPage = 0`
- Current user's employee ID passed at init (from `devContext`)

### Frontend UI

Rewrite `mis-evaluados/+page.svelte` following `rh/evaluaciones/+page.svelte` pattern:
- Header: "Mis Evaluados" + phase description
- Bar: search input (no `disabled={loading}`) + "Viendo X de Y empleados" / "Viendo X resultado(s)" + Previous/Next buttons
- Skeleton ONLY on initial load: `{#if loading && items.length === 0}`
- Controls always visible outside loading conditional
- `titleCase()` for profile labels
- Employee table in manager mode

### Scope Filtering (Team, No Self)

- The new endpoint filters by `manager_id = current_user` (already done by `ListByManager`)
- **Exclude self**: the current user's own employee ID is excluded from results
- Reference: "Asignación de metas" uses `GET /employees/{empId}/team` — we follow the same team concept but via the evaluatee endpoint (direct reports only, which is the manager's team)
- No org tree traversal needed server-side — `ListByManager` already returns direct reports

## Scope Boundaries

- **In scope**: Backend paginated endpoint, frontend store, frontend UI rewrite, team scoping (no self)
- **Out of scope**: Changes to RH evaluados, changes to assignment page, org hierarchy changes, new table columns, role-based visibility changes

## Chained PR Plan

| PR | Layer | Content |
|----|-------|---------|
| 1 | Backend | Paginated `ListByManagerPaginated` + `CountByManager` in repo, service, handler |
| 2 | Frontend Store | `misEvaluadosStore.svelte.ts` with offset/limit pattern |
| 3 | Frontend UI | Rewrite `mis-evaluados/+page.svelte` with search, pagination, team scoping |

## Open Questions

None — the pattern is well-established by the RH reference implementation.
