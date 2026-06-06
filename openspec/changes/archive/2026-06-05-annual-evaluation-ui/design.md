# Design: A5 — Annual Evaluation UI

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│  Pages                                                          │
│  /mi-evaluacion   /rh/evaluaciones   /mis-evaluados             │
│  (employee)       (RH formal)        (manager view)             │
├────────────┬────────────────┬───────────────────────────────────┤
│ evaluation │ competency     │ goals                             │
│ Store      │ Store          │ Store                             │
│ (NEW)      │ (existing)     │ (modified)                        │
├────────────┴────────────────┴───────────────────────────────────┤
│  Types: evaluation-result.ts (NEW)  goal.ts (MOD)  competency.ts│
├─────────────────────────────────────────────────────────────────┤
│  Components: ScaleRatingSelector, CompetencyRatingCard,         │
│  GoalClosureCard, ComparisonTable, EvaluationStatusBadge        │
│  Reused: AssigneePicker, ProgressIndicator, CategoryCard,       │
│  GoalRow, EmptyState, AppShell                                  │
├─────────────────────────────────────────────────────────────────┤
│  Fixtures: evaluations/self-evaluations.json                    │
│            evaluations/rh-evaluations.json                      │
│            evaluations/goal-closures.json                       │
└─────────────────────────────────────────────────────────────────┘
```

## Types

### New: `web/src/lib/types/evaluation-result.ts`

```ts
import type { EvaluationProfile } from './evaluation';

export type EvaluationStatus = 'pending' | 'in-progress' | 'completed';

export interface CompetencyRating {
  id: string;
  employeeId: string;
  competencyId: string;
  selfRating?: 1 | 2 | 3 | 4 | 5;
  selfComment?: string;
  rhRating?: 1 | 2 | 3 | 4 | 5;
  rhComment?: string;
}

