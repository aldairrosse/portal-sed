# Propuesta: C9 — goal-progress-log-focus-fix

## 1. Intent

Corregir el bug de UX al editar el avance de una meta en fase `medio-anio`/`fin-anio`: al escribir un valor en el input de progreso, el componente pierde el foco y fuerza un re-render que muestra loader. El fix mueve el guardado a un patrón **optimista en background** (sin reload ni refetch) y, de paso, agrega la tabla **`goal_progress_log`** para auditar cada cambio de avance (quién, cuándo, valor anterior → valor nuevo).

**Especificaciones base:** `api/openapi/goals-api.yaml` (`/goals/{goalId}/progress` → `updateGoalProgress`), `api/internal/schema/goal.go`, `api/internal/repository/goal/goal_repo.go`, `api/internal/service/goal/progress_service.go`, `web/src/lib/components/goals/GoalRow.svelte`, `web/src/lib/stores/goalsStore.svelte.ts`.

**Decisiones de arquitectura referenciadas:**
- C4 (goals-api): endpoint `PATCH /goals/{goalId}/progress` existe y valida fase (`avance`) + ownership.
- C8 (wire-api-replace-mocks, en curso): el frontend pasará de fixtures a API real vía `openapi-fetch`. Este change no depende de C8: el fix de foco es puramente de UI y el log es puramente de backend.
- El avance es hoy un campo único `current_value` en `goals` (sobrescrito en cada PATCH) — sin historial ni auditoría.

## 2. Scope

### In Scope

**Frontend (fix de foco + guardado optimista):**

- **`GoalRow.svelte`:**
  - El input de progreso pasa a guardar en `change` (blur) y **no** en cada `oninput` — una escritura por sesión de edición, sin debounce (YAGNI).
  - El estado local `progressValue` es la autoridad mientras el input tiene foco. Se elimina (o se blinda con `document.activeElement` check) el `$effect` de re-sincronización que sobreescribe el valor a mitad de escritura y provoca el re-render.
  - Guardado optimista: al blur se llama `onUpdateProgress(goal.id, val)`; el nodo DOM del input no se reemplaza ni se le reescribe el valor durante la edición.
- **`goalsStore.svelte.ts` — `updateGoalProgress`:** actualiza el estado local de forma optimista e inicia la persistencia en background. **No** llama `reload()`, `invalidateAll()` ni refetch tras el guardado. En error: restaura el valor anterior + expone el error.
- El patrón queda listo para que el wiring a la API real (C8) use exactamente este flujo sin recargar.

**Backend (goal_progress_log):**

- **Ent schema nuevo `api/internal/schema/goalprogresslog.go`:** entidad `GoalProgressLog` con campos `id` (uuid), `goal_id` (uuid, FK → goals), `employee_id` (uuid, quien registró), `previous_value` (float), `new_value` (float), `created_at` (timestamp). Edge 1:N `Goal → progress_logs`.
- **Migración `api/internal/migrate/schema.go`:** nueva entrada `GoalProgressLogsTable` (patrón `schema.Table` existente) con columnas + índices en `goal_id` y `employee_id`.
- **Repo `api/internal/repository/goal/goal_repo.go`:** método atómico que actualiza `current_value` (UPDATE existente, `version + 1`) **e inserta la fila de log en la misma transacción** (el repo ya tiene acceso a `db` raw y al cliente Ent).
- **Servicio `api/internal/service/goal/progress_service.go`:** `UpdateGoalProgress` lee el `current_value` actual (previous_value), valida fase + ownership (sin cambios), y delega en el método atómico del repo.
- **OpenAPI `goals-api.yaml`:** documentar en `updateGoalProgress` que cada PATCH persiste una fila de auditoría (sin cambios de contrato en request/response; `UpdateProgressRequest` queda igual con `current_value`).

### Out of Scope

- **Endpoint de lectura del historial** (`GET /goals/{goalId}/progress-log`) — se difiere; el log solo escribe por ahora (YAGNI hasta que una pantalla lo requiera).
- **Campo `comment` en `UpdateProgressRequest`** — no solicitado; se agrega cuando el producto lo pida.
- **Debounce/coalescing de escrituras** — el guardado en blur ya evita escrituras por tecla; si A4 (mid-year-progress-ui) requiere autosave con typing continuo, se evalúa debounce en change aparte.
- **UI de historial de avances** — pertenece a la pantalla A4, no a este change.
- **Cambios al RBAC** — se reutiliza `goal:progress` existente.
- **Migración de datos retroactiva** — el log solo registra cambios a partir de su despliegue.

