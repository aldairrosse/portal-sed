# Tasks: mid-year-progress-ui (A4)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~650 |
| 400-line budget risk | Medium |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No (under 400 per PR if split)

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Types, store, fixtures, ProgressIndicator | PR 1 | base = main; ~350 lines; extend Goal type, add CyclePhase, store getters/mutations, fixture cycle.json, ProgressIndicator component |
| 2 | GoalRow, CategoryCard, CommentPopover, page integration | PR 2 | base = main; ~300 lines; depends on PR 1; modify GoalRow/CategoryCard, create CommentPopover, wire page mode |

## Phase 1: Foundation (PR 1)

> Goal: tipos, store, fixtures y componente ProgressIndicator listos para PR 2.

- [x] 1.1 Extend `web/src/lib/types/goal.ts` — agregar `progress?: number`, `progressUpdatedAt?: string`, `comments?: GoalComment[]` a `Goal`. Crear interfaz `GoalComment` (`id`, `authorId`, `authorName`, `content`, `createdAt`). Crear tipo `CyclePhase = 'asignacion' | 'avance' | 'cierre'`. (~25 lines)
- [x] 1.2 Create `web/src/lib/fixtures/goals/cycle.json` — `{ "year": 2026, "phase": "avance" }`. (~3 lines)
- [x] 1.3 Extend `web/src/lib/stores/goalsStore.svelte.ts` — importar `cycle.json`, agregar state `cyclePhase`, crear getters: `getCyclePhase()`, `getGoalProgress(goalId)`, `getGoalComments(goalId)`, `getCategoryProgressAverage(categoryId)`, `getGoalPermissions(role, isOwner)`. (~60 lines)
- [x] 1.4 Extend store mutations — agregar `updateGoalProgress(goalId, progress)` (actualiza `progress` + `progressUpdatedAt`), `addGoalComment(goalId, authorId, authorName, content)`, `deleteGoalComment(goalId, commentId)`. Bloquear `deleteGoal`/`deleteCategory` en fase `avance`. (~50 lines)
- [x] 1.5 Extend fixtures `goals.json` — agregar `progress` y `comments: []` opcionales a 4-5 metas seed (valores variados: 0, 35, 65, 80, 100). (~20 lines diff)
- [x] 1.6 Create `web/src/lib/components/goals/ProgressIndicator.svelte` — barra de progreso DaisyUI `progress` + badge numérico. Props: `value` (number), `max` (number, default 100), `label?` (string). Colores: `progress-error` (< 40%), `progress-warning` (40–79%), `progress-success` (≥ 80%). Badge: `badge-error`/`badge-warning`/`badge-success`. (~60 lines)
- [x] 1.7 Run `pnpm run check` — verificar que tsc y lint pasan sin errores.

**Phase 1 acceptance criteria:**
- `CyclePhase` type existe y es importable.
- `getCyclePhase()` retorna `'avance'` desde fixture.
- `updateGoalProgress` actualiza `progress` y `progressUpdatedAt` en el store.
- `addGoalComment` inserta comentario con autor y fecha.
- `deleteGoal` lanza error silencioso en fase `avance`.
- `ProgressIndicator` renderiza barra con color correcto según valor.
- `pnpm run check` sin errores.

## Phase 2: UI Integration (PR 2)

> Goal: pantalla de avance completamente funcional — empleado edita avances, jefe edita avances de subordinados y comenta.

- [ ] 2.1 Modify `web/src/lib/components/goals/GoalRow.svelte` — en modo avance: renderizar campo `progress` inline editable (input number), ícono de comentario con badge count, sin botones Editar/Eliminar. Props adicionales: `phase`, `permissions`. (~50 lines diff)
- [ ] 2.2 Create `web/src/lib/components/goals/CommentPopover.svelte` — popover DaisyUI con lista de comentarios (autor, fecha relativa, contenido) + textarea para nuevo comentario + botón "Enviar". Se abre al clic en ícono de comentario. Props: `goalId`, `comments`, `onAdd`. (~80 lines)
- [ ] 2.3 Modify `web/src/lib/components/goals/CategoryCard.svelte` — en modo avance: renderizar `ProgressIndicator` agregado en header (promedio de avances), sin botón "Nueva meta", sin Editar/Eliminar categoría. (~30 lines diff)
- [ ] 2.4 Modify `web/src/routes/objetivos/asignacion/+page.svelte` — leer `getCyclePhase()`, pasar `phase` y `permissions` a componentes hijos. En modo avance: ocultar botones "Nueva categoría" y "Guardar asignación", mostrar `ProgressIndicator` global con promedio de avance. (~40 lines diff)
- [ ] 2.5 Update `ReadOnlyBanner` text — en modo avance: "Estás viendo las metas de {nombre}. Puedes editar avances y agregar comentarios." (~5 lines diff)
- [ ] 2.6 Write unit tests — `goalsStore`: `updateGoalProgress`, `addGoalComment`, `deleteGoalComment`, `getCategoryProgressAverage`, `getGoalPermissions`, bloqueo delete en fase avance. Vitest. (~100 lines)
- [ ] 2.7 Run `pnpm run check` — verificar que tsc y lint pasan sin errores.

**Phase 2 acceptance criteria:**
- `/objetivos/asignacion` en fase avance muestra campo de avance por meta.
- Empleado puede editar avance y agregar comentarios en sus metas.
- Jefe puede editar avance y agregar comentarios en metas de subordinados.
- Botones de crear/eliminar no existen en fase avance.
- Campos peso, targetValue, nombre, unidad son read-only en fase avance.
- `ProgressIndicator` muestra color correcto (< 40% rojo, 40–79% amarillo, ≥ 80% verde).
- `CommentPopover` muestra lista de comentarios y permite agregar nuevos.
- `ReadOnlyBanner` muestra texto correcto según fase.
- Unit tests cubren: updateGoalProgress, addGoalComment, deleteGoalComment, getCategoryProgressAverage, getGoalPermissions, bloqueo delete.
- `pnpm run check` sin errores.

## Files Summary

| File | Action | Phase | Est. lines |
|------|--------|-------|-----------|
| `web/src/lib/types/goal.ts` | Modify | 1 | 25 |
| `web/src/lib/fixtures/goals/cycle.json` | Create | 1 | 3 |
| `web/src/lib/stores/goalsStore.svelte.ts` | Modify | 1 | 110 |
| `web/src/lib/fixtures/goals/goals.json` | Modify | 1 | 20 (diff) |
| `web/src/lib/components/goals/ProgressIndicator.svelte` | Create | 1 | 60 |
| `web/src/lib/components/goals/GoalRow.svelte` | Modify | 2 | 50 (diff) |
| `web/src/lib/components/goals/CommentPopover.svelte` | Create | 2 | 80 |
| `web/src/lib/components/goals/CategoryCard.svelte` | Modify | 2 | 30 (diff) |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modify | 2 | 40 (diff) |
| `web/src/lib/components/goals/ReadOnlyBanner.svelte` | Modify | 2 | 5 (diff) |
| Tests (Vitest) | Create | 2 | 100 |
| **Total** | | | **~523** |

## Dependencies Graph (inter-phase)

```
Phase 1 (PR 1)  ──→ Phase 2 (PR 2)
```

- Phase 1: tipos, store, fixtures, ProgressIndicator.
- Phase 2: depende de Phase 1; modifica componentes existentes y crea CommentPopover.

## Out-of-scope (no tasks)

- Evaluación 1-5 (A5), matriz 9×9 (A6).
- Persistencia, API, auth.
- Edición de KPIs en fase avance.
- Notificaciones email.
