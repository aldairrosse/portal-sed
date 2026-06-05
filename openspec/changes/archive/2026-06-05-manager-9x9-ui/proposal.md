# Proposal: A6 — Manager 9×9 UI (Enhanced)

## Intent

Build the enhanced manager 9×9 UI: a 9×9 performance vs potential matrix with competency networks for single employees, top-down hierarchy drilling (director-general → director → jefe → colaborador), and profile-based views that adapt to the user's position in the org hierarchy.

## Scope

### In Scope
- **9×9 Matrix View**: Director/director-general sees 9×9 grid for all employees under their hierarchy (not just direct reports)
- **Competency Networks**: Table view of competency ratings (self vs RH) for a single employee; graph visualization as future enhancement
- **Hierarchy Drilling**: Recursive drill-down tree (director-general → directors → managers → employees); arbitrary depth support
- **Profile-Based Views**: Jefe sees matrix for direct reports; director sees matrix for all managers in their span; director-general sees entire org

### Out of Scope
- Backend persistence (fixtures only, API stubbed)
- Login/auth (dev persona selector)
- 9×9 matrix editing UI for RH (read-only aggregation view)
- Graph/network visualization (table first; D3/visjs as future)
- Historical comparison between cycles
- Export to PDF/Excel

## Capabilities

### New Capabilities
- `nine-box-matrix`: 9×9 grid with performance (1–9) and potential (1–9) axes; quadrant colors; employee dots positioned by score
- `org-hierarchy-tree`: Recursive tree with getChildren/getDescendants/getSubtree; arbitrary depth; drill-down UI
- `competency-network-view`: Table of self-rating vs RH rating per competency for a single employee

### Modified Capabilities
- `manager-9x9`: Extends existing spec — adds director-general profile, multi-level hierarchy scope, competency network sub-view

## Approach

**PR 1 — Foundation (~350 lines)**
- Types: `nine-box.ts`, `org-hierarchy.ts`
- Stores: `nineBoxStore.svelte.ts`, `orgHierarchyStore.svelte.ts`
- Fixtures: `matrix-entries.json`, `quadrant-definitions.json`, `scale-definitions.json`, `org-tree.json`
- Modify: `evaluation.ts` (add `director-general`), `goal.ts` (MANAGER_MAP), `assignments.json` (4-level hierarchy)

**PR 2 — 9×9 Matrix Core (~300 lines)**
- `NineBoxMatrix.svelte`, `NineBoxSliders.svelte`, `NineBoxEntryCard.svelte`
- Replace `/evaluacion/9x9/+page.svelte`
- Depends on PR 1

**PR 3 — Hierarchy + Competency Detail (~280 lines)**
- `OrgHierarchyTree.svelte`, `CompetencyNetworkView.svelte`
- New routes: `/evaluacion/9x9/competencias/[employeeId]`, `/evaluacion/9x9/jerarquia`
- Modify: `menuConfig.ts`, `profileUsers.ts`
- Depends on PR 1; parallel with PR 2

**PR 4 — Profile Views + Polish (~170 lines)**
- Profile-specific matrix views (jefe vs director vs director-general)
- Integration, visual polish, WCAG 2.1 AA
- Depends on PR 2 + PR 3

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/lib/types/nine-box.ts` | New | NineBoxMatrix, NineBoxEntry, NineBoxQuadrant, NineBoxScale |
| `web/src/lib/types/org-hierarchy.ts` | New | OrgNode tree structure |
| `web/src/lib/types/evaluation.ts` | Modified | Add `director-general` to EvaluationProfile |
| `web/src/lib/types/goal.ts` | Modified | Add `director-general` to MANAGER_MAP |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | New | Matrix state, quadrant calculation, CRUD |
| `web/src/lib/stores/orgHierarchyStore.svelte.ts` | New | Tree traversal: getChildren, getDescendants, getSubtree |
| `web/src/lib/fixtures/nine-box/*` | New | matrix-entries, quadrant-definitions, scale-definitions |
| `web/src/lib/fixtures/org-hierarchy/org-tree.json` | New | 4-level tree: DG → Director → Jefe → Colaborador |
| `web/src/lib/components/nine-box/NineBoxMatrix.svelte` | New | 9×9 grid with quadrant colors and employee dots |
| `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` | New | Detail popup for employee entry |
| `web/src/lib/components/nine-box/NineBoxSliders.svelte` | New | Dual slider for performance/potential |
| `web/src/lib/components/org-hierarchy/OrgHierarchyTree.svelte` | New | Recursive drill-down component |
| `web/src/lib/components/evaluation/CompetencyNetworkView.svelte` | New | Competency table for single employee |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Modified | Replace placeholder with matrix |
| `web/src/routes/evaluacion/9x9/competencias/[employeeId]/+page.svelte` | New | Competency network view |
| `web/src/routes/evaluacion/9x9/jerarquia/+page.svelte` | New | Hierarchy drill-down view |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Fixture hierarchy doesn't match real org | Medium | Make org-tree.json easily swappable; document structure |
| 9×9 + hierarchy + competency scope too large | High | Strict PR budget (~800 lines); defer graph to future |
| Performance issue with large tree rendering | Low | Lazy-load subtrees; virtualize if >100 nodes |

## Rollback Plan

- Revert file additions per PR; PR 1 is foundation (revert first)
- PR 2/3 can be reverted independently if matrix or hierarchy diverge
- PR 4 is additive polish; lowest risk

## Dependencies

- **A1**: UI shell and design tokens
- **A2**: Competency framework (pillars, competencies, scales)
- **A3/A4**: Goals (fijación/evaluación)
- **A5**: Annual evaluation UI (evaluation lifecycle)
- **B1–B5**: OpenAPI contracts (stub only, not implemented)

## Success Criteria

- [ ] Director-general can drill from DG → Director → Jefe → Colaborador
- [ ] 9×9 matrix shows all employees under viewer's hierarchy scope
- [ ] CompetencyNetworkView shows self-rating vs RH rating for single employee
- [ ] Profile-based menu items: jefe sees matrix for direct reports only; director sees all in their span
- [ ] All PRs pass lint and validation
- [ ] WCAG 2.1 AA compliance on all new components