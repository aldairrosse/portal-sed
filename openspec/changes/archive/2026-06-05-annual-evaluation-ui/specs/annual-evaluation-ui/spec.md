# Spec: Annual Evaluation UI (A5)

## Purpose

Year-end evaluation at `/mi-evaluacion` (employee self-rating + goal closure) and `/rh/evaluaciones` (RH formal evaluation with side-by-side comparison). All mutations gated to `cyclePhase === 'fin-anio'`. Fixture-driven; no backend.

## Scope

| In Scope | Out of Scope |
|----------|-------------|
| Competency self-rating: 3 pillars, 8 competencies, scale 1–5, acceptance level reference, optional comment per competency | 9×9 potential matrix (A6) |
| Goal closure: final progress + self-assessment per goal, status freeze | Backend persistence, auth, email |
| RH evaluation: competency ratings, self-rating comparison, RH assessment | Login/authentication |
| Comparison table: self vs RH vs acceptance level, color-coded gap | |
| Evaluation status badges: `pending` / `in-progress` / `completed` | |

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **CompetencyRating** | `employeeId`, `competencyId`, `selfRating?`(1–5), `selfComment?`, `rhRating?`(1–5), `rhComment?` | Per employee × competency |
| **GoalClosure** | `employeeId`, `goalId`, `finalProgress`(number), `selfAssessment?`, `rhAssessment?`, `closedAt`(ISO) | Per goal |
| **EvaluationStatus** | `pending` \| `in-progress` \| `completed` | Computed per employee |

## Requirements

| ID | Name | Priority |
|----|------|----------|
| R1 | Phase guard | Must |
| R2 | Competency self-rating | Must |
| R3 | Goal closure | Must |
| R4 | RH formal evaluation | Must |
| R5 | Comparison table | Must |
| R6 | Evaluation status badges | Should |

### R1: Phase Guard

The system SHALL block all evaluation mutations unless `cyclePhase === 'fin-anio'`. UI SHALL render `EmptyState` "Evaluación no disponible hasta fin de año" otherwise.

#### Scenario: Phase mismatch blocks access

- GIVEN `cyclePhase !== 'fin-anio'`
- WHEN navigating to `/mi-evaluacion`
- THEN inputs are disabled and `EmptyState` message is shown

#### Scenario: Correct phase enables evaluation

- GIVEN `cyclePhase === 'fin-anio'`
- WHEN employee accesses `/mi-evaluacion`
- THEN all rating inputs and goal closure fields are editable

### R2: Competency Self-Rating

The system SHALL render each pillar with its competencies as `CompetencyRatingCard` components. Employee SHALL select a rating 1–5 per competency via `ScaleRatingSelector` (radio buttons with level labels: No aceptable → Excepcional). Acceptance level (per competency × profile) SHALL be displayed as reference. Optional comment per competency.

#### Scenario: Employee rates a competency

- GIVEN employee on `/mi-evaluacion`, pillar "Liderazgo", competency "Comunicación efectiva"
- WHEN selects rating 4
- THEN `ScaleRatingSelector` highlights level 4 with `btn-primary`
- AND acceptance level (e.g., 3) is shown as `badge-ghost` "Mínimo esperado: Cumple (3)"

#### Scenario: Partial completion tracked

- GIVEN employee rates 5 of 8 competencies
- WHEN `EvaluationStatus` is computed
- THEN status is `in-progress` with `badge-warning`

#### Scenario: All ratings complete

- GIVEN employee rates all 8 competencies and closes all goals
- WHEN `EvaluationStatus` is computed
- THEN status is `completed` with `badge-success`

### R3: Goal Closure

The system SHALL render each goal as `GoalClosureCard`: read-only goal info (name, weight, target, KPIs), editable `finalProgress` (reuses `progress` field from A4), and `selfAssessment` textarea. On submit, goal is frozen (read-only).

#### Scenario: Employee sets final progress

- GIVEN goal with mid-year progress 65% and target 100
- WHEN employee updates `finalProgress` to 90
- THEN `ProgressIndicator` shows 90% bar with `progress-success`
- AND `GoalClosure.finalProgress = 90`

