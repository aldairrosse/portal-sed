# Delta: mid-year-progress-ui (MODIFIED)

## Propósito

Corregir la pérdida de foco del input de progreso en `GoalRow.svelte` durante la fase medio-año. La causa raíz es que `updateGoalProgress` en el store invoca `reload()` después de cada PATCH, lo que dispara `storeState.loading = true` y causa que `PageSkeleton` destruya el DOM completo del formulario. La solución es adoptar el patrón de update optimista local que ya existe en `addGoalComment` (líneas 472-492 del store).

## Requirements

### Requirement: Update optimista local sin reload

El sistema DEBE actualizar `storeState.data.goals[n].progress` localmente al invocar `updateGoalProgress`, sin disparar `reload()`. El PATCH al backend DEBE ejecutarse en segundo plano. Si el PATCH falla, el store DEBE hacer rollback al valor anterior. Este comportamiento reemplaza el actual que invoca `reload()` incondicionalmente tras cada PATCH.

#### Scenario: Progreso actualizado sin perder foco

- GIVEN input de progreso con foco activo y valor `65`
- WHEN usuario modifica a `70` y después de 500ms de inactividad se dispara el PATCH
- THEN `storeState.data.goals[n].progress` se actualiza a `70` localmente ANTES del PATCH
- AND el DOM del input NO se destruye (no hay `reload()`)
- AND el foco permanece en el input
- AND el `ProgressIndicator` refleja `70%`

#### Scenario: Rollback en error de PATCH

- GIVEN meta con `progress = 65` en el store
- WHEN usuario cambia a `70`, el PATCH falla con error de red
- THEN `storeState.data.goals[n].progress` vuelve a `65`
- AND el input muestra `65`
- AND el error se propaga para que el componente pueda mostrar feedback

#### Scenario: Múltiples metas — solo la afectada cambia

- GIVEN store con metas A (progress=50), B (progress=80)
- WHEN `updateGoalProgress(goalA_id, 55)`
- THEN solo `goals[goalA].progress` cambia a 55
- AND `goals[goalB].progress` sigue siendo 80
- AND no hay `reload()` que resetee todo el store

### Requirement: Debounce de 500ms

El sistema DEBE aplicar un debounce de 500ms dentro de `updateGoalProgress` para que múltiples keystrokes rápidos generen un solo PATCH. El debounce DEBE vivir en el store, NO en `GoalRow.svelte`.

#### Scenario: Múltiples keystrokes → un PATCH

- GIVEN usuario escribe "1", "2", "3" en rápida sucesión (intervalos < 500ms)
- WHEN el debounce se resuelve tras 500ms del último keystroke
- THEN solo se emite UN PATCH con `current_value: 123`
- AND el store se actualiza optimistamente con `123` desde el primer keystroke

#### Scenario: Debounce por meta independiente

- GIVEN usuario edita meta A y meta B simultáneamente (dos inputs)
- WHEN cada input tiene su propio timer de debounce
- THEN el debounce de meta A no interfiere con el de meta B
- AND cada meta emite su propio PATCH tras 500ms de inactividad en su input

#### Scenario: Llamada rápida cancela el timer anterior

- GIVEN `updateGoalProgress(goalId, 50)` llamado, timer de 500ms iniciado
- WHEN a los 200ms se llama `updateGoalProgress(goalId, 55)`
- THEN el timer anterior se cancela
- AND nuevo timer de 500ms inicia con valor 55
- AND solo se emite PATCH con `current_value: 55`

### Requirement: Guarda de valor sin cambio

El sistema DEBE omitir el PATCH si el nuevo valor es idéntico al `progress` actual de la meta en el store. Esto evita requests innecesarios y duplicados en `goal_progress_logs`.

#### Scenario: Mismo valor → no PATCH

- GIVEN meta con `progress = 45` en el store
- WHEN se llama `updateGoalProgress(goalId, 45)`
- THEN no se emite PATCH
- AND la promesa se resuelve inmediatamente (no-op)
- AND el store no se modifica

#### Scenario: Valor cambia → PATCH emitido

- GIVEN meta con `progress = 45`
- WHEN se llama `updateGoalProgress(goalId, 50)`
- THEN se emite PATCH con `current_value: 50`
- AND el store se actualiza optimistamente

### Requirement: GoalRow.svelte sin cambios

El componente `GoalRow.svelte` NO DEBE ser modificado. El debounce y la lógica optimista viven exclusivamente en el store `goalsStore.svelte.ts`. `GoalRow.svelte` sigue llamando a `updateGoalProgress` como lo hace actualmente (via `oninput`/`onchange`).

#### Scenario: GoalRow.svelte intacto

- GIVEN el change aplicado
- WHEN se revisa el diff
- THEN `GoalRow.svelte` no aparece modificado
- AND el binding `oninput` sigue igual
- AND el comportamiento de pérdida de foco está resuelto por cambios exclusivamente en el store

## Store contract

### Firma sin cambios

```ts
export async function updateGoalProgress(goalId: string, progress: number): Promise<void>
```

### Efectos colaterales (nuevos)

1. Cancela cualquier debounce pendiente para `goalId`
2. Compara `progress` con `goal.progress` actual en store → no-op si igual
3. Actualiza `storeState.data.goals[n].progress` y `progressUpdatedAt` inmediatamente
4. Inicia timer de 500ms
5. Al expirar: emite PATCH
6. Si PATCH falla: restaura `progress` anterior en store, rechaza promesa
7. Si PATCH ok: resuelve promesa

### Efectos colaterales (eliminados)

- ~~`await reload()`~~ — Ya no se invoca. Eliminado.

## Non-goals

- No se modifica `GoalRow.svelte` (el debounce y optimismo viven en el store)
- No se modifica `+page.svelte` ni `PageSkeleton`
- No se modifica el endpoint `PATCH /goals/{goalId}/progress` (el contrato es el mismo)
- No se agrega UI de feedback de error (el comportamiento actual ya muestra errores)
- No se modifica `oninput` por `onchange` en GoalRow (permanece igual)
