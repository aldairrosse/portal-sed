# Tasks: A5 — Annual Evaluation UI

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~900–1050 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 → PR 4 → PR 5 |
| Delivery strategy | auto-chain |
| Chain strategy | pending |
| Decision needed before apply | No |

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation (types, store, goalsStore mods, cycle.json) | PR 1 | Base for all evaluation pages |
| 2 | Employee self-evaluation (components + page) | PR 2 | Depends on PR 1; `/mi-evaluacion` page |
| 3 | RH formal evaluation (components + page) | PR 3 | Depends on PR 1; `/rh/evaluaciones` page |
| 4 | Manager view (page) | PR 4 | Depends on PR 1, 2; `/mis-evaluados` page |
| 5 | Integration (testing, polish, a11y) | PR 5 | Depends on all above |

## Phase 1: Foundation (PR 1)

> Goal: types, stores, fixture scaffolding, and permissions that Phase 2–4 depend on

- [x] 1.1 Create `web/src/lib/types/evaluation-result.ts` — `CompetencyRating`, `GoalClosure`, `EvaluationStatus` types (~60 lines)
- [x] 1.2 Create `web/src/lib/stores/evaluationStore.svelte.ts` — state, getters, and phase-gated mutations (`rateCompetency`, `rhRateCompetency`, `closeGoal`, `rhAssessGoal`, `addManagerComment`, `getEvaluationStatus`) (~120 lines)
- [x] 1.3 Modify `goalsStore.svelte.ts` — add `canClose` to `getGoalPermissions()` return type, add `'fin-anio'` branch (owner: `canClose:true`, others all false), extend `deleteCategory`/`deleteGoal` guard to block `'fin-anio'` as well (~40 lines changed)
- [x] 1.4 Update `cycle.json` — set `"phase"` from `"medio-anio"` to `"fin-anio"` (1 line)
- [x] 1.5 Create `web/src/lib/fixtures/evaluations/self-evaluations.json` — 3 employees with mixed states (pending, in-progress, completed) (~50 lines)
- [x] 1.6 Create `web/src/lib/fixtures/evaluations/goal-closures.json` — matching closures for the same employees (~40 lines)
- [x] 1.7 Create `web/src/lib/fixtures/evaluations/rh-evaluations.json` — RH ratings for 1 completed employee (~30 lines)

**Phase 1 acceptance:**
- [x] `pnpm run check` passes
- [x] `evaluationStore` loads fixture data and `getCompetencyRatings` returns correct rows per employee
- [x] `getGoalPermissions('colaborador', true)` returns `canClose:true` when phase = `'fin-anio'`
- [x] `deleteCategory` / `deleteGoal` are no-ops in `'fin-anio'` phase

## Phase 2: Employee Self-Evaluation (PR 2)

> Goal: `/mi-evaluacion` page with competency self-rating and goal closure

- [x] 2.1 Build `ScaleRatingSelector.svelte` — radio group 1–5 with level labels (No aceptable → Excepcional), acceptance level shown as `badge-ghost`, highlight selected with `btn-primary` (~60 lines)
- [x] 2.2 Build `CompetencyRatingCard.svelte` — pillar section with competencies list, each with `ScaleRatingSelector` + optional comment textarea; props: `pillar`, `competencies[]`, `ratings[]`, `onRate`, `mode: 'self'|'rh'`, `profileId` (~80 lines)
- [x] 2.3 Build `GoalClosureCard.svelte` — read-only goal info + `ProgressIndicator` for `finalProgress` + `selfAssessment` textarea; mode-dependent: `'self'` (editable for owner), `'rh'` (editable for RH), `'manager'` (comment only) (~90 lines)
- [x] 2.4 Modify `GoalRow.svelte` — add `'fin-anio'` branch: show `ProgressIndicator` with "Avance final" label, disable progress input for non-owners, hide delete button when `phase !== 'inicio-anio'` (~40 lines changed)
- [x] 2.5 Modify `CategoryCard.svelte` — extend phase guard: hide delete/add buttons when `phase !== 'inicio-anio'`, pass `canClose` through (~20 lines changed)
- [x] 2.6 Assemble `/mi-evaluacion` page — phase guard with `EmptyState`, Section 1 (competencias via pillars → `CompetencyRatingCard`), Section 2 (cierre de metas via categories → `GoalClosureCard`) (~60 lines)

