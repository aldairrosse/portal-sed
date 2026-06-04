# annual-assignment-ui Specification

## Purpose

Pantalla de **inicio de año** donde cada empleado (incluido RH) crea y edita sus metas anuales agrupadas en **categorías custom** (independientes de los pilares de competencias), con **doble ponderación 100%** (categorías suman 100% y metas dentro de cada categoría suman 100%) y **KPIs** indicadores (numérico/porcentaje/moneda) vinculables a una o más metas. Jefes/directores/gerentes pueden **ver** las metas de personas a cargo y **solicitar cambios** (ajustar KPIs y ponderaciones) sin agregar ni borrar metas. UI-first con fixtures JSON, sin API ni persistencia. Refleja las decisiones #1 (doble ponderación), #5 (categorías custom independientes), #6 (KPIs N:M), #7 (todos los perfiles evaluables, incluido RH) y #8 (jerarquía: ver + solicitar cambios, no agregar/borrar).

## Screens

| Screen | Route | Description |
|--------|-------|-------------|
| Asignación anual | `/objetivos/asignacion` | Editor de metas del dueño activo (modo editor) o lector + "Solicitar cambio" (modo jefe/lector). Una sola ruta; sub-rutas no aplican. |

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **KPI** | `id`, `name`, `unit` (`porcentaje` \| `moneda` \| `numero`), `description` | Indicador reutilizable. Independiente de cualquier meta. |
| **GoalKpiLink** | `goalId`, `kpiId` | Join N:M. Misma meta puede tener varios KPIs; mismo KPI puede alimentar varias metas. |
| **GoalCategory** | `id`, `name`, `description`, `weight` (0..100) | Categoría custom (no confundir con pilar). El peso es obligatorio. |
| **Goal** | `id`, `categoryId`, `name`, `description`, `unit` (`porcentaje` \| `moneda` \| `numero`), `weight` (0..100), `targetValue` (`> 0`) | Meta dentro de una categoría. `unit` y `targetValue` son independientes del KPI vinculado. |
| **EmployeeAssignment** | `id`, `employeeId`, `categoryIds[]`, `goalIds[]` | Una asignación por perfil de evaluación; mapea empleado simulado a sus categorías y metas. |
| **ChangeRequest** | `id`, `assignmentId`, `goalId?`, `categoryId?`, `requesterId`, `message`, `createdAt` (ISO) | Solicitud de cambio de un jefe/gerente. `goalId` y `categoryId` son mutuamente excluyentes. Mock local, sin persistencia. |

