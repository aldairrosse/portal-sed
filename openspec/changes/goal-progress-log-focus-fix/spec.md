# Spec: goal-progress-log-focus-fix

## Propósito

Corregir la pérdida de foco y el re-render con loader al editar el avance de una meta (fases `medio-anio`/`fin-anio`), moviendo el guardado a un patrón optimista en background sin reload; y persistir una fila de auditoría `goal_progress_log` por cada cambio de avance, atómicamente con el UPDATE de `current_value`.

## Alcance

### Requerimientos funcionales

| ID | Requerimiento | Prioridad |
|---|---|---|
| F1 | Escribir en el input de avance no pierde el foco en ninguna tecla | Alta |
| F2 | Escribir en el input de avance no dispara re-render ni loader | Alta |
| F3 | El avance se persiste al salir del input (blur/`change`), no por keystroke | Alta |
| F4 | La persistencia es optimista y en background: sin `reload()`, refetch ni `invalidateAll()` tras guardar | Alta |
| F5 | Si la persistencia falla, el input restaura el valor anterior y se muestra un error; no se recarga la página | Alta |
| F6 | Cada PATCH de progreso con valor distinto inserta una fila en `goal_progress_log` | Alta |
| F7 | La fila de log registra `goal_id`, `employee_id` (sesión), `previous_value`, `new_value`, `created_at` | Alta |
| F8 | UPDATE de `current_value` e INSERT del log ocurren en la misma transacción SQL | Alta |
| F9 | Guardar el mismo valor no crea fila de log ni incrementa `version` | Media |
| F10 | La validación existente se conserva: fase `avance` y ownership de la meta | Alta |
| F11 | RBAC existente (`goal:progress`) se reutiliza; sin permisos nuevos | Alta |

### No alcance

- Endpoint de lectura del historial (`GET /goals/{goalId}/progress-log`)
- Campo `comment` en `UpdateProgressRequest`
- Debounce/coalescing de escrituras por keystroke
- UI de historial de avances (pertenece a A4)
- Migración retroactiva de datos
- Cambios en el contrato OpenAPI de request/response

## Escenarios

### E1: Escribir avance mantiene el foco

```
Given fase = medio-anio y canEditProgress = true
When el usuario escribe un valor en el input de avance (cada tecla)
Then el input conserva el foco
And no se muestra loader ni re-render del componente
And el estado local refleja el valor escrito
And NO se ha persistido aún (solo al blur)
```

### E2: Guardar al salir del input (blur) en background

```
Given el usuario escribió "75" en el input de avance
When el input pierde el foco (onchange)
Then onUpdateProgress(goalId, 75) se ejecuta una vez
And el store aplica el valor optimista (progress = 75, progressUpdatedAt = now)
And la persistencia corre en background sin bloquear la UI
And NO se llama reload() / refetch / invalidateAll()
```

### E3: Fallo de persistencia → rollback sin reload

```
Given la persistencia falla (red / 5xx / 403 / 404)
When el guardado en background retorna error
Then el input/estado restaura el valor anterior (rollback local)
And se muestra un error al usuario (alerta existente errorMsg)
And NO se recarga la página ni se pierde el foco
```

### E4: PATCH con valor distinto → fila de log atómica

```
Given una meta con current_value = 50 y un empleado autenticado
When PATCH /goals/{goalId}/progress con {"current_value": 75}
Then UPDATE goals SET current_value = 75, version = version + 1
And INSERT goal_progress_logs (goal_id, employee_id, previous_value = 50, new_value = 75, created_at)
And ambos escritos ocurren en la misma transacción (commit o rollback total)
And el response 200 contiene el goal actualizado
```

### E5: Mismo valor → sin log ni versión

```
Given una meta con current_value = 75
When PATCH /goals/{goalId}/progress con {"current_value": 75}
Then NO se inserta fila en goal_progress_logs
And NO se incrementa version
And el response 200 contiene el goal sin cambios
```

### E6: Fase o ownership inválidos (regresión)

```
Given fase ≠ avance (p.ej. asignacion)
When PATCH /goals/{goalId}/progress
Then 403 PhaseRestricted y NO se inserta log

Given la meta pertenece a otro empleado
When PATCH /goals/{goalId}/progress
Then 404 GoalNotFound y NO se inserta log
```

### E7: Tabla migrada e indexada

```
Given se aplica la migración
Then existe la tabla goal_progress_logs
And existen índices en goal_id y employee_id
And FK goal_id → goals(id)
```

## Contrato de datos

### Tabla `goal_progress_logs`

| Columna | Tipo | Notas |
|---|---|---|
| `id` | uuid PK | default uuid v4 |
| `goal_id` | uuid NOT NULL | FK → goals(id), index |
| `employee_id` | uuid NOT NULL | quien registró el avance (sesión), index |
| `previous_value` | float NOT NULL | valor antes del PATCH |
| `new_value` | float NOT NULL | valor después del PATCH |
| `created_at` | timestamptz NOT NULL | AuditMixin |
| `created_by` | uuid NULL | AuditMixin |

### Entidades Ent

- `GoalProgressLog` (nuevo schema) + edge `Goal → progress_logs` (1:N).
- `Goal` no cambia de campos; solo agrega el edge.
