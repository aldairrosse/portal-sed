# Delta: goal_progress_logs (database)

## Propósito

Registrar cada cambio del `current_value` de una meta en la tabla `goal_progress_logs`. Write-only desde la aplicación: solo se inserta, no se consulta via API. Sirve como auditoría para futuras features (historial de progreso, reportes). El INSERT ocurre exclusivamente dentro de la transacción de `UpdateGoalCurrentValue` y solo cuando el valor realmente cambia (`current_value IS DISTINCT FROM $newValue`).

## Esquema

```sql
CREATE TABLE IF NOT EXISTS goal_progress_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id     UUID        NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    value       FLOAT       NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID        NULL REFERENCES employees(id)
);

CREATE INDEX IF NOT EXISTS idx_goal_progress_logs_goal_time
    ON goal_progress_logs (goal_id, recorded_at DESC);
```

### Columnas

| Columna | Tipo | Constraints | Notas |
|---------|------|-------------|-------|
| `id` | UUID | PK, DEFAULT gen_random_uuid() | Identificador único del registro de progreso |
| `goal_id` | UUID | NOT NULL, FK → goals(id) ON DELETE CASCADE | Meta a la que pertenece este registro. Si se elimina la meta, se eliminan sus logs. |
| `value` | FLOAT | NOT NULL | Valor registrado de `current_value` en este momento |
| `recorded_at` | TIMESTAMPTZ | NOT NULL, DEFAULT now() | Momento exacto del registro |
| `created_by` | UUID | NULL, FK → employees(id) | Empleado que realizó el cambio. NULL si es cambio de sistema (ej. sync de KPI). Sin ON DELETE: si se elimina el empleado, el registro de log sobrevive. |

### Índice

`idx_goal_progress_logs_goal_time` en `(goal_id, recorded_at DESC)`: optimiza la consulta futura "historial de progreso de una meta ordenado por fecha descendente".

## Requirements

### Requirement: Registro condicional de cambio de progreso

El sistema DEBE insertar una fila en `goal_progress_logs` SOLO cuando `current_value` en `goals` realmente cambia. Si el nuevo valor es igual al existente, NO debe insertarse registro de log.

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

El UPDATE en `goals` y el INSERT en `goal_progress_logs` DEBEN ocurrir en la misma transacción. Si cualquiera falla, ambos deben hacer rollback.

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

El índice compuesto `(goal_id, recorded_at DESC)` DEBE existir para que consultas futuras de historial ordenado por fecha sean eficientes.

#### Scenario: Plan de consulta usa índice

- GIVEN tabla `goal_progress_logs` con 100K registros
- WHEN se ejecuta `SELECT * FROM goal_progress_logs WHERE goal_id = $1 ORDER BY recorded_at DESC`
- THEN el plan de ejecución usa `idx_goal_progress_logs_goal_time`
- AND la consulta se resuelve en <10ms

## Non-goals

- No se expone endpoint REST para consultar logs
- No hay UI para visualizar el historial
- No se implementa Ent schema (la tabla es write-only desde app code)
- No se implementa soft-delete ni retención configurable
- No se particiona por ciclo u organización
