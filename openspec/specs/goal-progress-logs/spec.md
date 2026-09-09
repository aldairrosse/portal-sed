# goal-progress-logs Specification

## Purpose
Log goal progress history only when the current value actually changes.

## Requirements

### Requirement: Registro condicional de cambio de progreso

El sistema DEBE (SHALL) insertar una fila en `goal_progress_logs` SOLO cuando `current_value` en `goals` realmente cambia. Si el nuevo valor es igual al existente, NO debe insertarse registro de log.

#### Scenario: Cambio real de progreso

- GIVEN meta con `current_value = 45`
- WHEN se llama `UpdateGoalCurrentValue(goalID, 60)`
- THEN `goals.current_value` se actualiza a 60
- AND se inserta una fila en `goal_progress_logs` con `value = 60`, `goal_id = goalID`, `recorded_at = now()`

#### Scenario: Mismo valor — sin registro

- GIVEN meta con `current_value = 60`
- WHEN se llama `UpdateGoalCurrentValue(goalID, 60)`
- THEN `goals.current_value` permanece 60 (sin UPDATE)
- AND NO se inserta fila en `goal_progress_logs`
- AND la transacción hace COMMIT sin cambios

#### Scenario: Meta eliminada — logs en cascada

- GIVEN meta con 5 registros en `goal_progress_logs`
- WHEN la meta es eliminada via `DELETE FROM goals WHERE id = ...`
- THEN los 5 registros en `goal_progress_logs` son eliminados automáticamente (ON DELETE CASCADE)

#### Scenario: created_by opcional

- GIVEN actualización de progreso disparada por el sistema (sin usuario)
- WHEN se inserta en `goal_progress_logs` con `created_by = NULL`
- THEN el registro es válido
- AND la FK a employees no se viola (es nullable)

### Requirement: Atomicidad de UPDATE + INSERT

El UPDATE en `goals` y el INSERT en `goal_progress_logs` DEBEN (SHALL) ocurrir en la misma transacción. Si cualquiera falla, ambos deben hacer rollback.

#### Scenario: INSERT falla → rollback del UPDATE

- GIVEN un error en el INSERT a `goal_progress_logs` (ej. FK violation)
- WHEN ocurre dentro de la transacción
- THEN el UPDATE en `goals` hace rollback
- AND `goals.current_value` mantiene su valor original

#### Scenario: UPDATE sin cambio → sin INSERT

- GIVEN `current_value IS NOT DISTINCT FROM $newValue`
- WHEN el UPDATE afecta 0 filas
- THEN NO se ejecuta INSERT
- AND la transacción hace COMMIT (no-op exitoso)

### Requirement: Índice para consultas futuras

El índice compuesto `(goal_id, recorded_at DESC)` DEBE (SHALL) existir para que consultas futuras de historial ordenado por fecha sean eficientes.

#### Scenario: Plan de consulta usa índice

- GIVEN tabla `goal_progress_logs` con 100K registros
- WHEN se ejecuta `SELECT * FROM goal_progress_logs WHERE goal_id = $1 ORDER BY recorded_at DESC`
- THEN el plan de ejecución usa `idx_goal_progress_logs_goal_time`
- AND la consulta se resuelve en <10ms
