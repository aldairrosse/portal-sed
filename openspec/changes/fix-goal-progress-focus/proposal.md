# Proposal: Fix Goal Progress Focus Loss + goal_progress_logs

## Intent

Corregir la pérdida de foco del input de progreso en `GoalRow.svelte` durante la fase medio-año/fin-año, y agregar la tabla `goal_progress_logs` para registrar el historial de cambios de progreso sin exponer API.

## Root Cause

`updateGoalProgress` en `goalsStore.svelte.ts:451-457` hace `PATCH /goals/{goalId}/progress` y luego `await reload()` incondicionalmente. Esto dispara `storeState.loading = true` en `_doLoad()`, lo que causa que `PageSkeleton` en `+page.svelte:445-500` destruya el DOM completo del formulario y lo reemplace con un skeleton de carga, perdiendo el foco del input.

Además, `GoalRow.svelte:223` usa `oninput` sin debounce, lo que genera una llamada PATCH por cada keystroke, acelerando el ciclo de destrucción-reconstrucción del DOM.

El patrón correcto ya existe en el mismo store: `addGoalComment` (líneas 472-492) hace update optimista local sin `reload()`, preservando el DOM y el foco.

## Scope

### In Scope
- Frontend: `updateGoalProgress` con update optimista local (sin reload) + debounce 500ms + guarda de valor sin cambio
- Backend: migración goose `000027_goal_progress_logs` con FK cascade a goals
- Backend: insert en `goal_progress_logs` dentro de `UpdateGoalCurrentValue` solo si el valor cambió
- Ent schema opcional para `goal_progress_logs` (si el patrón del proyecto lo requiere)

### Out of Scope
- Endpoint REST para consultar `goal_progress_logs` (YAGNI — sin caso de uso)
- UI para visualizar historial de progreso (YAGNI — sin caso de uso actual)
- Cambios a `goal_comments` u otras tablas existentes
- Modificar `PageSkeleton` o el condicional de carga en `+page.svelte`

## Capabilities

### New Capabilities
- `goal-progress-logs`: Tabla de auditoría para registrar cambios de `current_value` en `goals`

### Modified Capabilities
- `mid-year-progress-ui`: El input de progreso no pierde el foco al escribir; una sola llamada PATCH tras 500ms de inactividad
- `goals-and-weighting`: `UpdateGoalCurrentValue` ahora registra historial en `goal_progress_logs` y solo actualiza cuando el valor realmente cambia

## Approach

1. **Frontend — update optimista**: Modificar `updateGoalProgress` en el store para actualizar `storeState.data` localmente (mismo patrón que `addGoalComment`) en lugar de invocar `reload()`. Agregar debounce de 500ms para que múltiples keystrokes rápidos generen un solo PATCH.

2. **Frontend — guarda sin cambio**: Si el valor recibido es igual al `current_value` actual en el store, no hacer PATCH (evita requests innecesarios y duplicados de log).

3. **Backend — migración**: Crear `000027_goal_progress_logs.up.sql` con tabla `goal_progress_logs`: `id UUID PK DEFAULT gen_random_uuid()`, `goal_id UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE`, `value FLOAT NOT NULL`, `recorded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()`, `created_by UUID REFERENCES employees(id)`.

4. **Backend — inserción condicional**: Modificar `UpdateGoalCurrentValue` para: (a) dentro de una transacción, comparar `current_value` actual con el nuevo valor usando `WHERE current_value IS DISTINCT FROM $newValue`; (b) si cambió, ejecutar UPDATE y luego INSERT en `goal_progress_logs`; (c) si no cambió, no hacer nada y retornar el row actual.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/lib/stores/goalsStore.svelte.ts` | Modified | `updateGoalProgress`: optimista local + debounce + guarda |
| `api/cmd/server/migrations/000027_goal_progress_logs.up.sql` | New | Crear tabla `goal_progress_logs` |
| `api/cmd/server/migrations/000027_goal_progress_logs.down.sql` | New | Dropear tabla |
| `api/internal/repository/goal/goal_repo.go` | Modified | `UpdateGoalCurrentValue`: transacción + insert condicional en log |
| `api/internal/schema/` | Optional | Ent schema para `goal_progress_logs` si el patrón lo requiere |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Debounce muy corto aún genera múltiples PATCH | Low | 500ms es conservador para typing humano; ajustable |
| Update optimista local puede divergir si el PATCH falla | Low | Rollback local en catch (restaurar valor anterior) |
| Migración falla en FK si `employees` no tiene datos | Low | `created_by` es nullable; sin FK a employees el registro igual se inserta |
| Competencia de escritura (dos tabs editando mismo goal) | Very Low | Última escritura gana; optimistic locking no es necesario para progreso |

## Rollback Plan

Revertir el commit. La migración tiene down.sql. El store frontend mantiene compatibilidad con la API existente (el endpoint PATCH no cambia su contrato).

## Dependencies

- `goals-and-weighting` spec (tabla `goals`, campo `current_value`)
- `mid-year-progress-ui` spec (componente `GoalRow.svelte`, `ProgressIndicator`)
- Migración `000026` (última aplicada)
- Tabla `employees` existente para FK de `created_by`

## Success Criteria

- [ ] E2E: El foco permanece en el input de progreso después de que el valor se persiste
- [ ] E2E: Solo se emite un PATCH tras 500ms de inactividad en el input (no uno por keystroke)
- [ ] E2E: No se inserta registro duplicado en `goal_progress_logs` al reescribir el mismo valor
- [ ] Test: `UpdateGoalCurrentValue` con mismo valor no modifica `goals` ni inserta en `goal_progress_logs`
- [ ] Test: `updateGoalProgress` en store actualiza `storeState.data` localmente (sin `reload`)
- [ ] `pnpm run check` y `go test ./...` pasan