**Phase 2 acceptance:**
- [x] All 8 competencies render with scale selector and acceptance level reference
- [x] Selecting a rating highlights the level and stores via `evaluationStore.rateCompetency`
- [x] Goals render as `GoalClosureCard` with `finalProgress` input and `selfAssessment` textarea
- [x] Submitting goal closure sets `closedAt` and freezes the card
- [x] `getEvaluationStatus` reflects correct state (pending → in-progress → completed)

## Phase 3: RH Formal Evaluation (PR 3)

> Goal: `/rh/evaluaciones` page with employee picker, competency rating, and comparison

- [x] 3.1 Build `EvaluationStatusBadge.svelte` — renders DaisyUI badge by status: `pending` → `badge-ghost`, `in-progress` → `badge-warning`, `completed` → `badge-success` (~20 lines)
- [x] 3.2 Build `ComparisonTable.svelte` — columns: Competencia | Autoevaluación | RH | Nivel aceptación | Brecha; color-coded gap (green self ≥ acceptance, red self < acceptance, amber |RH−self| ≥ 2) (~70 lines)
- [x] 3.3 Assemble `/rh/evaluaciones` page — phase guard, `AssigneePicker` for all 8 employees, `EvaluationStatusBadge` for selected, `ComparisonTable`, `CompetencyRatingCard` (mode='rh'), `GoalClosureCard` (mode='rh') (~80 lines)

**Phase 3 acceptance:**
- [x] Employee picker shows all 8 profiles with status badges
- [x] Selecting an employee loads their self-ratings and RH can rate each competency
- [x] Self-rating values are displayed read-only next to RH rating selector
- [x] Comparison table shows all data with correct gap coloring
- [x] Only RH profile can see this page in sidebar

## Phase 4: Manager View (PR 4)

> Goal: `/mis-evaluados` page — read-only view of evaluatees' evaluations with manager comment on goal closures

- [x] 4.1 Assemble `/mis-evaluados` page — phase guard, `AssigneePicker` for direct reports only (via inverse `MANAGER_MAP`), read-only competency view, `GoalClosureCard` (mode='manager') with `addManagerComment` integration (~80 lines)

**Phase 4 acceptance:**
- [x] Manager sees only their direct reports in the picker
- [x] Self-ratings and RH ratings are displayed read-only
- [x] Manager can add comments on each goal closure
- [x] Menu item visible for `jefe`, `gerente-tienda`, `divisional`, `regional`, `director` profiles only

## Phase 5: Integration (PR 5)

> Goal: testing, accessibility, and visual polish across all new screens

- [x] 5.1 Verify dev toolbar phase switching — all pages show `EmptyState` when phase ≠ `'fin-anio'` and full UI when phase = `'fin-anio'` (~10 lines)
- [x] 5.2 Test all 8 profiles — confirm correct behavior per role (colaborador/vendedor → self-evaluation, jefe/director → manager view, rh → RH view) (~15 lines)
- [x] 5.3 Visual polish + WCAG 2.1 AA — keyboard navigation on scale selectors, focus management on modals, `aria-label` on all interactive elements, color contrast on gap badges (~15 lines)

**Phase 5 acceptance:**
- [x] Dev toolbar phase switching works across `/mi-evaluacion`, `/rh/evaluaciones`, `/mis-evaluados`
- [x] All 8 profiles tested with fixtures showing correct role-based views
- [x] `pnpm run check` passes with no errors
- [x] Tab navigation works through all rating selectors and form fields
