# Design: data-model-core

## Overview

This document defines the Ent schema for all SED entities. Each schema file lives in `api/internal/schema/` and follows Ent conventions: `entc` annotations, explicit edges, composite indices.

## Schema Inventory

### Org Hierarchy

#### `Organization` (root/tenant)

```go
// Organization schema — top-level tenant or company entity
fields:
  - id: uuid, key
  - name: string (required)
  - slug: string (unique, required)
  - created_at: datetime
  - updated_at: datetime

edges:
  - to: OrgNode (1:N, cascade)
```

#### `OrgNode` (node in the org tree)

```go
// OrgNode — a node in the organizational tree
fields:
  - id: uuid, key
  - organization_id: uuid (FK, required)
  - parent_id: uuid (FK, optional — self-ref for tree)
  - name: string (required)
  - type: enum OrgNodeType (corporate | retail) // decision #2
  - code: string (unique within org)
  - metadata: jsonb (flexible extra data)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Organization: Organization (M:1)
  - Parent: OrgNode (O2M, back="Children")
  - Children: OrgNode (O2M)
  - Employees: Employee (1:N, cascade)

indices:
  - (organization_id, parent_id) — children listing
  - (organization_id, type) — filter by org type
```

#### `Employee`

```go
// Employee — person in the org chart
fields:
  - id: uuid, key
  - org_node_id: uuid (FK, required)
  - manager_id: uuid (FK, optional — self-ref for reporting line)
  - profile_id: uuid (FK, required → EvaluationProfile)
  - first_name: string
  - last_name: string
  - employee_number: string (unique per org)
  - email: string (unique)
  - is_active: bool (default true)
  - created_at: datetime
  - updated_at: datetime
  - created_by: uuid
  - updated_by: uuid

edges:
  - OrgNode: OrgNode (M:1)
  - Manager: Employee (O2M, back="DirectReports")
  - DirectReports: Employee (O2M)
  - Profile: EvaluationProfile (M:1)
  - GoalAssignments: GoalAssignment (1:N)
  - Evaluations: Evaluation (1:N)
  - NineBoxMatrices: NineBoxMatrix (1:N) // as evaluator
  - NineBoxEntries: NineBoxEntry (1:N) // as evaluatee

indices:
  - (org_node_id, manager_id) — reporting line queries
  - (profile_id) — filter by evaluation profile
  - (manager_id, is_active) — "my evaluatees" listing
```

#### `EvaluatorScope`

```go
// EvaluatorScope — defines what a manager can see/edit for their evaluatees
fields:
  - id: uuid, key
  - evaluator_id: uuid (FK → Employee)
  - scope_type: enum ScopeType (department | team | individual)
  - scope_data: jsonb (stores org_node_ids or employee_ids depending on scope_type)
  - cycle_id: uuid (FK → Cycle, optional — null means all cycles)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Evaluator: Employee (M:1)
  - Cycle: Cycle (M:1, optional)

indices:
  - (evaluator_id, cycle_id) — look up scope by evaluator + cycle
```

---

### Evaluation Lifecycle

#### `Cycle`

```go
// Cycle — annual evaluation cycle (one per year)
fields:
  - id: uuid, key
  - year: int (YYYY, unique per org)
  - organization_id: uuid (FK)
  - current_phase: enum Phase (asignacion | avance | cierre)
  - started_at: datetime (nullable)
  - finished_at: datetime (nullable)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Organization: Organization (M:1)
  - PhaseTransitions: PhaseTransition (1:N)
  - Phases: PhaseDefinition (1:N via transition table)
  - NineBoxMatrices: NineBoxMatrix (1:N)

indices:
  - (organization_id, year) — unique, for listing active cycle
  - (current_phase) — filter by phase
```

#### `PhaseDefinition`

```go
// PhaseDefinition — catalog of the 3 phases (static, not user-editable)
fields:
  - id: uuid, key
  - phase: enum Phase (asignacion | avance | cierre)
  - label: string (e.g., "Inicio de año")
  - order: int (1–3)
  - allowed_actors: jsonb (array of profile strings)
  - allowed_actions: jsonb (array of action strings)
  - blocked_actions: jsonb (array of action strings)

edges:
  - Cycle: Cycle (M:1)
  - OutgoingTransitions: PhaseTransition (1:N, back="FromPhase")
  - IncomingTransitions: PhaseTransition (1:N, back="ToPhase")
```

