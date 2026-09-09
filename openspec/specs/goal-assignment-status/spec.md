# goal-assignment-status Specification

## Purpose
Persist goal assignment lifecycle status and submission timestamp with validated transitions.

## Requirements

### Requirement: GoalAssignment lifecycle SHALL persist status and submission timestamp

The `goal_assignments` table SHALL store `status` (`borrador`|`enviada`, default `borrador`) and `submitted_at` (nullable TIMESTAMPTZ); transition to `enviada` SHALL require passing `ValidateDoubleWeighting` and SHALL stamp `submitted_at = now()` (see REQ-MEM-001/002 below).

#### Scenario: Creation starts as borrador

- WHEN an assignment is created before validation or submission
- THEN `status` is `borrador` and `submitted_at` is null

### Requirement: REQ-MEM-001 SHALL persist status y submitted_at en GoalAssignment

The system SHALL persist `status` and `submitted_at` for GoalAssignment.

La tabla `goal_assignments` y el esquema Ent `GoalAssignment` DEBEN almacenar los campos:
- `status`: enum (`borrador`, `enviada`), con valor por defecto `'borrador'`.
- `submitted_at`: timestamp con zona horaria (`TIMESTAMPTZ`), nulo inicialmente hasta el envío formal.

El endpoint `GET /api/v1/employees/{empId}/assignments` DEBE retornar `status` y `submitted_at` en `AssignmentResponse`.

#### Scenario: Creación inicial en borrador
- GIVEN un empleado que inicia la formulación de metas para el ciclo activo
- WHEN se crea la asignación antes de validar o enviar
- THEN `status` es `'borrador'`
- AND `submitted_at` es `null`

---

### Requirement: REQ-MEM-002 SHALL validar Pesos Double 100% como compuerta para status 'enviada'

The system SHALL gate transition to `enviada` on `ValidateDoubleWeighting` passing at 100%.

Al solicitar el guardado/envío formal de la asignación (`POST /api/v1/employees/{empId}/assignments`):
1. El backend DEBE ejecutar la validación de pesos Double 100% (`ValidateDoubleWeighting`):
   - La suma de pesos de las categorías DEBE ser 100% ($\pm \epsilon$).
   - La suma de pesos de las metas cuantitativas personales dentro de cada categoría DEBE ser 100% ($\pm \epsilon$).
2. Si la validación es exitosa:
   - La asignación transiciona a `status = 'enviada'`.
   - Se registra `submitted_at = now()`.
   - Se retorna `HTTP 201/200` con `AssignmentResponse` actualizado.
3. Si la validación no es exitosa (suma $<100\%$ o $>100\%$):
   - El estado NO cambia a `'enviada'`.
   - Permanece en `'borrador'`.
   - Se rechaza el envío formal con error `400` (`WeightInvalid`).

#### Scenario: Envío exitoso con ponderación al 100%
- GIVEN categorías que suman 100% y metas en cada categoría que suman 100%
- WHEN el colaborador o jefe envía la asignación
- THEN la validación retorna `Valid: true`
- AND `GoalAssignment.status` cambia a `'enviada'`
- AND `GoalAssignment.submitted_at` se actualiza con la fecha y hora actual

#### Scenario: Bloqueo de envío con ponderación incompleta o excedida
- GIVEN metas personales cuya suma de pesos es 80% o 110%
- WHEN se intenta enviar la asignación
- THEN el backend responde `400 Bad Request` indicando déficit o exceso de peso
- AND `GoalAssignment.status` permanece en `'borrador'`
- AND `GoalAssignment.submitted_at` permanece sin cambios (`null`)

---

### Requirement: REQ-MEM-003 SHALL eliminar creación automática de asignaciones vacías en el frontend

The system SHALL NOT auto-create empty assignments in `_doLoad()`.

En `goalsStore.svelte.ts`, la función `_doLoad()` NO DEBE realizar `POST /employees/{empId}/assignments` de forma automática ni silenciosa cuando no existe una asignación previa para un subordinado. La asignación solo debe crearse por acción deliberada del usuario (al guardar o formular metas).

#### Scenario: Consulta de subordinado sin asignación previa
- GIVEN un jefe que consulta las metas de un colaborador nuevo sin asignación en el ciclo
- WHEN el store ejecuta `loadForEmployee(empId)`
- THEN NO se dispara ninguna petición POST silenciosa en segundo plano
- AND el estado local refleja que el colaborador aún no cuenta con asignación (`borrador` o stub sin persistir)

---
