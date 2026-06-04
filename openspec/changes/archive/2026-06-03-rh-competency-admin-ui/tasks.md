# Tasks: rh-competency-admin-ui (A2)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1180 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation — types, fixtures, store, menu | PR 1 | base = main; ~380 lines; tests included |
| 2 | Pillars & Competencies — tables, forms, routes | PR 2 | base = main; ~400 lines; depends on store from PR 1 |
| 3 | Scale Criteria & Acceptance Levels — matrix, editor | PR 3 | base = main; ~400 lines; depends on store from PR 1 |

## Phase 1: Foundation

- [x] 1.1 Create `web/src/lib/types/competency.ts` — `Pillar`, `Competency`, `ScaleCriterion`, `AcceptanceLevel` interfaces
- [x] 1.2 Create `web/src/lib/fixtures/competency/pillars.json` — 3 seed pillars
- [x] 1.3 Create `web/src/lib/fixtures/competency/competencies.json` — 2-3 competencias per pillar
- [x] 1.4 Create `web/src/lib/fixtures/competency/scale-criteria.json` — level 1-5 per competency×pillar
- [x] 1.5 Create `web/src/lib/fixtures/competency/acceptance-levels.json` — labels per profile×level (8 profiles)
- [x] 1.6 Create `web/src/lib/stores/competencyStore.svelte.ts` — `$state` CRUD + cascade delete
- [x] 1.7 Modify `web/src/lib/nav/menuConfig.ts` — replace "Competencias" with 3 RH sidebar items
- [x] 1.8 Delete `web/src/routes/rh/competencias/+page.svelte` — replaced by pilares entry

## Phase 2: Pilares & Competencias UI

- [x] 2.1 Create `web/src/lib/components/competency/ConfirmDeleteModal.svelte` — shared delete dialog
- [x] 2.2 Create `web/src/lib/components/competency/PillarFormModal.svelte` — create/edit pillar form
- [x] 2.3 Create `web/src/lib/components/competency/PillarTable.svelte` — table with edit/delete
- [x] 2.4 Create `web/src/routes/rh/pilares/+page.svelte` — pillar list page
- [x] 2.5 Create `web/src/lib/components/competency/CompetencyFormModal.svelte` — create/edit competency
- [x] 2.6 Create `web/src/lib/components/competency/CompetencyTable.svelte` — table filtered by pillar
- [x] 2.7 Create `web/src/routes/rh/pilares/[id]/competencias/+page.svelte` — competencies per pillar

## Phase 3: Scale Criteria & Acceptance Levels UI

- [x] 3.1 Create `web/src/lib/components/competency/ScaleCriterionModal.svelte` — edit 5 levels per cell
- [x] 3.2 Create `web/src/lib/components/competency/ScaleCriteriaMatrix.svelte` — competencias×pilares grid
- [x] 3.3 Create `web/src/routes/rh/criterios-escala/+page.svelte` — scale criteria page
- [x] 3.4 Create `web/src/lib/components/competency/AcceptanceLevelEditor.svelte` — profile+level editor
- [x] 3.5 Create `web/src/routes/rh/niveles-aceptacion/+page.svelte` — acceptance levels page

## Phase 4: Testing & Verification

- [x] 4.1 Write unit tests: store CRUD, cascade delete, validation (Vitest)
- [x] 4.2 Verify type safety: `tsc` on fixture imports vs interfaces
- [x] 4.3 Write integration tests: each route renders store data (Testing Library)
- [x] 4.4 Verify: no weight/ponderación input in pillars (decisión #1)
- [x] 4.5 Verify: catálogo visible desde cualquier perfil (decisión #2)
- [x] 4.6 Verify: `pnpm run lint` and `tsc` pass with zero errors
