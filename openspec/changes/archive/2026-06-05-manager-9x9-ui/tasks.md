# Tasks: A6 — Manager 9×9 UI (Enhanced)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1100 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 → PR 4 |
| Delivery strategy | ask-on-risk |
| Chain strategy | feature-branch-chain |
| Decision needed before apply | No |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: feature-branch-chain
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation (types, stores, fixtures) | PR 1 | Base for all; targets `feature/a6-manager-9x9` |
| 2 | 9×9 Matrix Core (grid + sliders) | PR 2 | Depends on PR 1; target = PR 1 branch |
| 3 | Hierarchy + Competency (tree + table) | PR 3 | Depends on PR 1; parallel with PR 2; target = PR 1 branch |
| 4 | Profile Views + Integration | PR 4 | Depends on PR 2 + PR 3; target = PR 3 branch |

## PR 1: Foundation

> Types, stores, and fixture scaffolding that PR 2–3 depend on

- [x] 1.1 Modify `evaluation.ts` — add `'director-general'` to `EvaluationProfile`, `EVALUATION_PROFILES`, `PROFILE_LABELS`
- [x] 1.2 Modify `goal.ts` — add `director: 'director-general'` to `MANAGER_MAP`
- [x] 1.3 Create `types/nine-box.ts` — `NineBoxScale`, `NineBoxQuadrant`, `NineBoxEntry`, `NineBoxQuadrantDef`, `NineBoxMatrix`
- [x] 1.4 Create `types/org-hierarchy.ts` — `OrgNode` type
- [x] 1.5 Create `stores/nineBoxStore.svelte.ts` — state, getters (`getMatrixEntries`, `getQuadrantForScores`, `getQuadrantStats`), mutations (`setEntryScores`, `bulkSetEntries`), quadrant calc
- [x] 1.6 Create `stores/orgHierarchyStore.svelte.ts` — tree traversal: `getChildren`, `getDescendants`, `getSubtree`, `getNodeById`, `getScopeIds`, `getDepth`, `getAllLeafIds`
- [x] 1.7 Create `fixtures/nine-box/matrix-entries.json` — ~12 entries covering all profiles with mixed quadrants
- [x] 1.8 Create `fixtures/nine-box/quadrant-definitions.json` — 7 quadrant defs with DaisyUI color classes
- [x] 1.9 Create `fixtures/nine-box/scale-definitions.json` — anchor labels for 1–9 scale
- [x] 1.10 Create `fixtures/org-hierarchy/org-tree.json` — 4-level tree: DG → Director → Jefe → Colaborador
- [x] 1.11 Modify `assignments.json` — add DG employee entry (`emp-dg-01`)

**Acceptance:** `pnpm run check` passes; `orgHierarchyStore.getDescendants('emp-dg-01')` returns 7 nodes; `getQuadrantForScores(8, 7)` returns `'star'`

## PR 2: 9×9 Matrix Core

> Grid component, entry detail, slider controls, and main 9×9 page

- [x] 2.1 Create `nine-box/NineBoxMatrix.svelte` — 9×9 CSS grid, quadrant colors, employee dots, count badges, WCAG grid roles, keyboard nav
- [x] 2.2 Create `nine-box/NineBoxEntryCard.svelte` — popup detail with name, scores, quadrant, links to competencias & jerarquía
- [x] 2.3 Create `nine-box/NineBoxSliders.svelte` — dual range inputs (perf + pot), real-time quadrant label, DaisyUI `range` style
- [x] 2.4 Replace `routes/evaluacion/9x9/+page.svelte` — profile guard, scope from `orgHierarchyStore`, render `NineBoxMatrix` with scoped entries, `NineBoxEntryCard` on click

**Acceptance:** Matrix renders with 7 colored quadrants; clicking employee dot shows card; `pnpm run check` passes; keyboard navigates grid cells

## PR 3: Hierarchy + Competency Detail

> Tree drill-down, competency table, new routes, and menu items

- [x] 3.1 Create `org-hierarchy/OrgHierarchyTree.svelte` — recursive component (OrgHierarchyTree.svelte + TreeNode.svelte), expand/collapse, DaisyUI tree roles, WCAG keyboard nav (ArrowRight/Left/Up/Down/Enter)
- [x] 3.2 Create `evaluation/CompetencyNetworkView.svelte` — table: self vs RH ratings per competency grouped by pillar, gap highlight (>1 diff), reuses `ComparisonTable` pattern
- [x] 3.3 Create `routes/evaluacion/9x9/competencias/[employeeId]/+page.svelte` — breadcrumb, employee header (avatar, profile), `CompetencyNetworkView`
- [x] 3.4 Create `routes/evaluacion/9x9/jerarquia/+page.svelte` — profile guard (director/director-general), header, `OrgHierarchyTree`, node summary panel (profile, direct reports, team total, link to competencias)
- [x] 3.5 Modify `menuConfig.ts` — add `'director-general'` to Matriz 9×9 profiles; add "Jerarquía" `/evaluacion/9x9/jerarquia` (director/director-general) and "Competencias" `/evaluacion/9x9` (jefe/director/director-general) menu items

**Acceptance:** Tree renders 4 levels collapsed; expand loads children; competency table shows self vs RH with gap highlighting; menu shows 9×9 sub-items for jefe/director/director-general; `pnpm run check` passes

## PR 4: Profile Views + Integration

> Profile scope wiring, polish, and verification

- [x] 4.1 Wire profile-based scope on 9×9 page — `jefe` → `getChildren` (direct reports), `director` → `getDescendants`, `director-general` → full tree, `rh` → all employees with edit (already implemented in PR 2)
- [x] 4.2 Wire profile guard on jerarquía page — `director`/`director-general` only (others see `EmptyState`) (already implemented in PR 3)
- [x] 4.3 Verify EntryCard links — "Ver competencias" navigates to `/[employeeId]`, "Ver jerarquía" navigates to `/jerarquia` (already implemented in PR 2-3)
- [x] 4.4 WCAG 2.1 AA pass — focus trap + focus return on `NineBoxEntryCard` popup; roving tabindex on `TreeNode` for correct tree keyboard nav; `aria-label` verified on all interactive elements
- [x] 4.5 Visual polish — `readonly={profile !== 'rh'}` on `NineBoxMatrix`; empty states for unauthorized and no-data per profile; consistent heading hierarchy (`h1` → `h2`) across all pages

**Acceptance:** Jefe sees 2 evaluatees, director sees 7, DG sees all 8 (or 9 with DG user); `EmptyState` for unauthorized profiles on jerarquía; `pnpm run check` + `pnpm run lint` pass; Tab/Enter navigation works end-to-end
