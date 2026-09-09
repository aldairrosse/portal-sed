# Tasks: Fix Goal Progress Focus Loss + goal_progress_logs

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~120–170 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Backend — Migración

- [ ] 1.1 Crear `000027_goal_progress_logs.up.sql`
  — `CREATE TABLE goal_progress_logs` con columnas: `id UUID PK DEFAULT gen_random_uuid()`, `goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE`, `value FLOAT NOT NULL`, `recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()`, `created_by UUID NULL REFERENCES employees(id)`
  — Índice: `CREATE INDEX idx_goal_progress_logs_goal_time ON goal_progress_logs (goal_id, recorded_at DESC)`
  — AC: `goose up` crea tabla + índice sin errores

- [ ] 1.2 Crear `000027_goal_progress_logs.down.sql`
  — `DROP TABLE IF EXISTS goal_progress_logs`
  — AC: `goose down` elimina tabla sin errores

## Phase 2: Backend — Repo: UpdateGoalCurrentValue con transacción

- [ ] 2.1 Modificar `UpdateGoalCurrentValue` en `goal_repo.go`
  — Agregar parámetro `createdBy *uuid.UUID`
  — Envolver en `db.BeginTx`: SELECT `current_value`, UPDATE condicional (`WHERE current_value IS DISTINCT FROM $1`), INSERT en `goal_progress_logs` solo si `RowsAffected() > 0`, COMMIT
  — Defer `tx.Rollback()` para safety
  — Si SELECT no encuentra la meta → `ErrGoalNotFound`
  — AC: test unitario con sqlmock cubre: (a) mismo valor → sin UPDATE, sin INSERT, (b) valor distinto → UPDATE + INSERT, (c) goal not found → error

- [ ] 2.2 Actualizar firma en `GoalRepository` interface (`interfaces.go`)
  — `UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64, createdBy *uuid.UUID) (*GoalRow, error)`
  — AC: compila

## Phase 3: Backend — Service: pasar createdBy

- [ ] 3.1 Modificar `ProgressService.UpdateGoalProgress` en `progress_service.go`
  — Pasar `&empID` como `createdBy` a `UpdateGoalCurrentValue`
  — La variable `empID` ya está disponible en el scope del método
  — AC: compila; test de service actualizado

- [ ] 3.2 Actualizar mocks de `GoalRepository` en tests
  — `mockGoalRepoForValidation` en `validation_service_test.go`: actualizar firma
  — `mockGoalRepo` en `mocks_test.go` (handler): actualizar firma
  — AC: `go test ./...` pasa

## Phase 4: Frontend — Store: updateGoalProgress optimista + debounce

- [ ] 4.1 Reescribir `updateGoalProgress` en `goalsStore.svelte.ts`
  — Agregar `debounceTimers: Map<string, ReturnType<typeof setTimeout>>` en scope del módulo
  — Al entrar: cancelar timer existente para `goalId`
  — Guarda: si `goal.progress === progress` → resolver inmediatamente (no-op)
  — Snapshot `previousProgress = goal.progress` para rollback
  — Update optimista: mutar `storeState.data.goals[n].progress` y `progressUpdatedAt`
  — Devolver `new Promise` que arma timer de 500ms
  — Dentro del timer: PATCH + rollback en catch
  — Eliminar `await reload()` (ya no se invoca)
  — Usar mismo patrón que `addGoalComment` (líneas 472-492) para mutación local
  — AC: `pnpm run check` pasa; el store no dispara `reload()` desde `updateGoalProgress`

- [ ] 4.2 Verificar que `GoalRow.svelte` no necesita cambios
  — `GoalRow.svelte` sigue llamando `updateGoalProgress(goalId, value)` igual que antes
  — AC: diff no incluye `GoalRow.svelte`

## Phase 5: Validación

- [ ] 5.1 `go test ./...` pasa
  — Todos los tests existentes + nuevos tests de repo
  — AC: cero fallos

- [ ] 5.2 `pnpm run check` pasa
  — Svelte + TypeScript sin errores
  — AC: cero errores

- [ ] 5.3 `pnpm run lint` pasa
  — ESLint + Prettier sin warnings
  — AC: cero warnings

- [ ] 5.4 E2E manual: foco no se pierde
  — Abrir `/objetivos/asignacion` en fase medio-año
  — Editar progreso de una meta
  — Verificar que el foco permanece en el input después del PATCH
  — Verificar que `ProgressIndicator` se actualiza sin flicker
  — AC: comportamiento fluido, sin destrucción de DOM
