# mid-year-progress-ui Specification

## Purpose

Pantalla de **medio año** donde el empleado edita avances en sus metas existentes y agrega comentarios, sin poder eliminar metas ni editar pesos (decisión #3). El jefe/director/gerente puede **editar avances** y **agregar comentarios** en metas de subordinados (decisión #8). Indicador visual de progreso por meta y por categoría. UI-first con fixtures JSON, sin API ni persistencia.

## Screens

| Screen | Route | Description |
|--------|-------|-------------|
| Avance de metas | `/objetivos/asignacion` | Misma ruta que A3; se activa modo "avance" cuando `cyclePhase === 'avance'`. Editor de avances + comentarios. |

## Data Model

Entidades adicionales a las de A3:

| Entity | Fields | Notes |
|--------|--------|-------|
| **Goal** (extendido) | `progress?: number` (0–100 o valor absoluto), `progressUpdatedAt?: string` (ISO), `comments?: GoalComment[]` | Campos opcionales; en fase `asignación` no existen. |
| **GoalComment** | `id`, `authorId`, `authorName`, `content`, `createdAt` (ISO) | Comentario de empleado o jefe sobre el avance de una meta. |
| **CycleState** | `year`, `phase` (`asignacion` \| `avance` \| `cierre`) | Estado global del ciclo; controla el modo de la página. |

## Requirements

### Requirement: Modo avance en la página de asignación

El sistema SHALL activar el modo "avance" cuando `cyclePhase === 'avance'`. En este modo, la página SHALL renderizar campos de avance y ocultar botones de crear/eliminar metas y categorías.

#### Scenario: Empleado ve modo avance

- GIVEN dev persona activa = `colaborador` y `cyclePhase = 'avance'`
- WHEN navega a `/objetivos/asignacion`
- THEN ve sus metas con campo de avance editable por cada meta
- AND NO ve botones "Nueva categoría", "Nueva meta", "Editar categoría", "Eliminar categoría"
- AND NO ve botones "Editar meta", "Eliminar meta"
- AND ve `ReadOnlyBanner` no aplica (es dueño)

#### Scenario: Jefe ve modo avance en subordinado

- GIVEN dev persona activa = `jefe` y `cyclePhase = 'avance'`
- WHEN navega a `/objetivos/asignacion`
- THEN ve la asignación de `colaborador` (su subordinado)
- AND ve campo de avance editable en cada meta del subordinado
- AND ve campo de comentario accesible por cada meta
- AND NO ve botones de crear/eliminar
- AND ve `ReadOnlyBanner` "Estás viendo las metas de {nombre}. Puedes editar avances y agregar comentarios."

### Requirement: Registro de avance por meta

El sistema SHALL permitir al empleado y al jefe editar el avance de una meta. El avance SHALL ser un número: porcentaje (0–100) para unidad `porcentaje`, valor absoluto (≥ 0) para `moneda` o `numero`. El sistema SHALL mostrar un indicador visual de progreso por meta.

#### Scenario: Empleado edita avance de su meta

- GIVEN meta con `unit = 'porcentaje'` y `targetValue = 100`
- WHEN el empleado ingresa `progress = 65` en el campo de avance
- THEN el `ProgressIndicator` muestra barra al 65% con color amarillo (40–79%)
- AND `progressUpdatedAt` se actualiza a la fecha actual

#### Scenario: Jefe edita avance de meta subordinada

- GIVEN jefe en modo avance sobre la meta del colaborador
- WHEN el jefe cambia `progress` de 65 a 80
- THEN el `ProgressIndicator` muestra barra al 80% con color verde (≥ 80%)
- AND `progressUpdatedAt` se actualiza

#### Scenario: Avance con unidad moneda

- GIVEN meta con `unit = 'moneda'` y `targetValue = 50000`
- WHEN el empleado ingresa `progress = 32000`
- THEN el `ProgressIndicator` muestra barra proporcional (32000/50000 = 64%)
- AND el badge muestra "$32,000 / $50,000"

### Requirement: Indicador visual de progreso

El sistema SHALL mostrar un `ProgressIndicator` por meta y un indicador agregado por categoría. Los colores SHALL seguir la semántica: rojo (< 40%), amarillo (40–79%), verde (≥ 80%).

#### Scenario: ProgressIndicator por meta

- GIVEN meta con `progress = 45` y `targetValue = 100`
- WHEN se renderiza la `GoalRow`
- THEN el `ProgressIndicator` muestra barra al 45% con clase `progress-warning`
- AND badge numérico "45%"

#### Scenario: ProgressIndicator por categoría

- GIVEN categoría con 3 metas: avances 80, 60, 100
- WHEN se renderiza la `CategoryCard`
- THEN el indicador de categoría muestra promedio = 80%
- AND color verde (`progress-success`)

#### Scenario: Categoría sin avances

- GIVEN categoría con metas que no tienen `progress` definido
- WHEN se renderiza la `CategoryCard`
- THEN el indicador muestra 0% con color rojo (`progress-error`)
- AND badge "Sin avance"

### Requirement: Comentarios por meta

El sistema SHALL permitir al empleado y al jefe agregar comentarios de texto libre a cualquier meta. Los comentarios SHALL mostrar autor y fecha. El jefe SHALL poder comentar en metas de subordinados.

#### Scenario: Empleado agrega comentario

- GIVEN meta en modo avance, empleado es dueño
- WHEN hace clic en el ícono de comentario y escribe "Avanzado según plan"
- THEN el comentario aparece en la lista con autor "María López" y fecha actual
- AND el campo de textarea se limpia

#### Scenario: Jefe agrega comentario en meta subordinada

- GIVEN jefe en modo avance sobre meta del colaborador
- WHEN escribe "Buen avance, mantener ritmo" y confirma
- THEN el comentario aparece con autor del jefe
- AND el empleado podrá verlo en su vista

#### Scenario: Lista de comentarios

- GIVEN meta con 3 comentarios previos
- WHEN se abre el popover de comentarios
- THEN se muestran los 3 comentarios en orden cronológico (más antiguo arriba)
- AND cada comentario muestra: autor, fecha relativa, contenido

### Requirement: Bloqueo de eliminación en fase avance

El sistema SHALL prohibir la eliminación de metas y categorías cuando `cyclePhase === 'avance'`. Los botones de eliminar NO SHALL renderizarse.

#### Scenario: Sin botón eliminar meta

- GIVEN `cyclePhase = 'avance'`
- WHEN se renderiza cualquier `GoalRow`
- THEN NO existe botón "Eliminar" en las acciones

#### Scenario: Sin botón eliminar categoría

- GIVEN `cyclePhase = 'avance'`
- WHEN se renderiza cualquier `CategoryCard`
- THEN NO existe botón "Eliminar" en el header de la categoría

### Requirement: Campos bloqueados en fase avance

El sistema SHALL bloquear la edición de `weight`, `targetValue`, nombre, descripción y unidad de metas cuando `cyclePhase === 'avance'`. Solo `progress` y `comments` son editables.

#### Scenario: Peso no editable

- GIVEN meta en modo avance
- WHEN el usuario intenta editar el peso de la meta
- THEN el campo peso NO SHALL ser editable (readonly o sin botón de edición)
- AND no hay modal de edición de meta disponible

#### Scenario: Solo avance y comentario editables

- GIVEN meta en modo avance
- WHEN se renderiza la `GoalRow`
- THEN los campos visibles son: nombre (read-only), peso (read-only), targetValue (read-only), KPIs (read-only), avance (editable), comentarios (accesible)

### Requirement: Update optimista local sin reload

El sistema DEBE (SHALL) actualizar `storeState.data.goals[n].progress` localmente al invocar `updateGoalProgress`, sin disparar `reload()`. El PATCH al backend DEBE ejecutarse en segundo plano. Si el PATCH falla, el store DEBE hacer rollback al valor anterior. Este comportamiento reemplaza el actual que invoca `reload()` incondicionalmente tras cada PATCH.

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

El sistema DEBE (SHALL) aplicar un debounce de 500ms dentro de `updateGoalProgress` para que múltiples keystrokes rápidos generen un solo PATCH. El debounce DEBE vivir en el store, NO en `GoalRow.svelte`.

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

El sistema DEBE (SHALL) omitir el PATCH si el nuevo valor es idéntico al `progress` actual de la meta en el store. Esto evita requests innecesarios y duplicados en `goal_progress_logs`.

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

El componente `GoalRow.svelte` NO DEBE (SHALL NOT) ser modificado. El debounce y la lógica optimista viven exclusivamente en el store `goalsStore.svelte.ts`. `GoalRow.svelte` sigue llamando a `updateGoalProgress` como lo hace actualmente (via `oninput`/`onchange`).

#### Scenario: GoalRow.svelte intacto

- GIVEN el change aplicado
- WHEN se revisa el diff
- THEN `GoalRow.svelte` no aparece modificado
- AND el binding `oninput` sigue igual
- AND el comportamiento de pérdida de foco está resuelto por cambios exclusivamente en el store

## UI Components

| Component | Description |
|-----------|-------------|
| `ProgressIndicator` | Barra de progreso + badge numérico. Colores: `progress-error` (< 40%), `progress-warning` (40–79%), `progress-success` (≥ 80%). Reutilizable para meta individual y agregado de categoría. |
| `CommentPopover` | Popover con lista de comentarios y textarea para agregar nuevo. Se abre al hacer clic en ícono de comentario en `GoalRow`. |
| `GoalRow` (modificado) | En modo avance: campo `progress` inline editable, ícono de comentario con badge de count. Sin botones Editar/Eliminar. |
| `CategoryCard` (modificado) | En modo avance: `ProgressIndicator` agregado en header. Sin botón "Nueva meta". Sin Editar/Eliminar categoría. |
| `ReadOnlyBanner` (modificado) | Texto actualizado para modo avance: "Puedes editar avances y agregar comentarios." |

## Store

### Getters adicionales

| Función | Retorna |
|---------|---------|
| `getCyclePhase()` | `CyclePhase` |
| `getGoalProgress(goalId)` | `number \| undefined` |
| `getGoalComments(goalId)` | `GoalComment[]` |
| `getCategoryProgressAverage(categoryId)` | `number` (promedio de avances de metas de la categoría) |
| `getGoalPermissions(role, isOwner)` | `{ canEditProgress, canComment, canEditWeight, canDelete }` |

### Mutations adicionales

| Función | Efectos colaterales |
|---------|---------------------|
| `updateGoalProgress(goalId, progress)` | Actualiza `progress` y `progressUpdatedAt`; solo en fase `avance` |
| `addGoalComment(goalId, authorId, authorName, content)` | Inserta `GoalComment` en la meta |
| `deleteGoalComment(goalId, commentId)` | Elimina un comentario (solo el autor) |

### Restricciones

- `deleteGoal` y `deleteCategory` SHALL lanzar error si `cyclePhase === 'avance'`.
- `updateGoal` SHALL ignorar cambios a `weight`, `targetValue`, `name`, `description`, `unit` en fase `avance` (solo `progress` se actualiza).

## Validations

| Regla | Enforcement | UX |
|-------|-------------|----|
| `progress ≥ 0` | Input `min=0` + validación submit | `alert-error` |
| `progress ≤ targetValue` para moneda/numero | Validación submit | `alert-error` "El avance no puede superar el valor objetivo" |
| `progress ∈ [0, 100]` para porcentaje | Input `min=0 max=100` | `alert-error` |
| Comentario no vacío | Validación submit | Botón deshabilitado si textarea vacío |
| Eliminación bloqueada en avance | `deleteGoal`/`deleteCategory` | Error silencioso (botón no existe en UI) |

## Fixtures

### Modificado

| Archivo | Cambio |
|---------|--------|
| `goals.json` | Agregar `progress` (opcional) y `comments` (array vacío) a metas seed |

### Nuevo

| Archivo | Contenido |
|---------|-----------|
| `cycle.json` | `{ "year": 2026, "phase": "avance" }` — estado del ciclo para controlar modo |

## Out of scope

- Evaluación 1-5 de competencias (A5)
- Matriz 9×9 (A6)
- Eliminar metas (prohibido en medio año, decisión #3)
- Persistencia, API, auth
- Edición de KPIs vinculados en fase avance
- Notificaciones email
