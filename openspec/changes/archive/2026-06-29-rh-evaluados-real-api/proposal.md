# Proposal: Connect Evaluados RH Table to Real Employee API

## Intent

The RH evaluaciones table (`/rh/evaluaciones`) currently loads data from `goalsStore.getAssignments()`, which returns only the logged-in user's assignments — not all employees. RH needs to see, search, and paginate the full employee catalog through the existing `GET /api/v1/employees` backend endpoint. This change replaces fixture-scoped data with real API calls, adds server-side search, and cursor-based pagination — without altering the existing table columns.

## Scope

### In Scope

- New store (`rhEvaluadosStore`) consuming `GET /api/v1/employees` with cursor pagination
- Server-side search wired to the existing `q` query parameter
- Replace `EmployeeAssignment[]` with `EmployeeListItem[]` in the RH table
- Cursor-based pagination UI (prev/next) in the table view
- Profile label resolved from API `profileName` field (no local lookup)
- Loading, error, and empty states

### Out of Scope

- Backend changes — `GET /api/v1/employees` already exists
- New table columns — existing 5 columns are fixed (Empleado, Perfil, Progreso global, Estado, Acción)
- Filters by tree/node/profile — deferred to a future change
- Employee detail navigation from this table

## Non-goals

- Modifying `goalsStore` or `EmployeeEvaluationTable` component contract
- Adding org-tree filter sidebar (exists in `/rh/jerarquia`)
- Bulk actions or export

## Capabilities

### New Capabilities

- `rh-evaluados-employee-list`: Store, pagination, and server-side search for the RH evaluaciones table consuming `GET /api/v1/employees`

### Modified Capabilities

- `evaluados-progress-column`: Table component must accept `EmployeeListItem[]` in addition to (or instead of) `EmployeeAssignment[]` for the RH view

## Approach

1. Create `rhEvaluadosStore.svelte.ts` wrapping `GET /api/v1/employees` with reactive `$state` for `cursor`, `limit`, `q`, `items`, `hasMore`, `loading`, `error`
2. Add search input with debounce (300ms) triggering store reload with `q` param
3. Add pagination controls (prev/next) using cursor tokens from API response
4. Adapt `EmployeeEvaluationTable` to accept a generic row type or create a thin RH-specific wrapper that maps `EmployeeListItem` → table props
5. Replace `goalsStore.getAssignments()` call in `+page.svelte` with the new store

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/routes/rh/evaluaciones/+page.svelte` | Modified | Replace goalsStore with rhEvaluadosStore |
| `web/src/lib/stores/rhEvaluadosStore.svelte.ts` | New | API consumer with pagination + search |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modified | Accept `EmployeeListItem[]` or add RH variant |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Table component contract breaks for manager view | Low | Use union type or optional adapter; keep existing `EmployeeAssignment` path intact |
| Search debounce too slow for large datasets | Low | 300ms debounce + cursor reset on new query |

## Rollback Plan

Revert the `+page.svelte` import back to `goalsStore.getAssignments()`. The new store is additive — removing it restores prior behavior with zero data loss.

## Dependencies

- `GET /api/v1/employees` endpoint (already deployed)
- `EmployeeListItem` type from OpenAPI-generated schemas

## Success Criteria

- [ ] RH table loads all employees from `GET /api/v1/employees`, not just current user
- [ ] Search input queries server-side `q` parameter (not client-side filter)
- [ ] Pagination controls navigate through employee pages via cursor
- [ ] Profile column shows `profileName` from API response
- [ ] Loading skeleton shown during fetch; error state on failure
- [ ] `pnpm run check` passes
- [ ] Manager "Mis evaluados" view unaffected