export interface GoalClosure {
  id: string;
  employeeId: string;
  goalId: string;
  finalProgress: number;
  selfAssessment?: string;
  rhAssessment?: string;
  managerComment?: string;   // manager feedback on goal closure
  closedAt?: string;         // ISO timestamp, set when selfAssessment saved
}
```

### Modified: `goalsStore.svelte.ts`

`getGoalPermissions()` return type extended with `canClose: boolean`. New `'fin-anio'` phase branch:

| Role | canEditProgress | canComment | canEditWeight | canDelete | canClose |
|------|:-:|:-:|:-:|:-:|:-:|
| Owner (fin-anio) | false | false | false | false | true |
| Non-owner (fin-anio) | false | false | false | false | false |

`deleteCategory` / `deleteGoal` phase guard extended: block unless `phase === 'inicio-anio'` (currently only blocks `'medio-anio'`).

## Store: `evaluationStore.svelte.ts`

### State

```ts
let competencyRatings = $state<CompetencyRating[]>(structuredClone(selfEvaluationsData));
let goalClosures = $state<GoalClosure[]>(structuredClone(goalClosuresData));
```

### Getters

| Function | Signature | Notes |
|----------|-----------|-------|
| `getCompetencyRatings` | `(employeeId: string) => CompetencyRating[]` | Filter by employee |
| `getCompetencyRating` | `(employeeId: string, competencyId: string) => CompetencyRating \| undefined` | Single cell |
| `getGoalClosures` | `(employeeId: string) => GoalClosure[]` | Filter by employee |
| `getGoalClosure` | `(employeeId: string, goalId: string) => GoalClosure \| undefined` | Single cell |
| `getEvaluationStatus` | `(employeeId: string) => EvaluationStatus` | Computed: 0 rated = pending, partial = in-progress, all 8 + all goals closed = completed |

### Mutations (all gated: `getPhase() === 'fin-anio'`)

| Function | Guard | Side effects |
|----------|-------|-------------|
| `rateCompetency(employeeId, competencyId, level, comment?)` | phase check | Upsert CompetencyRating |
| `rhRateCompetency(employeeId, competencyId, level, comment?)` | phase + RH profile | Upsert rhRating/rhComment |
| `closeGoal(employeeId, goalId, finalProgress, selfAssessment)` | phase check | Set fields + `closedAt` |
| `rhAssessGoal(employeeId, goalId, rhAssessment)` | phase + RH profile | Set rhAssessment |
| `addManagerComment(employeeId, goalId, comment)` | phase check | Set managerComment |

## Components

### New

| Component | Props | Est. lines |
|-----------|-------|:----------:|
| `ScaleRatingSelector` | `value?(1-5)`, `onChange(level)`, `acceptanceLevel?(1-5)`, `disabled?` | ~60 |
| `CompetencyRatingCard` | `pillar`, `competencies[]`, `ratings[]`, `onRate(compId, level, comment?)`, `mode: 'self'\|'rh'`, `profileId`, `disabled?` | ~80 |
| `GoalClosureCard` | `goal`, `kpis[]`, `closure?`, `onSave(finalProgress, selfAssessment)`, `mode: 'self'\|'rh'\|'manager'`, `canEdit?` | ~90 |
| `ComparisonTable` | `ratings[]`, `profileId` | ~70 |
| `EvaluationStatusBadge` | `status: EvaluationStatus` | ~20 |

### Modified

| Component | Change |
|-----------|--------|
| `GoalRow` | Add `phase === 'fin-anio'` branch: show `ProgressIndicator` (read-only for non-owner) + "Avance final" label. Add `canClose` permission handling. Hide delete button when `phase !== 'inicio-anio'`. |
| `CategoryCard` | Extend phase guard: hide delete/add buttons when `phase !== 'inicio-anio'`. Pass `canClose` through. |

## Pages

### `/mi-evaluacion` — Employee Self-Evaluation

**Phase guard**: If `phase !== 'fin-anio'`, render `EmptyState` "Evaluación no disponible hasta fin de año".

**Layout** (when phase = `'fin-anio'`):
1. Header: "Mi evaluación" + `EvaluationStatusBadge` for current user
2. Section 1 — Competencias: iterate pillars → `CompetencyRatingCard` (mode='self')
3. Section 2 — Cierre de metas: iterate categories → `CategoryCard` with `phase='fin-anio'`, each goal rendered as `GoalClosureCard` (mode='self')

**Data flow**: `getProfile()` → `getAssignmentsByProfile()` → employeeId → `evaluationStore` getters/setters.

### `/rh/evaluaciones` — RH Formal Evaluation

**Phase guard**: Same as above.

**Layout**:
1. Header: "Evaluaciones RH" + `AssigneePicker` (all 8 employees) + `EvaluationStatusBadge` for selected
2. Section 1 — `ComparisonTable` (self vs RH vs acceptance, color-coded gap)
3. Section 2 — Competencias: `CompetencyRatingCard` (mode='rh') per pillar
4. Section 3 — Cierre de metas: `GoalClosureCard` (mode='rh') per goal

**Data flow**: RH selects employee → `evaluationStore.getCompetencyRatings(selectedId)` + `getGoalClosures(selectedId)`.

### `/mis-evaluados` — Manager View

**Phase guard**: Same `EmptyState` when `phase !== 'fin-anio'`.

**Layout**:
1. Header: "Mis evaluados" + `AssigneePicker` (direct reports only, via inverse `MANAGER_MAP`)
2. Section 1 — Resumen: `EvaluationStatusBadge` + competency summary (read-only `ComparisonTable` without RH column)
3. Section 2 — Competencias: read-only view of each competency with self-rating (display only, no rating inputs)
4. Section 3 — Cierre de metas: `GoalClosureCard` (mode='manager') per goal — shows final progress + self-assessment (read-only) + manager comment textarea (editable)

**Manager permissions**:
- See evaluatee's self-ratings: **yes** (read-only)
- Rate competencies: **no** (RH only)
- See RH ratings: **yes** (read-only, after RH completes)
- Add goal feedback: **yes** (`managerComment` via `evaluationStore.addManagerComment`)

**Data flow**: `getProfile()` → inverse `MANAGER_MAP` → `subordinateProfiles` → `getAssignmentsByProfile()` → `AssigneePicker` → `evaluationStore` getters.

## Fixtures

### `web/src/lib/fixtures/evaluations/self-evaluations.json`

3 employees with mixed states:
- `emp-colaborador-01`: 5/8 competencies rated (in-progress)
- `emp-vendedor-01`: 8/8 rated (completed competencies)
- `emp-jefe-01`: 0 rated (pending)

### `web/src/lib/fixtures/evaluations/goal-closures.json`

Matching closures for above employees, mixed `closedAt` states.

### `web/src/lib/fixtures/evaluations/rh-evaluations.json`

RH ratings for `emp-vendedor-01` (completed RH eval), others empty.

## Integration Points

| Dependency | How A5 uses it |
|------------|---------------|
| `competencyStore` | `getPillars()`, `getCompetenciesByPillar()`, `getCompetencyAcceptanceLevel()`, `getLevelDefinition()` |
| `goalsStore` | `getCategories()`, `getGoalsByCategory()`, `getKpisForGoal()`, `getGoalPermissions()`, `updateGoalProgress()` |
| `devContext` | `getPhase()` for phase guard, `getProfile()` for role detection |
| `AssigneePicker` | RH picker (all assignments), Manager picker (subordinate assignments) |
| `CategoryCard` / `GoalRow` | Reused with `phase='fin-anio'` and new `canClose` permission |
| `menuConfig.ts` | `/mis-evaluados` already configured for manager profiles |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `getGoalPermissions` signature change breaks existing callers | Add `canClose` as optional with default `false`; existing `medio-anio` / `inicio-anio` branches unchanged |
| Manager view scope creep (rating, editing) | Explicitly read-only; `mode='manager'` on `GoalClosureCard` only enables comment textarea |
| Fixture data inconsistency across 8 profiles | Seed 3 key profiles; others show `pending` status by absence from fixture |
| Phase naming (`'cierre'` vs `'fin-anio'`) | Codebase uses `'fin-anio'`; design follows existing `CyclePhase` type — no rename needed |