Perfiles de evaluación: 8 perfiles (`colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `rh`). `director-general` queda fuera hasta B4 (`org-hierarchy`).

Mock de jerarquía (`MANAGER_MAP`): `Record<employeeId, managerId>` plano; reemplazable por el árbol corporativo/retail real cuando B4 cierre.

## Requirements

### Requirement: Edición de categorías de metas custom

El sistema SHALL permitir a un empleado crear, renombrar, reponderar y eliminar sus **categorías de metas custom**. Las categorías son **independientes** de los pilares de competencias (decisión #5). Eliminar una categoría SHALL eliminar en cascada todas sus metas hijas (sin persistir el cambio en disco: recargar restaura fixtures).

#### Scenario: Crear categoría

- GIVEN empleado autenticado (cualquier perfil, incluido `rh`) en `/objetivos/asignacion`
- WHEN hace clic en "Nueva categoría", completa nombre, descripción y peso, y confirma
- THEN la categoría aparece en la lista con su peso
- AND la `WeightIndicator` global se actualiza mostrando la nueva suma parcial vs 100

#### Scenario: Eliminar categoría con metas hijas

- GIVEN categoría con al menos una meta hija
- WHEN confirma eliminación en `ConfirmDeleteModal`
- THEN la categoría y todas sus metas se eliminan del store local
- AND cualquier `GoalKpiLink` que apunte a esas metas se elimina también (cascada)
- AND la `WeightIndicator` global se recalcula

#### Scenario: Nombre único por asignación

- GIVEN empleado editando sus categorías
- WHEN intenta guardar una categoría con un nombre que ya existe en su asignación
- THEN el modal muestra `alert-error` "Ya existe una categoría con ese nombre."
- AND no se persiste el cambio

### Requirement: Edición de metas dentro de una categoría

El sistema SHALL permitir crear, editar y eliminar metas dentro de cada categoría. Cada meta SHALL tener: `name`, `description`, `unit` (`porcentaje` \| `moneda` \| `numero`), `weight` (0..100), `targetValue` (`> 0`) y 0..N KPIs vinculados.

#### Scenario: Crear meta

- GIVEN categoría con peso asignado
- WHEN hace clic en "Nueva meta" dentro de la `CategoryCard`, completa campos y confirma
- THEN la meta aparece en la tabla de la categoría
- AND la `WeightIndicator` de la categoría se actualiza con la nueva suma parcial vs 100

#### Scenario: Validar `targetValue` positivo

- GIVEN modal de meta abierto
- WHEN intenta guardar con `targetValue <= 0`
- THEN el modal muestra `alert-error` y no guarda

#### Scenario: Validar peso entre 0 y 100

- GIVEN modal de meta abierto
- WHEN intenta guardar con `weight < 0` o `weight > 100`
- THEN el input rechaza el valor (atributo `min=0 max=100` + validación al submit)

### Requirement: Doble validación 100% en tiempo real

El sistema SHALL mostrar en tiempo real dos `WeightIndicator` independientes:

1. **Suma de pesos de categorías** del empleado actual: SHALL ser 100 ± 0.01 para habilitar el guardado.
2. **Suma de pesos de metas** dentro de cada categoría: SHALL ser 100 ± 0.01 por categoría, o la categoría SHALL estar **vacía** (0 metas) — en ese caso la categoría se marca con `badge-warning` "Sin metas" pero no bloquea el guardado (decisión transitoria; ver `design.md` Open Questions).

#### Scenario: Badge verde al cumplir 100%

- GIVEN empleado con 3 categorías de pesos 40, 30, 30 (suma = 100)
- WHEN observa la `WeightIndicator` global
- THEN muestra `progress-success` con valor 100 y badge `badge-success` "100%"

#### Scenario: Badge ámbar al no cumplir

- GIVEN empleado con 2 categorías de pesos 50, 30 (suma = 80)
- WHEN observa la `WeightIndicator` global
- THEN muestra `progress-warning` con valor 80 y badge `badge-warning` "80% — faltan 20"
- AND el botón "Guardar asignación" está deshabilitado

#### Scenario: Categoría con 0 metas

- GIVEN categoría recién creada sin metas hijas
- WHEN observa la `CategoryCard`
- THEN la `WeightIndicator` interna muestra estado neutral con badge `badge-warning` "Sin metas"
- AND la `WeightIndicator` global se calcula solo con las categorías que tienen peso > 0, pero el guardado sigue bloqueado hasta que **todas** las categorías con peso tengan suma 100 (o estén explícitamente marcadas como vacías — ver Open Questions)

#### Scenario: Tolerancia flotante

- GIVEN empleado con pesos que suman 99.99
- WHEN observa la `WeightIndicator` global
- THEN el badge se muestra verde (tolerancia ε = 0.01)

### Requirement: Vinculación de KPIs a metas (N:M)

El sistema SHALL permitir crear una librería de KPIs (nombre + unidad + descripción) y vincular 0..N KPIs a cada meta. Un mismo KPI SHALL poder alimentar 1..N metas.

#### Scenario: Vincular KPI a meta

- GIVEN meta en edición
- WHEN selecciona uno o más KPIs de la lista (checkboxes dentro de `GoalFormModal`)
- THEN los KPIs aparecen como `KpiBadge` chips en la fila de la meta
- AND se crea/elimina un `GoalKpiLink` correspondiente en el store

#### Scenario: KPI reutilizado en varias metas

- GIVEN KPI "Ingresos trimestrales" vinculado a 2 metas
- WHEN edita la meta 1
- THEN la meta 2 sigue mostrando el `KpiBadge` "Ingresos trimestrales"
- AND la lista de "KPIs en uso" en la librería de KPIs lo refleja

### Requirement: Jerarquía de edición (decisión #8)

El sistema SHALL detectar si el usuario activo es **dueño** de la asignación o **jefe/lector**. En modo lector SHALL ocultar todos los botones de crear/editar/eliminar y SHALL mostrar "Solicitar cambio" por categoría y por meta. El modal de solicitud es **mock** (sin persistencia, sin email).

#### Scenario: Modo editor para el dueño

- GIVEN dev persona activa = `colaborador`
- WHEN navega a `/objetivos/asignacion`
- THEN ve su propia asignación con todos los botones (Nueva categoría, Nueva meta, Editar, Eliminar)

#### Scenario: Modo lector para jefe

- GIVEN dev persona activa = `jefe`
- WHEN navega a `/objetivos/asignacion`
- THEN `MANAGER_MAP['colaborador'] === 'jefe'`, por lo que el sistema detecta que el `jefe` ve la asignación de `colaborador`
- AND muestra `ReadOnlyBanner` "Estás viendo las metas de María López García. Solo puedes solicitar cambios."
- AND oculta los botones Nueva/Editar/Eliminar
- AND muestra "Solicitar cambio" en cada categoría y meta

#### Scenario: Solicitar cambio (mock)

- GIVEN modo lector activo sobre la meta X de la persona Y
- WHEN hace clic en "Solicitar cambio" en la meta X
- THEN se abre `RequestChangeModal` con el contexto de la meta X (read-only) y un textarea de feedback
- WHEN confirma el envío
- THEN se inserta un `ChangeRequest` en el store local
- AND se muestra `alert-success` "Tu solicitud fue registrada"
- AND el modal se cierra (sin email, sin API)

#### Scenario: RH como dueño (decisión #7)

- GIVEN dev persona activa = `rh`
- WHEN navega a `/objetivos/asignacion`
- THEN ve su propia asignación en **modo editor** (RH también tiene metas, decisión #7)

#### Scenario: RH viendo a otro

- GIVEN dev persona activa = `rh`
- WHEN (futuro) abre la asignación de otro perfil
- THEN entra en modo lector (RH administra catálogo de competencias, no metas ajenas en esta fase)

### Requirement: Acceso RH al menú (decisión #7)

El sistema SHALL hacer visible el ítem "Asignación anual" del sidebar para el perfil `rh`, igual que para los otros 7 perfiles.

#### Scenario: RH ve el menú

- GIVEN dev persona activa = `rh`
- WHEN observa el `Sidebar`
- THEN ve el ítem "Asignación anual" con icono `Target` apuntando a `/objetivos/asignacion`
- AND puede navegar a la ruta como cualquier otro perfil

## UI Components

| Component | Description |
|-----------|-------------|
| `WeightIndicator` | Progress bar + badge numérico. Verde (`progress-success` + `badge-success`) cuando suma = 100 ± 0.01; ámbar (`progress-warning` + `badge-warning`) en otro caso. Muestra el déficit/exceso exacto. |
| `CategoryCard` | Card DaisyUI con header (nombre, peso, badge de suma de metas), tabla de `GoalRow`, botón "Nueva meta" (modo editor). |
| `GoalRow` | Fila de meta: nombre, unidad, peso, `targetValue`, `KpiBadge[]`, acciones. Acciones condicionales: editor → Editar/Eliminar; lector → "Solicitar cambio". |
| `KpiBadge` | Pill DaisyUI con nombre de KPI; tooltip opcional con `unit` y `description`. |
| `CategoryFormModal` | Modal `<dialog>` con inputs: nombre (requerido, único por asignación), descripción (requerida), peso (0–100). Botones: "Crear categoría" / "Editar categoría", "Cancelar". |
| `GoalFormModal` | Modal `<dialog>` con: nombre, descripción, `unit` (select), peso, `targetValue` (numérico), lista de checkboxes para KPIs (0..N). Validación inline. |
| `KpiFormModal` | Modal opcional para gestionar la librería de KPIs. |
| `RequestChangeModal` | Modal mock: contexto (categoría o meta) en read-only + textarea + botón "Enviar solicitud". Muestra `alert-success` al confirmar. |
| `ReadOnlyBanner` | Banner amarillo (`alert alert-warning`) sobre la página cuando el modo es lector. |
| `AssigneePicker` | Selector simple (visible para jefes en modo lectura) para cambiar de "evaluado" entre los 3 mock seed. |

Reutilizados de A1/A2: `EmptyState`, `PageSkeleton`, `ErrorState`, `ConfirmDeleteModal`, `CustomSelect`, `AppShell`, `Sidebar`.

### State classes

| Estado | DaisyUI class | Uso |
|--------|---------------|-----|
| Loading | `skeleton` via `PageSkeleton` | Carga inicial de fixtures (timeout simulado 300ms; opcional, porque la carga es síncrona) |
| Vacío | `EmptyState` | Sin asignación o sin categorías |
| Error | `alert-error` | Fallo en mutación (validación) |
| Success | `alert-success` | Confirmación CRUD y "Solicitar cambio" enviada |
| Forbidden | `ForbiddenState` | No aplica en esta ruta (todos los perfiles tienen acceso) |
| Read-only | `alert-warning` (banner) | Modo lector/jefe |
| Weight OK | `progress-success` + `badge-success` | Suma de pesos = 100 |
| Weight KO | `progress-warning` + `badge-warning` | Suma de pesos ≠ 100 |

Botones: `btn-primary` (guardar/crear), `btn-ghost` (cancelar), `btn-error` (eliminar), `btn-warning` (solicitar cambio). Modales: `<dialog>` nativo con `modal`, `modal-box`, `modal-action`. Tablas: `table-zebra`. Sin estilos inline ni CSS raw. Sentence case en todos los textos visibles.

## Fixtures

Archivos JSON en `web/src/lib/fixtures/goals/`:

| Archivo | Contenido |
|---------|-----------|
| `kpis.json` | 6 KPIs seed (mix `porcentaje` / `moneda` / `numero`): ej. "Ingresos trimestrales", "NPS clientes", "Tickets resueltos", "Tasa de rotación", "Margen bruto", "Horas de capacitación" |
| `goal-categories.json` | 4 categorías seed con pesos sugeridos (ej. "Resultados de negocio" 40, "Desarrollo de personas" 30, "Operaciones" 20, "Innovación" 10) |
| `goals.json` | 8–10 metas seed distribuidas en las categorías |
| `goal-kpi-links.json` | 10–12 links N:M seed (algunos KPIs en 2–3 metas) |
| `assignments.json` | 8 `EmployeeAssignment` (uno por perfil de evaluación), con `managerId` mock para 3 perfiles (jefe, director, gerente-tienda) |

Carga síncrona via `import`. Sin `fetch`. El store usa `structuredClone` para evitar mutar el fixture original.

## Store

`web/src/lib/stores/goalsStore.svelte.ts` — store reactivo Svelte 5 (runas `$state`).

### Getters

| Función | Retorna |
|---------|---------|
| `getKpis()` | `KPI[]` |
| `getCategories()` | `GoalCategory[]` |
| `getGoals()` | `Goal[]` |
| `getLinks()` | `GoalKpiLink[]` |
| `getAssignments()` | `EmployeeAssignment[]` |
| `getChangeRequests()` | `ChangeRequest[]` |
| `getAssignmentForEmployee(employeeId)` | `EmployeeAssignment \| undefined` |
| `getCategoriesForAssignment(assignmentId)` | `GoalCategory[]` filtrado |
| `getGoalsForCategoryInAssignment(categoryId, assignmentId)` | `Goal[]` filtrado |
| `getKpisForGoal(goalId)` | `KPI[]` filtrado via `GoalKpiLink` |
| `getGoalsForKpi(kpiId)` | `Goal[]` filtrado via `GoalKpiLink` |
| `getManagerOf(employeeId)` | `string \| undefined` (de `MANAGER_MAP`) |
| `getCategoriesWeightSum(assignmentId)` | `number` |
| `getGoalsWeightSumForCategory(categoryId, assignmentId)` | `number` |
| `isAssignmentValid(assignmentId)` | `boolean` (categorías 100 + metas 100 por categoría no vacía) |
| `isCategoryGoalsWeightValid(categoryId, assignmentId)` | `boolean` |

### Mutations

| Función | Efectos colaterales |
|---------|---------------------|
| `addKpi(k)` / `updateKpi(id, u)` / `deleteKpi(id)` | CRUD KPIs (sin cascada porque no hay hijos) |
| `addCategory(c)` / `updateCategory(id, u)` / `deleteCategory(id)` | CRUD categorías; `deleteCategory` cascada a metas y links |
| `addGoal(g)` / `updateGoal(id, u)` / `deleteGoal(id)` | CRUD metas; `deleteGoal` cascada a `GoalKpiLink` |
| `linkKpiToGoal(goalId, kpiId)` / `unlinkKpiFromGoal(goalId, kpiId)` | N:M |
| `assignCategoryToEmployee(assignmentId, categoryId)` / `unassignCategoryFromEmployee(...)` | Wiring forward-compat (no usado en UI en A3) |
| `assignGoalToEmployee(assignmentId, goalId)` / `unassignGoalFromEmployee(...)` | Wiring forward-compat (no usado en UI en A3) |
| `recordChangeRequest(req)` | Inserta `ChangeRequest` (mock local) |

> **Nota:** en A3 cada perfil ve su asignación **completamente poblada** desde fixtures; los métodos de asignación/un-asignación están cableados pero no se invocan desde la UI. Quedan para que B3 (`goals-and-weighting`) y C4 (`goals-api`) los adopten.

## Validations

| Regla | Enforcement | UX |
|-------|-------------|----|
| Suma de pesos de categorías = 100 ± 0.01 | `isAssignmentValid` | `WeightIndicator` global; botón "Guardar" deshabilitado |
| Suma de pesos de metas por categoría = 100 ± 0.01 (o categoría vacía) | `isCategoryGoalsWeightValid` | `WeightIndicator` por categoría |
| Nombre de categoría único por asignación | `addCategory` / `updateCategory` | `alert-error` en modal |
| `targetValue > 0` | submit del `GoalFormModal` | `alert-error` |
| `weight` ∈ [0, 100] | input `min=0 max=100` + validación al submit | `alert-error` |
| Categoría con 0 metas | `isCategoryGoalsWeightValid` | `badge-warning` "Sin metas" (no bloquea) |
| `GoalKpiLink` no duplicado | `linkKpiToGoal` | Insert idempotente; sin error si ya existe |

## Security baseline (checklist)

- [x] Validación de entrada: HTML5 + validador de formulario en submit (no `innerHTML`).
- [x] Sin almacenamiento de tokens (no aplica: sin auth).
- [x] Sin concatenación de SQL (no aplica: sin backend).
- [x] Render de comentarios/freetext escapado por Svelte (defensa por default).
- [x] Modales con `role="dialog"` y `aria-modal`; cierre por backdrop o ESC; foco restaurado.
- [x] Inputs de peso `type="number"` con `min`/`max` y `inputmode="decimal"`.

## Non-functional

- **Accesibilidad**: WCAG 2.1 AA — contraste mínimo, foco visible (`focus-visible:ring`), `aria-label` en botones de acción, `<th>` semánticos, `role="dialog"` en modales.
- **Responsive**: Tablas con scroll horizontal en viewports < `md`; cards colapsables en móvil; modales full-width en móvil.
- **Rendimiento**: Lazy route único (code-splitting). Sin llamadas de red. Carga síncrona de fixtures. `structuredClone` en inicialización del store.
- **Estilo**: Sin box-shadow decorativo, sin border-left/right como acento. Respeto a `prefers-reduced-motion`.
- **Store**: Svelte 5 runas (`$state`) con `structuredClone` en inicialización. Mutaciones inmutables (nuevo array en cada cambio).
- **Trazabilidad**: `ChangeRequest` incluye `requesterId` y `createdAt` para auditoría (mock local).

## Out of scope explícito

Ver `proposal.md` → "Out of Scope (Non-goals)". Resumen:

- **Medio año**: edición de avances y revisión de KPIs (cambio A4).
- **Evaluación final**: autoevaluación y cierre RH (cambio A5).
- **Matriz 9×9**: potencial/desempeño por jefe (cambio A6).
- API real, persistencia, autenticación, RBAC servidor.
- Agregación de metas entre personas; "mis evaluados" en sí mismo (cambio A7).
- Borrado de metas vía solicitud de cambio (decisión #8: jefe NO borra ni agrega).
- Notificaciones email (la solicitud de cambio es mock local).
- Vínculo entre metas y competencias/pilares (decisión #5).
- Importación desde Excel; wizard multi-step; plantillas de metas.