## 3. Approach

### Frontend: input con guardado en blur + estado local autoritativo

```
GoalRow input (bind:value local) → onchange (blur) → onUpdateProgress(goalId, val)
                                                        ↓
                               goalsStore.updateGoalProgress (optimista, sin await)
                                                        ↓
                                    PATCH /goals/{goalId}/progress  (background)
                                                        ↓
                                    error → rollback local + alerta  (sin refetch)
```

- El `$effect` de resync se elimina: `progressValue` se siembra desde `goal.progress` al montar y se actualiza únicamente al confirmar (blur). Cualquier cambio externo (otra pestaña/empleado) se sincroniza solo si el input no está enfocado.
- Como el guardado ya no se dispara por tecla, el store no recrea el array en cada keystroke → no hay churn de props → el DOM no se toca → **el foco se conserva** y no aparece loader.

### Backend: log atómico con el UPDATE

- Ent schema nuevo + migración manual en `migrate/schema.go` (patrón `GoalsTable` existente).
- Repo: `UpdateGoalCurrentValueWithLog(ctx, goalID, employeeID, currentValue, previousValue) (*GoalRow, error)`:
  - `BeginTx` → `UPDATE goals SET current_value=$1, updated_at=$2, version=version+1 WHERE id=$3` (RowsAffected==0 → `ErrGoalNotFound`) → `INSERT INTO goal_progress_log (goal_id, employee_id, previous_value, new_value, created_at)` → `Commit`.
  - Re-fetch del goal actualizado (igual que hoy) para el response.
- Servicio: `UpdateGoalProgress` obtiene `existing.CurrentValue` como `previousValue` antes del update (ya hace `GetGoal` para ownership) y pasa ambos valores al repo.

## 4. Dependencies

- **C1 (data-model-core):** esquema Ent y migraciones existentes — base para la nueva tabla.
- **C4 (goals-api):** endpoint `PATCH /goals/{goalId}/progress` + `ProgressServicer` — el servicio y repo existen y se extienden.
- **C8 (wire-api-replace-mocks, en curso):** el patrón de guardado background queda preparado para el wiring API; este change no lo bloquea ni depende de él.

## 5. Risk & Mitigation

### Riesgo: Perder escrituras si el usuario no hace blur (navega/recarga antes de salir del input)

- **Mitigación:** se mantiene `oninput` actualizando solo el estado local (sin persistir); al blur se persiste. Cobertura adicional (beforeunload guard) se evalúa en A4 si se confirma el caso.

### Riesgo: Regresión en GoalRow (otras fases/columnas)

- **Mitigación:** el cambio se limita al bloque del input de progreso; el modo editor/reader y demás columnas no se tocan. `svelte-check` + tests de componente existentes.

### Riesgo: Log y UPDATE fuera de sincronía (crash entre ambos)

- **Mitigación:** ambos escritos en una sola transacción SQL. Si falla, no se actualiza ni el valor ni el log (rollback).

### Riesgo: Ruido de log por guardados repetidos del mismo valor

- **Mitigación:** el servicio solo persiste cuando `new_value != previous_value` (con tolerancia EPSILON); si no hay cambio, no se escribe log ni se toca `version`.

## 6. Success Criteria

- [ ] Escribir en el input de avance **no** pierde el foco y no muestra loader en ninguna tecla.
- [ ] El avance se persiste al salir del input (blur) y en background, sin `reload()`/refetch/`invalidateAll()`.
- [ ] Si el PATCH falla, el input restaura el valor anterior y se muestra un error; no se recarga la página.
- [ ] Cada PATCH exitoso con valor distinto inserta una fila en `goal_progress_log` con `goal_id`, `employee_id`, `previous_value`, `new_value`, `created_at`.
- [ ] El log y el UPDATE de `current_value` ocurren en la misma transacción (sin estados intermedios).
- [ ] Guardar el mismo valor no crea fila de log ni incrementa `version`.
- [ ] La migración crea la tabla `goal_progress_log` con índices en `goal_id` y `employee_id`.
- [ ] `go test ./...` en `api/` y `pnpm run check` en `web/` pasan sin errores.
- [ ] `openspec validate --all` sin errores.
