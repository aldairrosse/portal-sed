# Tasks: Connect Evaluados RH Table to Real API

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 220-280 |
| 400-line budget risk | Low |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: store → PR 2: table + page |
| Delivery strategy | chained |
| Chain strategy | stacked-to-main |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | New `rhEvaluadosStore.svelte.ts` with load/search/pagination | PR 1 | Base: main; self-contained; test debounce + cursor logic |
| 2 | Wire table + page to new store | PR 2 | Base: PR 1 branch; modify `EmployeeEvaluationTable` + `+page.svelte` |

## Phase 1: Store Foundation (PR 1)

- [x] 1.1 Create `web/src/lib/stores/rhEvaluadosStore.svelte.ts` with module-level `$state`: `items`, `loading`, `error`, `hasMore`, `hasPrev`
- [x] 1.2 Implement `load(q?, cursor?)` calling `client.GET("/employees", { params: { query: { q, cursor, limit: 50 } } })` and populating state from `EmployeeListResponse`
- [x] 1.3 Implement `next()` / `prev()` setting cursor from response meta and calling `load()`
- [x] 1.4 Implement `search(query)` with inline 300ms `setTimeout`/`clearTimeout` debounce, resetting cursor on new query

## Phase 2: Table + Page Wiring (PR 2)

- [x] 2.1 Add `mode: 'rh' | 'manager'` prop to `EmployeeEvaluationTable.svelte`; when `'rh'`, accept `rows: EmployeeListItem[]`, render `profileName` directly, show "—" for progress/status columns
- [x] 2.2 Modify `web/src/routes/rh/evaluaciones/+page.svelte`: replace `getAssignments()` with `rhEvaluadosStore` items; add search input wired to `search()`; add prev/next buttons bound to `hasPrev`/`hasMore`; pass `mode="rh"` to table
- [x] 2.3 Add loading skeleton (5 rows), error banner with retry, and empty state "Sin empleados para mostrar" to the RH page

## Phase 3: Verification

- [x] 3.1 Run `pnpm run check` — no type errors from local `EmployeeListItemExtended` or new prop
- [ ] 3.2 Manual smoke: load RH page → employees appear with `profileName` in Perfil column → search filters via API → prev/next paginate
- [ ] 3.3 Manual smoke: manager `/evaluaciones` page unaffected (still uses `EmployeeAssignment[]` path)
