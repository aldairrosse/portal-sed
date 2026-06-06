# evaluados-progress-column Specification

## Purpose

Extiende la tabla "Mis evaluados" con columna de progreso agregado de metas y énfasis visual para evaluaciones incompletas. Frontend-only, fixture-based, reutiliza `ProgressIndicator` existente y getters de `goalsStore` / `evaluationStore` sin modificarlos.

## Requirements

### Requirement: Columna de progreso global

El sistema SHALL mostrar una columna "Progreso global" entre "Área" y "Estado" en `EmployeeEvaluationTable`. El valor SHALL calcularse como `(Σ progress) / (Σ targetValue) × 100` usando `getGoals()` y `getAssignments()`. El resultado SHALL renderizarse con `ProgressIndicator` (colores: <40% rojo, <80% amarillo, ≥80% verde).

| Caso | Comportamiento |
|------|---------------|
| Metas con avance | Barra coloreada + badge numérico (`ProgressIndicator`) |
| Sin metas (`goalIds` vacío) | Muestra "—" (guion), sin barra |
| `progress` indefinido en todas las metas | Barra 0%, rojo, badge "0%" |
| `Σ targetValue = 0` | Muestra "—" como fallback |

#### Scenario: Progreso con metas múltiples

- GIVEN empleado con Meta A (progress=32, targetValue=100) y Meta B (progress=45, targetValue=50)
- WHEN renderiza la tabla
- THEN columna muestra `ProgressIndicator` al 51% ((32+45)/(100+50))
- AND color amarillo por estar en rango 40–79%

#### Scenario: Empleado sin metas

- GIVEN empleado con `goalIds: []` en su `EmployeeAssignment`
- WHEN renderiza su fila
- THEN columna "Progreso global" muestra "—"

### Requirement: Visibilidad de evaluaciones incompletas

El sistema SHALL destacar filas con evaluación `pending` o `in-progress` y SHALL mostrar un chip resumen "X de Y completaron" sobre la tabla usando `getEvaluationStatus()`.

#### Scenario: Chip resumen

- GIVEN 5 empleados: 3 `completed`, 1 `in-progress`, 1 `pending`
- WHEN renderiza la tabla
- THEN chip "3 de 5 completaron" visible arriba de la tabla
- AND usa `badge-success` si ≥80%, `badge-warning` si <80%

#### Scenario: Fila pendiente con énfasis visual

- GIVEN empleado con estado `pending`
- WHEN renderiza su fila
- THEN fila tiene clase `bg-warning/10` (fondo sutil)
- AND mantiene hover `hover:bg-base-200`

#### Scenario: Fila completada sin énfasis

- GIVEN empleado con estado `completed`
- WHEN renderiza su fila
- THEN fila NO tiene fondo de énfasis
- AND mantiene solo el hover estándar

### Requirement: Inmutabilidad de stores

El sistema SHALL usar exclusivamente getters existentes: `getGoals()`, `getAssignments()`, `getEvaluationStatus()`. No SHALL crear mutaciones ni modificar fixtures.

#### Scenario: Cálculo sin side effects

- GIVEN fixtures cargados y componente montado
- WHEN se calcula progreso para todos los empleados
- THEN `getGoals()` y `getAssignments()` retornan los mismos datos
- AND `goals.json` y `assignments.json` no se modifican

## Non-goals

- Backend API ni persistencia
- Ordenamiento/filtrado por columna de progreso
- Exportación o acciones bulk
- Nuevas mutaciones de store
- Creación de componentes nuevos (se reutiliza `ProgressIndicator` y `EvaluationStatusBadge`)
