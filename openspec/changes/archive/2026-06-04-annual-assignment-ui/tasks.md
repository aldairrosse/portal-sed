# Tasks: annual-assignment-ui (A3)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~1450 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 → PR 2 → PR 3 |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Foundation — types, fixtures, store, menu | PR 1 | base = main; ~420 lines; types + 5 fixtures + store + menu edit + `MANAGER_MAP` constant |
| 2 | Editor UI — modals, cards, weight indicator, route | PR 2 | base = main; ~480 lines; depends on PR 1; 10 new components + page integration |
| 3 | Read-only / request-change mode + verification | PR 3 | base = main; ~400 lines; depends on PR 1; `ReadOnlyBanner`, `RequestChangeModal`, `AssigneePicker`, mode switch in page, tests, verify |

> **Dependency graph:**
> - PR 1 ← (no deps) ← implementable alone.
> - PR 2 ← PR 1 ← implementable after PR 1.
> - PR 3 ← PR 1 ← implementable in parallel with PR 2 (touches only page-level mode logic + a few new components). May be merged after PR 2 to avoid touching `+page.svelte` twice.

## Phase 1: Foundation (PR 1)

> Goal: types, fixtures, store, menu, and `MANAGER_MAP` ready so PRs 2 and 3 can build on them.

- [x] 1.1 Create `web/src/lib/types/goal.ts` — `GoalUnit`, `KpiUnit`, `KPI`, `GoalKpiLink`, `GoalCategory`, `Goal`, `ChangeRequest`, `EmployeeAssignment` interfaces. Export `MANAGER_MAP` constant. (~110 lines)
- [x] 1.2 Create `web/src/lib/fixtures/goals/kpis.json` — 6 KPIs seed (mix `porcentaje`/`moneda`/`numero`). (~25 lines)
- [x] 1.3 Create `web/src/lib/fixtures/goals/goal-categories.json` — 4 categorías seed con pesos (40, 30, 20, 10). (~30 lines)
- [x] 1.4 Create `web/src/lib/fixtures/goals/goals.json` — 8–10 metas seed distribuidas en las 4 categorías. (~80 lines)
- [x] 1.5 Create `web/src/lib/fixtures/goals/goal-kpi-links.json` — 10–12 links N:M seed (KPIs reutilizados). (~40 lines)
- [x] 1.6 Create `web/src/lib/fixtures/goals/assignments.json` — 8 `EmployeeAssignment` (uno por perfil), con `managerId` mock para 3 perfiles (jefe, director, gerente-tienda). (~70 lines)
- [x] 1.7 Create `web/src/lib/stores/goalsStore.svelte.ts` — `$state` con `structuredClone`, getters (incl. `isAssignmentValid`, `isCategoryGoalsWeightValid`, `getManagerOf`) y mutaciones (CRUD + cascada + N:M + `recordChangeRequest`). (~260 lines)
- [x] 1.8 Modify `web/src/lib/nav/menuConfig.ts` — agregar `'rh'` al array `profiles` del ítem "Asignación anual" (decisión #7). (~3 lines diff)

**Phase 1 acceptance criteria:**
- `pnpm run lint` y `tsc` sin errores.
- `import` de los 5 fixtures desde el store funciona y `structuredClone` evita mutaciones del JSON original.
- `isAssignmentValid` retorna `false` para suma de categorías ≠ 100 y `true` para suma = 100 ± 0.01.
- `MANAGER_MAP['colaborador'] === 'jefe'`, `MANAGER_MAP['rh'] === 'director'`, etc.
- El ítem "Asignación anual" aparece en el sidebar para todos los 8 perfiles (incluido `rh`).

## Phase 2: Editor UI (PR 2)

> Goal: pantalla de inicio de año completamente funcional en **modo editor** (dueño). El modo lector llega en PR 3.

- [x] 2.1 Create `web/src/lib/components/goals/WeightIndicator.svelte` — progress + badge numérico (verde/ámbar), usado dos veces (categorías global + metas por categoría). (~80 lines)
- [x] 2.2 Create `web/src/lib/components/goals/KpiBadge.svelte` — pill DaisyUI con nombre de KPI. (~30 lines)
- [x] 2.3 Create `web/src/lib/components/goals/CategoryFormModal.svelte` — modal crear/editar categoría (nombre, descripción, peso). Reusa patrón de `PillarFormModal` (A2). (~120 lines)
- [x] 2.4 Create `web/src/lib/components/goals/GoalFormModal.svelte` — modal crear/editar meta (nombre, descripción, `unit`, peso, `targetValue`, KPIs checkboxes). (~200 lines)
- [x] 2.5 Create `web/src/lib/components/goals/KpiFormModal.svelte` — modal opcional para gestionar librería de KPIs. (~100 lines)
- [x] 2.6 Create `web/src/lib/components/goals/GoalRow.svelte` — fila de meta con KPIs como chips, acciones editar/eliminar (modo editor). (~90 lines)
- [x] 2.7 Create `web/src/lib/components/goals/CategoryCard.svelte` — card con header (categoría + peso + `WeightIndicator` interno), tabla de `GoalRow`, botón "Nueva meta". (~110 lines)
- [x] 2.8 Modify `web/src/routes/objetivos/asignacion/+page.svelte` — reemplazar `EmptyState` placeholder por layout: header con `WeightIndicator` global, lista de `CategoryCard`, botón "Nueva categoría", `KpiFormModal` y modales condicionales. Modo fijo a `editor` en esta fase (lectura llega en PR 3). (~200 lines)

**Phase 2 acceptance criteria:**
- Renderiza `/objetivos/asignacion` con las fixtures del perfil activo (cualquiera de los 8).
- CRUD de categorías y metas funciona; cascada al eliminar.
- `WeightIndicator` global y por categoría actualiza en tiempo real al cambiar pesos.
- Botón "Guardar asignación" deshabilitado cuando alguna validación falla.
- Validaciones de formulario (`targetValue > 0`, `weight 0..100`, nombre único) muestran `alert-error`.
- KPIs se vinculan/desvinculan desde `GoalFormModal`; chips visibles en `GoalRow`.
- Recarga de página restaura fixtures (sin persistencia).
- `pnpm run lint` y `tsc` sin errores.

## Phase 3: Read-only / Request-change mode + verification (PR 3)

> Goal: completar la pantalla con el modo lector (jefe/director/gerente) + "Solicitar cambio" mock, y verificar todo el change.

- [x] 3.1 Create `web/src/lib/components/goals/ReadOnlyBanner.svelte` — banner amarillo "Estás viendo las metas de {nombre}. Solo puedes solicitar cambios.". (~40 lines)
- [x] 3.2 Create `web/src/lib/components/goals/RequestChangeModal.svelte` — modal mock: contexto (categoría o meta) en read-only + textarea + "Enviar solicitud" → `alert-success` + insert `ChangeRequest`. (~120 lines)
- [x] 3.3 Create `web/src/lib/components/goals/AssigneePicker.svelte` — selector simple (visible para perfiles con subordinados según `MANAGER_MAP` inverso) para cambiar entre "evaluados" mock. (~80 lines)
- [x] 3.4 Modify `web/src/routes/objetivos/asignacion/+page.svelte` — agregar detección de modo (`editor` vs `reader`) vía `getManagerOf`; ocultar botones de crear/editar/eliminar y mostrar `ReadOnlyBanner` + "Solicitar cambio" en modo lector. Renderizar `AssigneePicker` cuando aplique. (~80 lines diff)
- [x] 3.5 Modify `web/src/lib/components/goals/GoalRow.svelte` — agregar render condicional de acciones: editor → Editar/Eliminar; lector → "Solicitar cambio". (~25 lines diff)
- [x] 3.6 Modify `web/src/lib/components/goals/CategoryCard.svelte` — agregar render condicional: editor → "Nueva meta" + Editar/Eliminar categoría; lector → "Solicitar cambio" de categoría. (~25 lines diff)
- [x] 3.7 Write unit tests — `goalsStore`: CRUD, cascada (categoría → metas → links), `linkKpiToGoal` idempotente, `recordChangeRequest`, `isAssignmentValid` con tolerancias, `isCategoryGoalsWeightValid` con categoría vacía, `getManagerOf`. Vitest. (~150 lines)
- [x] 3.8 Write integration tests — `+page.svelte` renderiza editor para el dueño y reader para el jefe. Svelte Testing Library + `fireEvent`. (~80 lines)
- [x] 3.9 Verify decisión #1 — no hay input de peso en pilares ni en competencias; doble 100% se valida en metas y categorías. (`tsc` + grep test).
- [x] 3.10 Verify decisión #7 — perfil `rh` ve el menú "Asignación anual" y puede editar su propia asignación.
- [x] 3.11 Verify decisión #8 — perfil `jefe` ve asignación de `colaborador` con banner de solo lectura y botón "Solicitar cambio" (sin botones de crear/eliminar).
- [x] 3.12 Verify general — `pnpm run lint` y `tsc` con cero errores; sin console errors ni warnings; recargar restaura fixtures; WCAG 2.1 AA (foco visible, `aria-label`, `role="dialog"`, `<th>` semánticos, sentence case).

**Phase 3 acceptance criteria:**
- Perfil `jefe` activo → modo lector sobre la asignación de `colaborador` (su subordinado en `MANAGER_MAP`).
- Perfil `director` activo → modo lector sobre `jefe`, `regional`.
- Perfil `rh` activo → modo editor sobre su propia asignación; modo lector si elige ver la asignación de otro (vía `AssigneePicker`).
- "Solicitar cambio" abre `RequestChangeModal`, al confirmar muestra `alert-success` y guarda `ChangeRequest` en el store.
- Unit tests cubren: CRUD, cascada, validación doble 100% (con tolerancias), `MANAGER_MAP`, `recordChangeRequest`.
- Integración: al cambiar dev persona a `jefe` y recargar, ve la asignación de `colaborador` con banner; al cambiar a `colaborador`, ve su editor.
- `pnpm run lint` y `tsc` sin errores.
- Sentence case, sin sombras decorativas, sin border-left/right.
- WCAG 2.1 AA: foco visible, `aria-label` en acciones, `role="dialog"` en modales.

## Files Summary

| File | Action | Phase | Est. lines |
|------|--------|-------|-----------|
| `web/src/lib/types/goal.ts` | Create | 1 | 110 |
| `web/src/lib/fixtures/goals/kpis.json` | Create | 1 | 25 |
| `web/src/lib/fixtures/goals/goal-categories.json` | Create | 1 | 30 |
| `web/src/lib/fixtures/goals/goals.json` | Create | 1 | 80 |
| `web/src/lib/fixtures/goals/goal-kpi-links.json` | Create | 1 | 40 |
| `web/src/lib/fixtures/goals/assignments.json` | Create | 1 | 70 |
| `web/src/lib/stores/goalsStore.svelte.ts` | Create | 1 | 260 |
| `web/src/lib/nav/menuConfig.ts` | Modify | 1 | 3 (diff) |
| `web/src/lib/components/goals/WeightIndicator.svelte` | Create | 2 | 80 |
| `web/src/lib/components/goals/KpiBadge.svelte` | Create | 2 | 30 |
| `web/src/lib/components/goals/CategoryFormModal.svelte` | Create | 2 | 120 |
| `web/src/lib/components/goals/GoalFormModal.svelte` | Create | 2 | 200 |
| `web/src/lib/components/goals/KpiFormModal.svelte` | Create | 2 | 100 |
| `web/src/lib/components/goals/GoalRow.svelte` | Create | 2 | 90 |
| `web/src/lib/components/goals/CategoryCard.svelte` | Create | 2 | 110 |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modify | 2 + 3 | 280 (200 + 80 diff) |
| `web/src/lib/components/goals/ReadOnlyBanner.svelte` | Create | 3 | 40 |
| `web/src/lib/components/goals/RequestChangeModal.svelte` | Create | 3 | 120 |
| `web/src/lib/components/goals/AssigneePicker.svelte` | Create | 3 | 80 |
| `GoalRow.svelte` (modify for reader actions) | Modify | 3 | 25 (diff) |
| `CategoryCard.svelte` (modify for reader actions) | Modify | 3 | 25 (diff) |
| Tests (Vitest + Testing Library) | Create | 3 | 230 |
| **Total** | | | **~1450** |

## Dependencies Graph (inter-phase)

```
Phase 1 (PR 1)  ─────────────────────┐
                                     ├──→ Phase 2 (PR 2)  ──┐
                                     │                       ├──→ Phase 3 (PR 3) verifies all
                                     └──→ Phase 3 (PR 3)  ──┘
```

- Phase 1: standalone; no other phase can start without it.
- Phase 2: depends on Phase 1 (types, store, fixtures).
- Phase 3: depends on Phase 1 (store `recordChangeRequest`, `getManagerOf`); touches `+page.svelte` and 2 components, so it can land either after PR 2 (cleaner) or in parallel with PR 2 on a separate branch.

## Out-of-scope (no tasks)

- Persistencia, API, auth: ver `proposal.md` → Non-goals.
- Medio año (A4), evaluación final (A5), 9×9 (A6), mis evaluados (A7).
- Vínculo metas ↔ pilares/competencias (decisión #5).
- Notificaciones email reales.
