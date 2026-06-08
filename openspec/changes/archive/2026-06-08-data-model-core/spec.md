# Delta Spec: data-model-core

## Purpose

This delta spec documents the database schema requirements for the SED data model. It supplements the domain specs (B1–B5) with database-level validation, constraints, and non-functional requirements. All requirements here are enforced at the PostgreSQL level via Ent ORM schema definitions.

**Decisiones reflejadas:** #1, #2, #3, #4, #5, #6, #7, #8 (all).

## Scope

Applies to `api/internal/schema/` Ent files and `api/migrations/` SQL files.

## Requirements

### Requirement: Organization and org hierarchy

#### Scenario: Organization can have multiple org nodes

- GIVEN an `Organization` with id `org-1`
- WHEN creating multiple `OrgNode` records with `organization_id = org-1`
- THEN all nodes belong to that organization
- AND nodes can have a `parent_id` forming a tree structure

#### Scenario: Employee belongs to one org node

- GIVEN an `Employee` record
- WHEN `org_node_id` is set
- THEN the employee is linked to exactly one `OrgNode`
- AND changing `org_node_id` moves the employee in the org structure

### Requirement: Employee hierarchy (decision #8)

#### Scenario: Manager can view employee definitions

- GIVEN an `Employee` with `manager_id` pointing to another employee
- WHEN querying the reporting line
- THEN the `manager_id` FK correctly links to the manager's `Employee` record
- AND no orphaned `manager_id` values exist (FK constraint)

### Requirement: Cycle and phase transitions

#### Scenario: Only one active cycle per year per org

- GIVEN `organization_id = org-1` and `year = 2025`
- WHEN inserting a `Cycle`
- THEN a unique constraint on `(organization_id, year)` prevents duplicates
- AND the system can have exactly one active cycle per year

#### Scenario: Phase transitions are linear

- GIVEN a `PhaseTransition` from `asignacion` to `avance`
- WHEN the transition record exists
- THEN the unique constraint on `(from_phase, to_phase)` prevents duplicate transition definitions
- AND the DB enforces the DAG nature (no cycles) via application logic

### Requirement: Competency framework integrity (decision #2)

#### Scenario: Deleting a competency cascades to its scale criteria and acceptance levels

- GIVEN a `Competency` with related `ScaleCriterion` and `CompetencyAcceptanceLevel` records
- WHEN the `Competency` row is deleted
- THEN `ScaleCriterion` rows with that `competency_id` are cascade-deleted
- AND `CompetencyAcceptanceLevel` rows with that `competency_id` are cascade-deleted

#### Scenario: Level definitions are static

- GIVEN there are exactly 5 `LevelDefinition` records (levels 1–5)
- WHEN no new levels are inserted
- THEN the scale remains fixed at 1–5

### Requirement: Double weighting validation (decision #1)

#### Scenario: Category weight is between 0 and 100

- GIVEN a `GoalCategory` with `weight = 50`
- WHEN the record is inserted
- THEN a check constraint `weight >= 0 AND weight <= 100` is enforced
- AND a weight of 150 is rejected by the database

#### Scenario: Goal weight is between 0 and 100

- GIVEN a `Goal` with `weight = 30`
- WHEN the record is inserted
- THEN a check constraint `weight >= 0 AND weight <= 100` is enforced

### Requirement: Target value must be positive (decision #6)

#### Scenario: Goal target value must be greater than 0

- GIVEN a `Goal` with `target_value = 100`
- WHEN the record is inserted
- THEN a check constraint `target_value > 0` is enforced
- AND `target_value = 0` or negative values are rejected

### Requirement: KPIs are reusable (decision #6)

#### Scenario: Same KPI linked to multiple goals

- GIVEN a `KPI` record and two `Goal` records
- WHEN creating two `GoalKpiLink` records linking the KPI to each goal
- THEN both links are persisted
- AND deleting one goal does NOT delete the KPI (cascade is on GoalKpiLink → Goal only)

### Requirement: 9×9 matrix quadrant calculation

#### Scenario: Quadrant is computed on write

- GIVEN a `NineBoxEntry` with `performance_score = 8` and `potential_score = 9`
- WHEN the record is inserted
- THEN `quadrant` is set to `9` (derived from performance=7-9 AND potential=7-9)
- AND changing `performance_score` to `4` recalculates `quadrant` to `6`

### Requirement: Audit fields

#### Scenario: All evaluation-domain tables have audit fields

- GIVEN any of: `Goal`, `GoalCategory`, `Evaluation`, `EvaluationCompetency`, `EvaluationGoal`, `NineBoxEntry`
- WHEN a record is inserted
- THEN `created_at` and `updated_at` are set automatically (default: now)
- AND `created_by` and `updated_by` are set from the application context (not null for mutations)

### Requirement: No orphan employees

- GIVEN an `Employee` with `org_node_id` set to an `OrgNode`
- WHEN the `OrgNode` is deleted
- THEN the cascade delete on `OrgNode.employees` removes the `Employee` record

### Requirement: Goal categories belong to one employee (decision #8)

- GIVEN a `GoalCategory` with `employee_id = emp-1`
- WHEN the `Employee` with id `emp-1` is deleted
- THEN the cascade delete on `Employee.categories` removes the `GoalCategory` and its child `Goal` records

## Non-functional Requirements

### Indices

All indices listed in `design.md` MUST be created. No listing query should require a full table scan for tables with > 10,000 rows.

### Constraints

| Constraint | Entity | Rule |
|-----------|--------|------|
| Unique | `organizations.slug` | One slug per organization |
| Unique | `employees.email` | One email per employee |
| Unique | `employees.employee_number` | One employee number per organization |
| Unique | `cycles(organization_id, year)` | One cycle per org per year |
| Unique | `goal_assignments(employee_id, cycle_id)` | One assignment per employee per cycle |
| Unique | `evaluations(employee_id, cycle_id)` | One evaluation per employee per cycle |
| Unique | `nine_box_entries(matrix_id, evaluatee_id)` | One entry per evaluatee per matrix |
| Unique | `competency_acceptance_levels(competency_id, profile_id)` | One acceptance level per pair |
| Unique | `goal_kpi_links(goal_id, kpi_id)` | No duplicate links |
| Unique | `goal_categories(employee_id, name)` | No duplicate category names per employee |
| Check | `goal_categories.weight` | 0 ≤ weight ≤ 100 |
| Check | `goals.weight` | 0 ≤ weight ≤ 100 |
| Check | `goals.target_value` | > 0 |
| Check | `nine_box_entries.performance_score` | 1–9 |
| Check | `nine_box_entries.potential_score` | 1–9 |
| Check | `nine_box_entries.quadrant` | 1–9 |
| Check | `scale_criteria.level` | 1–5 |
| Check | `competency_acceptance_levels.level` | 1–5 |
| Check | `level_definitions.level` | 1–5 |

### Naming Conventions

- **Tables**: snake_case, plural (`employees`, `goal_categories`, `nine_box_matrices`)
- **Columns**: snake_case (`employee_id`, `current_phase`, `created_at`)
- **Indexes**: `idx_<table>_<columns>` (e.g., `idx_employee_manager_active`)
- **Constraints**: `chk_<table>_<rule>` (e.g., `chk_goals_weight`)
- **Enums**: snake_case (`phase_type`, `goal_unit`)

### Migration Requirements

- All migrations are **versioned** and stored in `api/migrations/`
- Migrations are **deterministic** — running `up` twice does not error
- Migrations support **rollback** via `down` migrations
- No migration modifies a previously applied migration
- Enum types are created before tables that reference them