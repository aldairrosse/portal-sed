# Archive Report: A6 — Manager 9×9 UI (Enhanced)

**Change**: manager-9x9-ui
**Archived**: 2026-06-05
**Previous location**: `openspec/changes/manager-9x9-ui/`
**Archive location**: `openspec/changes/archive/2026-06-05-manager-9x9-ui/`

## Change Summary

Enhanced manager 9×9 UI to support the `director-general` profile with multi-level hierarchy scope, visual 9×9 matrix grid with quadrant coloring, competency network (self vs RH) for single employees, hierarchy tree drill-down (DG → Director → Jefe → Colaborador), and profile-based views that adapt to the user's position in the org hierarchy.

## Deliverables

### PR 1 — Foundation (~350 lines)
Types (`nine-box.ts`, `org-hierarchy.ts`), stores (`nineBoxStore.svelte.ts`, `orgHierarchyStore.svelte.ts`), fixtures (matrix-entries, quadrant-definitions, scale-definitions, org-tree), and evaluation/goal type extensions for `director-general`.

### PR 2 — 9×9 Matrix Core (~300 lines)
`NineBoxMatrix.svelte`, `NineBoxSliders.svelte`, `NineBoxEntryCard.svelte` components replacing `/evaluacion/9x9/+page.svelte` placeholder.

### PR 3 — Hierarchy + Competency Detail (~280 lines)
`OrgHierarchyTree.svelte`, `CompetencyNetworkView.svelte` with new routes `/evaluacion/9x9/competencias/[employeeId]` and `/evaluacion/9x9/jerarquia`.

### PR 4 — Profile Views + Polish (~170 lines)
Profile-specific matrix scopes (jefe → direct reports, director → all descendants, director-general → full org), WCAG 2.1 AA, empty states, visual polish.

**Total estimated changed lines**: ~1100

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| `manager-9x9` | Updated | 1 modified requirement, 5 added requirements (0 removed) |
| `org-hierarchy` | Updated | 2 modified requirements, 3 added requirements (0 removed); non-goals updated |

## Key Decisions Made

| Decision | Choice | Rationale |
|----------|--------|-----------|
| 9×9 scale | Independent 1-9 (separate from competency 1-5) | Finer granularity for performance/potential |
| Hierarchy scope | Recursive OrgNode tree with BFS/DFS traversal | Supports arbitrary depth and drill-down UX |
| Competency detail | Table first (ComparisonTable pattern), graph deferred | Simpler, accessible, reusable |
| DG drill-down | One director at a time | Reduces visual clutter for large orgs |
| Store pattern | Two split stores (nineBoxStore + orgHierarchyStore) | Follows existing pattern (evaluationStore, competencyStore) |

## Files Created/Modified

### New Files
- `web/src/lib/types/nine-box.ts` — NineBoxMatrix, NineBoxEntry, NineBoxQuadrant, NineBoxScale types
- `web/src/lib/types/org-hierarchy.ts` — OrgNode tree structure
- `web/src/lib/stores/nineBoxStore.svelte.ts` — Matrix state, quadrant calculation, CRUD
- `web/src/lib/stores/orgHierarchyStore.svelte.ts` — Tree traversal (getChildren, getDescendants, getSubtree)
- `web/src/lib/components/nine-box/NineBoxMatrix.svelte` — 9×9 grid with quadrant colors and employee dots
- `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` — Detail popup for employee entry
- `web/src/lib/components/nine-box/NineBoxSliders.svelte` — Dual slider for performance/potential
- `web/src/lib/components/org-hierarchy/OrgHierarchyTree.svelte` — Recursive drill-down component
- `web/src/lib/components/evaluation/CompetencyNetworkView.svelte` — Competency table for single employee
- `web/src/lib/fixtures/nine-box/matrix-entries.json` — ~12 entries covering all profiles
- `web/src/lib/fixtures/nine-box/quadrant-definitions.json` — 7 quadrant definitions
- `web/src/lib/fixtures/nine-box/scale-definitions.json` — Anchor labels for 1–9 scale
- `web/src/lib/fixtures/org-hierarchy/org-tree.json` — 4-level hierarchy (DG → Director → Jefe → Colaborador)
- `web/src/routes/evaluacion/9x9/competencias/[employeeId]/+page.svelte` — Competency network view
- `web/src/routes/evaluacion/9x9/jerarquia/+page.svelte` — Hierarchy drill-down view

### Modified Files
- `web/src/lib/types/evaluation.ts` — Added `'director-general'` to EvaluationProfile
- `web/src/lib/types/goal.ts` — Added `director: 'director-general'` to MANAGER_MAP
- `web/src/routes/evaluacion/9x9/+page.svelte` — Replaced placeholder with full matrix
- `web/src/lib/config/menuConfig.ts` — Added director-general profiles and hierarchy/competency menu items
- `web/src/lib/fixtures/assignments.json` — Added DG employee entry (emp-dg-01)

## Lessons Learned

1. **Scope management**: The original scope was large but the 4-PR chain approach kept each slice reviewable (<400 lines each)
2. **Stub-first vs full implementation**: Fixtures and types as foundation allowed parallel work on PR 2 (matrix) and PR 3 (hierarchy/competency)
3. **Profile extensions**: Adding `director-general` required updates across multiple files (types, stores, menu config, pages) — thorough grep for exhaustive switches was essential
4. **WCAG 2.1 AA**: Keyboard navigation for the 9×9 grid (arrow key cell navigation) and tree (up/down/left/right/enter) required careful `role` and `aria-*` attribute management

## Verification

All 24 tasks verified complete across 4 PRs:
- PR 1: 11/11 tasks ✅
- PR 2: 4/4 tasks ✅
- PR 3: 5/5 tasks ✅
- PR 4: 5/5 tasks ✅
