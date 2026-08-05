# Tasks: goal-progress-log-focus-fix

Estrategia: **2 PRs independientes** (backend no comparte archivos con frontend). PR1 = frontend (fix de foco + guardado optimista), PR2 = backend (goal_progress_log). PR3 = verificación transversal.

---

## PR1 — Frontend: fix de foco + guardado optimista en background

### T1.1 Corregir GoalRow.svelte (input de avance)

- [ ] Cambiar `value={progressValue}` + `oninput={handleProgressInput}` por `bind:value={progressValue}` + `onchange={handleProgressBlur}` (aplica a los dos bloques: `medio-anio` línea ~187 y `fin-anio` línea ~173)
- [ ] Eliminar el `$effect` incondicional de resync (líneas 50-52): `progressValue` se siembra al montar
- [ ] Agregar `handleProgressFocus()` (set `progressFocused = true`) y `handleProgressBlur()` (persiste solo si `progressValue !== goal.progress`)
- [ ] Resync externo opcional: `$effect` que actualiza `progressValue` solo cuando `!progressFocused`
- [ ] Eliminar `handleProgressInput` (o dejarlo solo actualizando estado local, sin `onUpdateProgress`)
- **AC:** Escribir no pierde foco ni re-render; persiste una vez al blur; `pnpm run check` pasa

### T1.2 Actualizar goalsStore.updateGoalProgress (optimista, background)

- [ ] Convertir `updateGoalProgress` en async: aplicar valor optimista, persistir en background, sin `reload()`
- [ ] Extraer frontera `persistProgress(goalId, progress)` — hoy no-op en DEV; será `client.PATCH` en C8
- [ ] En error: rollback local del valor + exponer `lastProgressError`
- [ ] Exponer `getLastProgressError()` y `clearLastProgressError()`
- **AC:** Guardar no recarga; error restaura valor; API pública compatible con el consumidor actual

### T1.3 Conectar error en la página asignacion

- [ ] En `+page.svelte`, mostrar `lastProgressError` en la alerta `errorMsg` existente tras un fallo de persistencia
- **AC:** Fallo de PATCH muestra alerta sin recargar

### T1.4 Test de componente GoalRow

- [ ] Test Vitest + Testing Library: escribir N teclas no pierde foco; blur dispara `onUpdateProgress` una vez; valor sin cambio no dispara persistencia
- **AC:** Tests pasan con Vitest

---

## PR2 — Backend: goal_progress_log

### T2.1 Ent schema GoalProgressLog

- [ ] Crear `api/internal/schema/goalprogresslog.go` (id, goal_id, employee_id, previous_value, new_value + AuditMixin + índices)
- [ ] Agregar edge `edge.To("progress_logs", GoalProgressLog.Type)` en `Goal.Edges()`
- [ ] Regenerar código Ent (`go generate ./...`)
- **AC:** `go build ./...` compila; entidad y edge generados

### T2.2 Migración manual en migrate/schema.go

- [ ] Crear `GoalProgressLogsTable` siguiendo el patrón de `GoalsTable` (columnas, índices goal_id/employee_id, relación BelongsTo goals)
- [ ] Registrar la tabla en la colección de migración existente
- **AC:** La migración crea `goal_progress_logs` con índices y FK; se puede aplicar contra PostgreSQL local

### T2.3 Repo: UpdateGoalCurrentValueWithLog (transacción)

- [ ] Implementar método atómico en `goal_repo.go` (BeginTx → UPDATE current_value/version → INSERT log → Commit; RowsAffected==0 → ErrGoalNotFound; re-fetch del goal)
- **AC:** UPDATE e INSERT atómicos; tests unitarios del repo con transacción

### T2.4 Servicio: conectar previous_value y log

- [ ] En `progress_service.go`, usar `existing.CurrentValue` como `previousValue`
- [ ] Si `equalFloat(current, previous)` → retornar `existing` sin escribir log ni version++
- [ ] Delegar en `UpdateGoalCurrentValueWithLog(ctx, goalID, empID, current, previous)`
- **AC:** Misma firma pública (ProgressServicer sin cambios); `go test ./api/...` pasa

### T2.5 OpenAPI: documentar auditoría

- [ ] Actualizar `summary` de `updateGoalProgress` en `goals-api.yaml` indicando la persistencia del log
- **AC:** `openspec validate --all` pasa; contrato request/response sin cambios

### T2.6 Tests backend

- [ ] Handler/service test: PATCH con valor distinto → `current_value` actualizado + fila de log con previous/new/employee_id
- [ ] Test: mismo valor → sin fila de log, sin incremento de version
- [ ] Test: fase no-avance → 403 y sin fila de log
- **AC:** `go test ./...` pasa con los nuevos casos

---

## PR3 — Verificación transversal

### T3.1 Verificación manual del flujo

- [ ] Levantar BD + API (docker-compose), aplicar migración
- [ ] En web con `VITE_USE_API=true` (o wiring C8), editar avance: sin pérdida de foco, sin loader, guardado en blur
- [ ] Consultar `goal_progress_logs` tras guardar: fila correcta (previous/new/employee)
- **AC:** Flujo end-to-end verificado; log poblado

### T3.2 Validación final

- [ ] `pnpm run check` en `web/` y `go test ./...` en `api/` sin errores
- [ ] `openspec validate --all` sin errores
- **AC:** Todos los checks verdes

---

## Resumen de PRs

| PR | Archivos | Tareas | Depende de |
|---|---|---|---|
| PR1 — Frontend | GoalRow.svelte, goalsStore.svelte.ts, +page.svelte, test | T1.1–T1.4 | Ninguna |
| PR2 — Backend | schema/goalprogresslog.go, migrate/schema.go, goal_repo.go, progress_service.go, goals-api.yaml, tests | T2.1–T2.6 | Ninguna |
| PR3 — Verificación | — | T3.1–T3.2 | PR1 + PR2 |

PR1 y PR2 pueden implementarse en paralelo (no comparten archivos). PR3 al final.
