# Proposal: annual-assignment-ui (A3)

## Intent

Habilitar la pantalla de **inicio de año**: cada empleado (incluido RH) crea y edita sus metas anuales agrupadas en **categorías custom** (independientes de pilares de competencias), con **doble ponderación 100%** (categorías suman 100% y metas dentro de cada categoría suman 100%) y **KPIs** indicadores (numérico/porcentaje/moneda) vinculables a una o más metas. Jefes/directores/gerentes pueden ver y solicitar cambios (ajustar KPIs y ponderaciones) en metas de personas a cargo, sin agregar ni borrar metas. UI-first con fixtures JSON, sin API. Desbloquea A4 (medio año) y A5 (evaluación final).

## Scope

### In Scope

- CRUD de **categorías de metas custom** (independientes de pilares).
- CRUD de **metas** agrupadas dentro de cada categoría, con unidad (`%` o `$`) y peso.
- **Doble validación 100%**: suma de pesos de categorías = 100; suma de pesos de metas dentro de cada categoría = 100.
- **KPIs**: indicadores (numérico/porcentaje/moneda) reutilizables, vinculables a 1+ metas (relación N:M).
- **Vista de solo lectura** para jefes/directores/gerentes sobre las metas de personas a cargo, con botón "Solicitar cambio" (mock: solo abre modal de feedback sin persistencia).
- Fixtures JSON estáticas; carga síncrona.
- Reutilizar shell, tokens, selector dev persona de A1.
- Reutilizar `CustomSelect`, `EmptyState`, `PageSkeleton`, `ConfirmDeleteModal`, patrón de modales `<dialog>` y sentence case de A2.
- **Incluir perfil `rh`** en menú y acceso (decisión #7): RH también tiene metas.
- Estados UI: skeleton, vacío, error (heredados de A1).

### Data Model (types)

Tipos TS centralizados en `web/src/lib/types/goal.ts` (sin import desde `competency.ts`, decisión #5):

| Tipo | Campos clave | Notas |
|------|--------------|-------|
| `GoalUnit` | `'porcentaje' \| 'moneda' \| 'numero'` | Unidad de la meta |
| `KPI` | `id, name, unit, description` | Indicador reutilizable (N:M con metas) |
| `GoalKpiLink` | `goalId, kpiId` | Join N:M entre meta y KPI |
| `GoalCategory` | `id, name, description, weight (0..100)` | Categoría custom (peso obligatorio) |
| `Goal` | `id, categoryId, name, description, unit, weight (0..100), targetValue` | Meta dentro de categoría |
| `EmployeeAssignment` | `id, employeeId, categoryIds[], goalIds[]` | Asignación por empleado (perfil) |
| `ChangeRequest` | `id, assignmentId, goalId? , categoryId? , requesterId, message, createdAt` | Solicitud de cambio de un jefe (mock) |

`MANAGER_MAP` (mock plano de jerarquía) reemplazable en B4 (`org-hierarchy`) sin tocar UI.

### Components

Componentes nuevos en `web/src/lib/components/goals/`:

- `CategoryFormModal` — crear/editar categoría (nombre, descripción, peso).
- `GoalFormModal` — crear/editar meta (título, descripción, unidad, peso, valor objetivo, KPIs vinculados).
- `KpiFormModal` — alta/baja/edición de KPIs (opcional, para gestionar librería).
- `CategoryCard` — card con header (categoría + peso + badge de suma), tabla de metas hijas, acción "Nueva meta".
- `GoalRow` — fila de meta con KPIs como chips, acciones.
- `WeightIndicator` — progress + badge numérico (verde/ámbar); usado dos veces (categorías vs 100, metas vs 100 por categoría).
- `KpiBadge` — pill DaisyUI con nombre de KPI.
- `RequestChangeModal` — modal mock de feedback para jefes.
- `ReadOnlyBanner` — banner amarillo "Estás viendo las metas de {nombre}".
- `AssigneePicker` — selector de evaluado (visible para jefes en modo lectura).

Reutilizados: `EmptyState`, `PageSkeleton`, `ConfirmDeleteModal` (A2), `CustomSelect` (A2), `AppShell` y `Sidebar` (A1).

### Out of Scope (Non-goals)

- **Medio año**: edición de avances y revisión de KPIs (cambio A4).
- **Evaluación final**: autoevaluación y cierre RH (cambio A5).
- **Matriz 9×9**: potencial/desempeño por jefe (cambio A6).
- API real, persistencia, autenticación, RBAC servidor (recargar restaura fixtures).
- Agregación de metas entre personas; "mis evaluados" en sí mismo (cambio A7).
- Borrado de metas vía solicitud de cambio (decisión #8 explícita: jefe NO borra ni agrega).
- Notificaciones email (la solicitud de cambio es mock local).
- Jerarquía organizacional real (corporativa/retail): se mockea con un mapa plano "managerId".
- Importación desde Excel; wizard multi-step; plantillas de metas.
- Vínculo entre metas y competencias/pilares (decisión #5: categorías de metas son independientes).
- Cierre definitivo de criterios escala por perfil (regla abierta `openspec/config.yaml` l.97, ya cubierta por A2).

## Capabilities

### New Capabilities

- `annual-assignment-ui`: pantalla de inicio de año donde cada empleado crea categorías custom y metas con doble ponderación 100% y KPIs vinculables; refleja decisiones #1, #5, #6, #7 y #8. UI-first con fixtures JSON.

### Modified Capabilities

- `ui-shell`: extender `menuConfig.ts` para que "Asignación anual" sea visible también al perfil `rh` (decisión #7: RH también tiene metas). Sin cambio de layout ni de shell.

## Approach

- UI-first estricto: cero llamadas de red; fixtures importadas en `web/src/lib/fixtures/goals/`.
- Tipos TS centralizados en `web/src/lib/types/goal.ts` para que B3 (`goals-and-weighting`) y C4 (`goals-api`) los adopten sin acoplar a la estructura de mock.
- Ruta única `/objetivos/asignacion` que reemplaza el placeholder; no requiere sub-rutas. El mismo `+page.svelte` decide render:
  - Modo **dueño** (perfil activo = persona dueña de la asignación): editor completo.
  - Modo **lector/jefe** (perfil activo es jefe de la persona dueña): solo lectura con botón "Solicitar cambio".
  - Modo **RH-otros**: si el perfil activo es `rh` pero la persona dueña NO es la simulada actual, vista jefe (lectura + solicitar cambio).
- La persona "dueña" de la asignación se determina por el perfil activo + una constante de fixture (cada perfil ve su propia asignación; en este change mockeamos con `EmployeeAssignment.employeeId` mapeado a `EvaluationProfile`).
- Componentes DaisyUI (`btn`, `input`, `textarea`, `select`, `table`, `dialog`, `badge`, `alert`, `progress`).
- Accesibilidad WCAG 2.1 AA heredada de A1/A2; `aria-label` en inputs, `role="dialog"` en modales, foco visible.
- Validación de doble 100% en tiempo real con badges de estado (`badge-success` cuando suma = 100, `badge-warning` cuando no, deshabilita "Guardar" hasta cumplir).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modified | Reemplazar `EmptyState` placeholder por editor/visor real |
| `web/src/lib/types/goal.ts` | New | `Goal`, `GoalCategory`, `KPI`, `EmployeeAssignment`, `GoalUnit` |
| `web/src/lib/fixtures/goals/*.json` | New | `kpis.json`, `goal-categories.json`, `goals.json`, `assignments.json` |
| `web/src/lib/stores/goalsStore.svelte.ts` | New | Store con CRUD + validación 100% + detección modo (dueño/lector) |
| `web/src/lib/components/goals/*.svelte` | New | Componentes de metas (form, tabla, KPIs, badge, etc.) |
| `web/src/lib/nav/menuConfig.ts` | Modified | Agregar perfil `rh` al array `profiles` del ítem "Asignación anual" |
| `openspec/specs/annual-assignment-ui/spec.md` | New | Spec duradera de la capability |

## Decisions Reflected

| # | Decisión | Cómo se refleja en este change |
|---|----------|--------------------------------|
| #1 | Doble ponderación 100% | Dos badges de progreso independientes: (a) suma de pesos de categorías vs 100, (b) suma de pesos de metas dentro de cada categoría seleccionada vs 100. Botón "Guardar asignación" deshabilitado hasta que ambos cumplan. |
| #5 | Categorías custom independientes de pilares | `GoalCategory` es entidad propia sin FK a `Pillar`. Visualmente idéntica pero lógicamente disjunta del dominio `competency`. |
| #6 | KPIs indicadores vinculables a 1+ metas | `KPI` entidad propia + `goalKpiLinks: { goalId, kpiId }[]` para relación N:M. Cada meta puede tener 0..N KPIs y cada KPI puede alimentar 1..N metas. |
| #7 | Todos los perfiles (incluido RH) son evaluables | `menuConfig.ts`: añadir `rh` al array `profiles` del ítem "Asignación anual". Fixtures incluyen asignación para todos los 8 perfiles. |
| #8 | Jerarquía de edición: jefe ve y solicita cambios; NO borra ni agrega | Render condicional: si `viewerProfile` es jefe de la persona dueña (mock plano), ocultar botones "Nueva meta"/"Eliminar" y mostrar "Solicitar cambio" en su lugar. El modal de solicitud es mock (solo `alert-success` con texto del feedback). |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Doble validación 100% permita valores flotantes que rompan suma exacta | Med | Validación con tolerancia ε = 0.01; documentar en spec; alinear con B3 cuando se cierre detalle de rounding |
| Reutilización indebida de tipos de `competency.ts` (mezclar pilares con categorías) | Med | Tipos en archivo aparte (`goal.ts`); sin import cruzado desde `competency.ts` |
| Jerarquía mock plana no refleje corporativo/retail dual | Low | Documentar en design; cierre real en B4 (`org-hierarchy`); UI preparada para tomar `viewerIsManagerOf` por assignment |
| "Solicitar cambio" se confunda con edición real | Low | Botón claramente etiquetado "Solicitar cambio"; modal indica "no es una edición directa"; sin persistencia |
| RH espera que su menú ya tenga sub-ítems de admin al ganar acceso a asignación | Low | El ítem "Asignación anual" sigue siendo uno solo; menú RH conserva sus 3 ítems propios + este nuevo compartido |

## Rollback Plan

- Revertir el placeholder de `web/src/routes/objetivos/asignacion/+page.svelte` (es un solo archivo).
- Borrar `web/src/lib/types/goal.ts`, `web/src/lib/fixtures/goals/`, `web/src/lib/stores/goalsStore.svelte.ts`, `web/src/lib/components/goals/`.
- Revertir `menuConfig.ts` (sacar `rh` del array `profiles` de "Asignación anual").
- Si ya archivado: abrir change correctivo, no revertir archivos manualmente.

## Dependencies

- **A1 (`ui-shell-and-design-tokens`)**: COMPLETED, archivado en `openspec/changes/archive/2026-06-02-ui-shell-and-design-tokens/`. Proveedor de shell, tokens, `devContext` store, selector dev persona, `EmptyState`, `PageSkeleton`, `ErrorState`, `ForbiddenState`, `CustomSelect`.
- **A2 (`rh-competency-admin-ui`)**: COMPLETED, archivado en `openspec/changes/archive/2026-06-03-rh-competency-admin-ui/`. Referencia de patrones (modales `<dialog>`, `structuredClone`, `ConfirmDeleteModal`, sentence case, store runes). **Independiente lógicamente** (decisión #5): este change NO usa pilares ni competencias.
- **B3 (`goals-and-weighting`)**: paralelo en Fase B. Sincronizar al cierre: nombres de campos y reglas de redondeo.
- **B4 (`org-hierarchy`)**: paralelo en Fase B. La detección `viewerIsManagerOf` se mockea con un mapa plano; cuando B4 cierre, reemplazar el mock por el árbol real sin tocar UI.

## Success Criteria

- [ ] Ruta `/objetivos/asignacion` renderiza editor de metas con fixtures; sin errores de consola ni `tsc`.
- [ ] CRUD de categorías custom (crear, renombrar, eliminar) con cascada a metas hijas.
- [ ] CRUD de metas dentro de categoría con campos: título, descripción, unidad (`%` o `$`), peso, valor objetivo, KPI(s) vinculado(s).
- [ ] Validación en tiempo real: suma de pesos de categorías = 100 (badge verde cuando OK).
- [ ] Validación en tiempo real: suma de pesos de metas dentro de cada categoría = 100 (badge por categoría).
- [ ] Botón "Guardar asignación" deshabilitado mientras alguna validación falle.
- [ ] KPIs reutilizables: 5+ KPIs seed; cada meta puede vincular 0..N; un mismo KPI puede aparecer en varias metas.
- [ ] Perfil `rh` ve "Asignación anual" en menú y tiene su propia asignación editable.
- [ ] Vista de solo lectura para jefe: oculta botones crear/eliminar; muestra "Solicitar cambio" que abre modal mock.
- [ ] Recarga de página restaura fixtures (sin persistencia).
- [ ] Estados skeleton, vacío, error operativos.
- [ ] Sentence case, sin sombras decorativas, sin border-left/right.
- [ ] `pnpm run lint` y `tsc` sin errores.
- [ ] WCAG 2.1 AA: foco visible, `aria-label` en acciones, `<th>` semánticos, `role="dialog"` en modales.