#### `PhaseTransition`

```go
// PhaseTransition — defines valid phase transitions
fields:
  - id: uuid, key
  - from_phase: enum Phase
  - to_phase: enum Phase
  - trigger: enum TriggerType (auto | manual_rh)
  - conditions: jsonb (optional — e.g., date constraints)
  - created_at: datetime

edges:
  - FromPhase: PhaseDefinition (M:1)
  - ToPhase: PhaseDefinition (M:1)

indices:
  - (from_phase, to_phase) — unique, ensure no duplicate transition definitions
```

---

### Competency Framework

#### `Pillar`

```go
// Pillar — unique catalog of pillars for all profiles (decision #2)
fields:
  - id: uuid, key
  - name: string (unique)
  - description: text
  - created_at: datetime
  - updated_at: datetime

edges:
  - Competencies: Competency (1:N, cascade)

indices:
  - (name) — unique
```

#### `Competency`

```go
// Competency — belongs to one pillar
fields:
  - id: uuid, key
  - pillar_id: uuid (FK, required)
  - name: string
  - description: text
  - created_at: datetime
  - updated_at: datetime

edges:
  - Pillar: Pillar (M:1)
  - ScaleCriteria: ScaleCriterion (1:N, cascade)
  - AcceptanceLevels: CompetencyAcceptanceLevel (1:N, cascade)

indices:
  - (pillar_id) — listing competencies by pillar
```

#### `ScaleCriterion`

```go
// ScaleCriterion — descriptive criterion for a competency × level
fields:
  - id: uuid, key
  - competency_id: uuid (FK, required)
  - pillar_id: uuid (FK, required) // denormalized for query efficiency
  - level: int (1–5)
  - description: text
  - created_at: datetime
  - updated_at: datetime

edges:
  - Competency: Competency (M:1)

indices:
  - (competency_id, pillar_id, level) — lookup criteria for a cell
```

#### `LevelDefinition`

```go
// LevelDefinition — global 1–5 scale definitions (shared by all profiles/competencies)
fields:
  - level: int (1–5, key)
  - label: string (e.g., "No aceptable", "En desarrollo", "Cumple", "Supera", "Excepcional")
  - description: text

indices:
  - (level) — primary, for ordering
```

#### `EvaluationProfile`

```go
// EvaluationProfile — the 8 evaluation profiles (decision #2)
fields:
  - id: uuid, key
  - name: string (unique — e.g., "colaborador", "jefe", "vendedor", etc.)
  - description: text

edges:
  - Employees: Employee (1:N)
  - AcceptanceLevels: CompetencyAcceptanceLevel (1:N)
```

#### `CompetencyAcceptanceLevel`

```go
// CompetencyAcceptanceLevel — minimum acceptable level per competency × profile
fields:
  - id: uuid, key
  - competency_id: uuid (FK)
  - profile_id: uuid (FK)
  - level: int (1–5)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Competency: Competency (M:1)
  - Profile: EvaluationProfile (M:1)

indices:
  - (competency_id, profile_id) — unique, one acceptance level per pair
```

---

### Goals and Weighting

#### `GoalCategory`

```go
// GoalCategory — custom category defined by an employee (independent of pillars, decision #5)
fields:
  - id: uuid, key
  - employee_id: uuid (FK — owner, decision #8)
  - name: string
  - description: text
  - weight: float (0–100)
  - created_at: datetime
  - updated_at: datetime
  - created_by: uuid
  - updated_by: uuid

edges:
  - Employee: Employee (M:1)
  - Goals: Goal (1:N, cascade)

indices:
  - (employee_id) — all categories for an employee
  - (employee_id, name) — unique, no duplicate category names per employee
```

#### `Goal`