#### Scenario: Employee adds self-assessment

- GIVEN goal closure form open
- WHEN employee writes "Objetivo superado, se logró 90% a pesar de recortes" and saves
- THEN `GoalClosure.selfAssessment` is set
- AND the goal becomes read-only

### R4: RH Formal Evaluation

RH SHALL select an employee via `AssigneePicker` on `/rh/evaluaciones`. For each competency, RH SHALL rate 1–5, see employee's self-rating displayed adjacent, and optionally add `rhComment`. RH SHALL add `rhAssessment` per goal. Self-rating values SHALL be read-only in this view.

#### Scenario: RH rates competency with self-rating visible

- GIVEN RH selects employee "María López" in `/rh/evaluaciones`
- WHEN viewing competency "Comunicación efectiva"
- THEN shows: "Autoevaluación: 4" (read-only) next to RH rating selector (1–5)
- WHEN RH selects 3
- THEN `CompetencyRating.rhRating = 3` and self-rating stays 4

#### Scenario: RH adds goal assessment

- GIVEN RH viewing employee goal closure
- WHEN enters "Aceptable, pero pudo alcanzar 95%" in `rhAssessment` field
- THEN `GoalClosure.rhAssessment` is saved

### R5: Comparison Table

The system SHALL render a comparison table in RH view with columns: Competency | Self-Rating | RH Rating | Acceptance Level | Gap. Gap SHALL be color-coded: green (self ≥ acceptance), red (self < acceptance), amber (RH differs from self by ≥ 2).

#### Scenario: Comparison with all data

- GIVEN employee self-rating = 4, RH rating = 3, acceptance level = 3
- WHEN comparison table renders
- THEN row shows: 4 | 3 | 3 | gap = 0 (green)

#### Scenario: Gap detected

- GIVEN employee self-rating = 2, acceptance level = 3
- WHEN comparison table renders
- THEN gap cell shows `badge-error` "−1 (por debajo)"

### R6: Evaluation Status Badges

The system SHALL compute and display `EvaluationStatus` per employee in the RH employee picker.

| Status | Condition | DaisyUI Class |
|--------|-----------|---------------|
| `pending` | No self-ratings submitted | `badge-ghost` |
| `in-progress` | Some ratings or goals incomplete | `badge-warning` |
| `completed` | All 8 competencies rated + all goals closed | `badge-success` |

## UI Components

| Component | Props | Description |
|-----------|-------|-------------|
| `ScaleRatingSelector` | `value`(1–5), `onChange`, `acceptanceLevel` | Radio group with level labels |
| `CompetencyRatingCard` | `pillar`, `competencies[]`, `ratings`, `onRate` | Pillar-section with competencies |
| `GoalClosureCard` | `goal`, `progress`, `onSaveClosure` | Read-only goal + progress + textarea |
| `ComparisonTable` | `ratings[]`, `acceptanceLevels` | Side-by-side self/RH/acceptance |
| `EvaluationStatusBadge` | `status` | DaisyUI badge by status |

Reused: `AssigneePicker`, `ProgressIndicator`, `EmptyState`, `PageSkeleton`, `AppShell`, `Sidebar`.

## Behavior

- **Store**: `evaluationStore.svelte.ts` holds `Map<employeeId, CompetencyRating[]>` and `Map<employeeId, GoalClosure[]>`. Mutations: `rateCompetency`, `closeGoal`, `rhRateCompetency`, `rhAssessGoal`. All gated by phase.
- **Permissions**: Employee → self-rating + goal closure. RH → all employees + formal evaluation. Jefe → view-only in `fin-anio`.
- **Fixture files**: `evaluations/self-evaluations.json` (3 employees), `evaluations/rh-evaluations.json`, `evaluations/goal-closures.json`.

## Dependencies

- `competencyStore` (A2): pillars, competencies, acceptance levels, scale criteria
- `goalsStore` (A3/A4): goals, progress, `getGoalPermissions()`
- `AssigneePicker`, `ProgressIndicator` (existing)
- `cycleStore` / `cycle.json`: phase state

## Open Questions

None.
