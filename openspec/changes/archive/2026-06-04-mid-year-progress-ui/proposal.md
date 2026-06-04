## Why

A3 entregó la pantalla de inicio de año donde cada empleado crea sus metas. Ahora necesitamos la fase de medio año: editar metas existentes y registrar avances, sin eliminar metas (decisión #3). Es el segundo paso del ciclo anual antes de la evaluación final (A5).

## What Changes

- Extender la pantalla de asignación anual para soportar la fase "avance" del ciclo.
- Editor de metas: campos editables restringidos según rol y fase.
- Registro de avance por meta: valor numérico o porcentaje, con indicador visual de progreso.
- Bloqueo de eliminación de metas en fase "avance" (para todos los roles).
- Indicador de avance global (semáforo o barra) por categoría y total.
- Comentarios del jefe por meta: campo de texto para observaciones sobre el avance.
- Fixtures extendidos con metas precargadas, datos de avance y comentarios.

### Permisos por rol en fase "avance"

| Rol | Editar avance | Agregar comentario | Editar peso | Editar valor objetivo | Eliminar meta |
|-----|:---:|:---:|:---:|:---:|:---:|
| Empleado (dueño) | ✅ | ✅ (propio) | ❌ | ❌ | ❌ |
| Jefe/director/gerente (lector) | ✅ (subordinados) | ✅ (subordinados) | ❌ | ❌ | ❌ |
| RH | ✅ (propio) | ✅ (propio) | ❌ | ❌ | ❌ |

- **Empleado:** solo edita avance y agrega comentarios en sus propias metas. No puede editar pesos, valores objetivo ni eliminar metas.
- **Jefe/director/gerente:** puede editar avance y agregar comentarios en metas de subordinados (vista solo lectura de pesos y estructura). No puede borrar ni agregar metas, ni modificar ponderaciones.
- **RH:** behave como empleado en su propia asignación.

## Capabilities

### New Capabilities

- `mid-year-progress-ui`: Pantalla de medio año — edición de metas existentes, registro de avances, indicadores de progreso, bloqueo de eliminación. UI-first con fixtures JSON.

### Modified Capabilities

- `goal-assignment-ui`: Reutiliza la ruta `/objetivos/asignacion` pero con modo "avance" (fase del ciclo). Se agregan reglas de edición restringida y bloqueo de eliminar. El spec existente se extiende con un delta.

## Impact

- **Archivos afectados:**
  - `web/src/routes/objetivos/asignacion/+page.svelte` — modo "avance" condicional
  - `web/src/lib/stores/goalsStore.svelte.ts` — función `updateGoalProgress`, `addGoalComment`, bloqueo delete en fase avance
  - `web/src/lib/components/goals/GoalRow.svelte` — campo avance inline, indicador visual, campo comentario
  - `web/src/lib/components/goals/CategoryCard.svelte` — progress indicator por categoría
  - `web/src/lib/types/goal.ts` — campo `progress` en Goal, campo `comments`, tipo `CyclePhase`
  - Nuevos componentes: `ProgressIndicator.svelte`, `GoalProgressRow.svelte`
- **Fixtures:** `goals.json` extendido con `progress`, `cyclePhase: "avance"`.
- **Decisiones reflejadas:** #3 (editar avances, prohibido eliminar metas), #8 (jefe puede editar avance y comentar, no borrar ni editar pesos).
- **Depende de:** A3 (mismas fixtures y store extendidos).
- **Non-goals:** evaluación 1-5 final, matriz 9×9, borrado de metas, persistencia, API, auth.