```go
// Goal — a single goal within a category (decision #1: only goals weight, sum to 100%)
fields:
  - id: uuid, key
  - category_id: uuid (FK, required)
  - name: string
  - description: text
  - unit: enum GoalUnit (porcentaje | moneda | numero)
  - weight: float (0–100)
  - target_value: float (> 0)
  - current_value: float (default 0)
  - state: enum GoalState (borrador | fijada | en_seguimiento | evaluada | cerrada)
  - created_at: datetime
  - updated_at: datetime
  - created_by: uuid
  - updated_by: uuid

edges:
  - Category: GoalCategory (M:1)
  - KpiLinks: GoalKpiLink (1:N, cascade)
  - EvaluationGoals: EvaluationGoal (1:N)

indices:
  - (category_id) — goals in a category
  - (state) — filter goals by state
  - (created_by) — audit
```

#### `KPI`

```go
// KPI — reusable indicator, independent of any goal (decision #6)
fields:
  - id: uuid, key
  - name: string
  - unit: enum GoalUnit (porcentaje | moneda | numero)
  - description: text
  - created_at: datetime
  - updated_at: datetime

edges:
  - GoalLinks: GoalKpiLink (1:N, cascade)

indices:
  - (name) — unique
```

#### `GoalKpiLink`

```go
// GoalKpiLink — N:M relationship between goals and KPIs (decision #6)
fields:
  - goal_id: uuid (FK)
  - kpi_id: uuid (FK)
  - created_at: datetime

edges:
  - Goal: Goal (M:1)
  - Kpi: KPI (M:1)

indices:
  - (goal_id, kpi_id) — unique, no duplicate links
  - (kpi_id) — reverse lookup: which goals use this KPI
```

#### `GoalAssignment`

```go
// GoalAssignment — maps an employee to their categories and goals for a cycle
fields:
  - id: uuid, key
  - employee_id: uuid (FK)
  - cycle_id: uuid (FK)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Employee: Employee (M:1)
  - Cycle: Cycle (M:1)
  - Categories: GoalCategory (1:N)
  - Goals: Goal (1:N via category)

indices:
  - (employee_id, cycle_id) — unique, one assignment per employee per cycle
```

---

### 9×9 Matrix

#### `NineBoxMatrix`

```go
// NineBoxMatrix — one matrix per evaluator (manager) per cycle
fields:
  - id: uuid, key
  - cycle_id: uuid (FK, required)
  - evaluator_id: uuid (FK → Employee)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Cycle: Cycle (M:1)
  - Evaluator: Employee (M:1)
  - Entries: NineBoxEntry (1:N, cascade)

indices:
  - (cycle_id, evaluator_id) — unique, one matrix per evaluator per cycle
  - (evaluator_id) — all matrices for an evaluator
```

#### `NineBoxEntry`

```go
// NineBoxEntry — rating for one evaluatee in the 9x9 matrix
fields:
  - id: uuid, key
  - matrix_id: uuid (FK, required)
  - evaluatee_id: uuid (FK → Employee)
  - performance_score: int (1–9)
  - potential_score: int (1–9)
  - quadrant: int (computed on write — 1 through 9)
  - comments: text (optional)
  - created_at: datetime
  - updated_at: datetime
  - updated_by: uuid

edges:
  - Matrix: NineBoxMatrix (M:1)
  - Evaluatee: Employee (M:1)

indices:
  - (matrix_id, evaluatee_id) — unique
  - (evaluatee_id) — which matrices include a given evaluatee
```

#### `NineBoxQuadrant`

```go
// NineBoxQuadrant — catalog of the 9 quadrants
fields:
  - id: uuid, key
  - quadrant: int (1–9, key)
  - label: string
  - description: text
  - color: string (hex color for UI)
  - action_recommendation: text

indices:
  - (quadrant) — primary
```

#### `NineBoxScale`

```go
// NineBoxScale — definitions of the 9 levels per axis
fields:
  - id: uuid, key
  - axis: enum Axis (performance | potential)
  - level: int (1–9, key)
  - label: string
  - description: text

indices:
  - (axis, level) — primary
```

---

### Evaluation

#### `Evaluation`

```go
// Evaluation — the main evaluation record for an employee in a cycle
fields:
  - id: uuid, key
  - employee_id: uuid (FK)
  - cycle_id: uuid (FK)
  - phase: enum Phase
  - state: enum EvaluationState (pendiente_asignacion | pendiente_avance | pendiente_evaluacion_final | completada)
  - self_evaluation_completed_at: datetime (nullable)
  - rh_evaluation_completed_at: datetime (nullable)
  - created_at: datetime
  - updated_at: datetime
  - created_by: uuid
  - updated_by: uuid

edges:
  - Employee: Employee (M:1)
  - Cycle: Cycle (M:1)
  - CompetencyRatings: EvaluationCompetency (1:N, cascade)
  - GoalRatings: EvaluationGoal (1:N)

indices:
  - (employee_id, cycle_id) — unique, one evaluation per employee per cycle
  - (cycle_id, state) — listings by cycle and state
  - (cycle_id, phase) — filter by phase
```

