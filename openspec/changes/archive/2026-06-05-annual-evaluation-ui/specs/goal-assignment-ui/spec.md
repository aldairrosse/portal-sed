# Delta for Goal Assignment UI

## ADDED Requirements

### Requirement: Goal permissions for fin-anio phase

`getGoalPermissions()` SHALL return phase-aware permissions. For `fin-anio`: owners get `canClose: true` (all other permissions `false`); non-owners get all `false`.

#### Scenario: Owner in fin-anio can only close

- GIVEN `cyclePhase === 'fin-anio'` and employee is goal owner
- WHEN `getGoalPermissions()` is called
- THEN returns `{ canEditProgress: false, canComment: false, canEditWeight: false, canDelete: false, canClose: true }`

#### Scenario: Non-owner in fin-anio has no write access

- GIVEN `cyclePhase === 'fin-anio'` and user is RH viewing another employee
- WHEN `getGoalPermissions()` is called
- THEN all permissions are `false`

#### Scenario: RH can still view all goals

- GIVEN `cyclePhase === 'fin-anio'` and RH viewing any employee
- WHEN goals are rendered
- THEN goals are visible in read-only mode (no edit/delete actions)

### Requirement: Goal closure self-assessment field in fin-anio

When `cyclePhase === 'fin-anio'`, `GoalRow` SHALL render a `selfAssessment` textarea for the goal owner. This field is persisted in `GoalClosure` (via `evaluationStore`), separate from mid-year `GoalComment`.

#### Scenario: Employee adds self-assessment

- GIVEN employee in `/mi-evaluacion` (`fin-anio`), goal X rendered
- WHEN employee enters self-assessment text and saves
- THEN `GoalClosure.selfAssessment` is set in `evaluationStore`
- AND the textarea shows saved content on subsequent renders

#### Scenario: Read-only goals for non-owners in fin-anio

- GIVEN `cyclePhase === 'fin-anio'` and RH viewing employee's goals
- WHEN `GoalRow` renders
- THEN name, weight, targetValue, KPIs are read-only
- AND `selfAssessment` shows content but is not editable (RH uses separate `rhAssessment`)

#### Scenario: Final progress input replaces mid-year progress

- GIVEN `cyclePhase === 'fin-anio'` and goal with existing mid-year progress
- WHEN `GoalRow` renders
- THEN the progress field is relabeled "Avance final" and is editable only for owner
- AND mid-year progress value is pre-filled as starting point

### Requirement: Goal freeze on closure

After a goal is closed (self-assessment submitted), it SHALL become fully read-only for the employee. Only RH can modify `rhAssessment` after closure.

#### Scenario: Closed goal is frozen for employee

- GIVEN employee has submitted self-assessment for goal X
- WHEN reloading `/mi-evaluacion`
- THEN goal X shows all fields as read-only with `badge-ghost` "Cerrada"
- AND self-assessment text is visible but not editable

## MODIFIED Requirements

### Requirement: Bloqueo de eliminación en fase avance (extended to cierre)

The system SHALL prohibit deletion of goals and categories when `cyclePhase !== 'inicio-anio'`. The deletion guard now covers `medio-anio` **and** `fin-anio`.
(Previously: guard covered only `avance` [medio-anio] phase)

#### Scenario: Sin botón eliminar meta en cualquier fase excepto inicio-anio

- GIVEN `cyclePhase !== 'inicio-anio'`
- WHEN se renderiza cualquier `GoalRow`
- THEN NO existe botón "Eliminar" en las acciones

#### Scenario: Sin botón eliminar categoría en cualquier fase excepto inicio-anio

- GIVEN `cyclePhase !== 'inicio-anio'`
- WHEN se renderiza cualquier `CategoryCard`
- THEN NO existe botón "Eliminar" en el header

## REMOVED Requirements

None.
