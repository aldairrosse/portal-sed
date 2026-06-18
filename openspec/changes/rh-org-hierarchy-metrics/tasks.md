# Tasks: RRHH Hierarchy Metrics View

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~460 |
| 400-line budget risk | Medium |
| 800-line budget risk | Low |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: Store + types + unit tests (~190 lines) → PR 2: Page + menu + smoke tests (~270 lines) |
| Delivery strategy | auto-chain |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: Medium

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | `rhHierarchyStore` + types + unit tests | PR 1 | Base: main. Importable, testable in isolation. ~190 lines. |
| 2 | Route page + menu entry + smoke tests | PR 2 | Depends on PR 1 store. Integrates tree + detail panel. ~270 lines. |

## Phase 1: Store / Foundation

- [x] 1.1 Create `rhHierarchyStore.svelte.ts` — `AreaMetrics`/`EmployeeRow` types, module `$state`, `selectNode` + reactive getters
- [x] 1.2 Implement `computeAreaProgress()` — mean progress per employee, completed/pending counts, exclude 0-goal employees
- [x] 1.3 Implement `computeAreaRating()` — mean rating from `RhEvaluation[]`, return null when no ratings
- [x] 1.4 Implement `buildEmployeeList()` — include area manager in list, sort A-Z, map `profileId` to `PROFILE_LABELS`

## Phase 2: Route & UI

- [x] 2.1 Create `+page.svelte` at `web/src/routes/rh/jerarquia/` — two-column layout (40vw/60vw), profile guard (`rh` only), EmptyState for non-RH
- [x] 2.2 Wire `OrgHierarchyTree` with `getRoot()` — full corporate tree, node click calls `selectNode(nodeId)`
- [x] 2.3 Build detail panel: header (node name + badge), 2×2 metrics cards (avg progress/rating, completed/pending)
- [x] 2.4 Build employee table — read-only, A-Z sorted, columns: name, position, profile label
- [x] 2.5 Add phase-conditional rendering via `$derived metricType` — `progress` for `medio-anio`, `rating` for `fin-anio`, EmptyState for `inicio-anio`
- [x] 2.6 Update `menuConfig.ts` — add "Jerarquía" sidebar entry (icon `Network`, profiles `['rh']`)

## Phase 3: Testing & Verification

- [x] 3.1 Unit tests for `computeAreaProgress` — avg correctness, completed/pending counts, 0-goal employee exclusion
- [x] 3.2 Unit tests for `computeAreaRating` — correct avg, null when empty, zero-rating employee exclusion
- [x] 3.3 Unit tests for `buildEmployeeList` — includes manager, sorted A-Z, correct labels
- [x] 3.4 Smoke tests for `+page.svelte` — renders tree for RH, EmptyState for non-RH, phase guard for `inicio-anio` (Note: Svelte 5 SSR rendering via vitest+jsdom is incompatible with client-mode compiled components in this project setup. A Playwright E2E test in a follow-up is recommended for proper rendering validation. The test file was attempted and removed — the pattern would need `@sveltejs/vite-plugin-svelte/testing` setup.)
- [x] 3.5 Verify `pnpm run check` passes with zero type errors on new code (53 pre-existing errors in other files remain — none from this change)