#### `EvaluationCompetency`

```go
// EvaluationCompetency — rating of one competency for an evaluation
fields:
  - id: uuid, key
  - evaluation_id: uuid (FK)
  - competency_id: uuid (FK)
  - profile_id: uuid (FK — the profile being evaluated)
  - rating: int (1–5) — the rating given
  - comments: text (optional)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Evaluation: Evaluation (M:1)
  - Competency: Competency (M:1)
  - Profile: EvaluationProfile (M:1)

indices:
  - (evaluation_id, competency_id) — unique
```

#### `EvaluationGoal`

```go
// EvaluationGoal — final rating of a goal within an evaluation
fields:
  - id: uuid, key
  - evaluation_id: uuid (FK)
  - goal_id: uuid (FK)
  - final_rating: int (1–5, nullable) — only set at cierre
  - final_comments: text (optional)
  - created_at: datetime
  - updated_at: datetime

edges:
  - Evaluation: Evaluation (M:1)
  - Goal: Goal (M:1)

indices:
  - (evaluation_id, goal_id) — unique
```

---

## Indices Summary

| Index | Table | Columns | Purpose |
|-------|-------|---------|---------|
| `idx_employee_manager_active` | `employees` | `(manager_id, is_active)` | "My evaluatees" listing |
| `idx_employee_org_profile` | `employees` | `(org_node_id, profile_id)` | Filter by org node and profile |
| `idx_cycle_org_year` | `cycles` | `(organization_id, year)` | Unique cycle lookup |
| `idx_goal_category` | `goals` | `(category_id)` | Goals in a category |
| `idx_goal_state` | `goals` | `(state)` | Goals filtered by state |
| `idx_goalassignment_employee_cycle` | `goal_assignments` | `(employee_id, cycle_id)` | Unique assignment |
| `idx_nineboxentry_matrix_evaluatee` | `nine_box_entries` | `(matrix_id, evaluatee_id)` | Unique entry |
| `idx_evaluation_employee_cycle` | `evaluations` | `(employee_id, cycle_id)` | Unique evaluation |
| `idx_evaluation_cycle_state` | `evaluations` | `(cycle_id, state)` | List by cycle and state |
| `idx_scalecriterion_cell` | `scale_criteria` | `(competency_id, pillar_id, level)` | Scale criterion lookup |
| `idx_competencyacceptance_cp` | `competency_acceptance_levels` | `(competency_id, profile_id)` | Unique per pair |
| `idx_goal_kpi_link` | `goal_kpi_links` | `(kpi_id)` | Reverse KPI lookup |
| `idx_evaluatorscope_eval_cycle` | `evaluator_scopes` | `(evaluator_id, cycle_id)` | Scope lookup |

## Migration Strategy

1. Generate Ent schemas → `api/internal/schema/`
2. Run `ent generate ./...` to produce code
3. Use **Atlas** (`atlas migrate diff`) or **golang-migrate** to generate versioned SQL migrations in `api/migrations/`
4. Apply migrations: `atlas migrate apply` or `migrate -path api/migrations up`
5. All migrations are **versioned** and **repeatable** — no manual schema edits after apply

## PostgreSQL-Specific Considerations

- **Enums** created as `CREATE TYPE` before referencing tables
- **JSONB** for `metadata`, `allowed_actions`, `scope_data` fields — with GIN index on `scope_data` for evaluator scope queries
- **Partial indices**: e.g., only active employees `(is_active = true)` for reporting line queries
- **Cascade delete**: explicit on all composition edges (e.g., competency → scale_criteria, category → goals)
- **No circular FKs**: org hierarchy uses self-referential `parent_id` without cycles via deferrable constraints if needed
- **UUID primary keys**: `uuid_generate_v4()` as default

## Out of Scope

See `proposal.md` — Non-goals apply to design as well.