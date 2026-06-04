# Design: annual-assignment-ui (A3)

## Technical Approach

Single-route editor at `/objetivos/asignacion` that replaces the A1 placeholder. The page reads the **dev persona** from `devContext.svelte` and decides between two render modes (owner editor vs manager read-only) using a flat `managerMap` fixture. A `goalsStore.svelte.ts` (Svelte 5 runes, `structuredClone`) holds the goals, categories, KPIs, and per-employee assignments; mutations enforce the **double-100%** validation in real time. KPI linking is N:M via a join array. All manager requests are a mocked local modal (no persistence, no email, no API) per decision #8.

Inherits the A2 patterns: native `<dialog>` modals, DaisyUI tables, sentence case, accessibility baseline, immutable updates with spread.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| State management | Single `goalsStore.svelte.ts` with `$state` runes | Per-route local state; Svelte stores legacy | Same pattern as `competencyStore.svelte.ts`; cross-component reads/writes for KPIs + goals + assignments |
| Fixture loading | Static `import` of JSON at store init | `fetch()` at route load | Synchronous; reload resets state; matches A1/A2 spec |
| Mode detection (owner vs manager) | Flat `managerMap: Record<employeeId, employeeId>` derived from `assignments.json` | Tree traversal; B4 org-hierarchy spec | Mock per B4-not-closed-yet; UI ready to swap for tree lookup |
| Double 100% validation | Two derived values: `categoriesWeightSum` and `goalsWeightSumByCategory`; both compared with `Math.abs(sum - 100) < 0.01` | Single combined validator | Two independent checks per decision #1 (doble, no combinada) |
| KPI linking | N:M via `goalKpiLinks: { goalId, kpiId }[]` array | `kpiIds: string[]` embedded on Goal; separate join table | Matches relational shape; same goal can have same KPI multiple times if needed; B3/C4 will use this directly |
| Goal form granularity | One `GoalFormModal` for create/edit (fields: title, description, unit, weight, targetValue, kpiIds[]) | Multi-step wizard; per-unit variants | A2 precedent: one modal per entity; fields small enough to fit one dialog |
| Category form | Single `CategoryFormModal` (name + weight) | Reuse of generic PillarFormModal | Categories are weight-bearing (different from A2 pillars); different validation; clear separation |
| Manager request | `RequestChangeModal` mock: form for feedback text; on submit shows `alert-success` and clears the form. No persistence. | Send to backend; queue email | Decision #8 forbids delete/add; request is feedback, not a CRUD mutation. Email/postponed to A7 or later. |
| Save button gating | Disabled when either validation fails OR form is unchanged | Always enabled with inline error | Matches A2 `AcceptanceLevelEditor` pattern (button enabled only on changes); stricter here: also require valid sums |
| KPI selector | Multi-select via checkbox list inside `GoalFormModal` (small set of 5-10 KPIs) | `CustomSelect` with multi; token chip input | 5-10 KPIs is small; checkboxes are clearest; reused `CustomSelect` would need new multi-mode |

## Data Flow

```
fixtures/goals/*.json ──import──→ goalsStore.svelte.ts ──$derived──→ Route page
                                          ↑                              │
                                          └──── mutation fns ────────────┘
                                                                 │
                                          ┌── viewerIsManagerOf() ──┐
                                          │                        │
                              devContext.svelte ──→ page decides editor vs reader
```

Store initializes from fixtures on first access. Page reads via `$derived` getters. Mutations: `addCategory`, `updateCategory`, `deleteCategory` (cascade), `addGoal`, `updateGoal`, `deleteGoal`, `addKpi`, `updateKpi`, `deleteKpi`, `linkKpiToGoal`, `unlinkKpiFromGoal`, `recordChangeRequest` (mock). All mutations use spread (immutable).

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/types/goal.ts` | Create | `Goal`, `GoalCategory`, `KPI`, `EmployeeAssignment`, `GoalUnit`, `GoalKpiLink`, `ChangeRequest` |
| `web/src/lib/fixtures/goals/kpis.json` | Create | 6 seed KPIs (mix de numérico/%/$): ej. "Ingresos trimestrales", "NPS clientes", "Tickets resueltos" |
| `web/src/lib/fixtures/goals/goal-categories.json` | Create | 4 seed categorías con pesos sugeridos (ej. "Resultados de negocio" 40, "Desarrollo de personas" 30, "Operaciones" 20, "Innovación" 10) |
| `web/src/lib/fixtures/goals/goals.json` | Create | 8-10 metas seed distribuidas en las categorías |
| `web/src/lib/fixtures/goals/goal-kpi-links.json` | Create | 10-12 links N:M seed (algunos KPIs en 2-3 metas) |
| `web/src/lib/fixtures/goals/assignments.json` | Create | 8 `EmployeeAssignment` (uno por perfil de evaluación), con `managerId` mock para 3 perfiles (jefe, director, gerente-tienda) |
| `web/src/lib/stores/goalsStore.svelte.ts` | Create | `$state` store: CRUD + `categoriesWeightSum`, `goalsWeightSumByCategory`, `isAssignmentValid`, `viewerIsManagerOf`, `recordChangeRequest` |
| `web/src/lib/components/goals/CategoryFormModal.svelte` | Create | Modal de crear/editar categoría custom (nombre, peso) |
| `web/src/lib/components/goals/GoalFormModal.svelte` | Create | Modal de crear/editar meta (título, descripción, unidad, peso, valor objetivo, KPIs vinculados) |
| `web/src/lib/components/goals/KpiFormModal.svelte` | Create | Modal de crear/editar KPI (nombre, tipo de unidad, descripción) |
| `web/src/lib/components/goals/CategoryCard.svelte` | Create | Card con header (categoría + peso + badge de suma), tabla de metas hijas, botón "Nueva meta" |
| `web/src/lib/components/goals/GoalRow.svelte` | Create | Fila de meta: nombre, unidad, peso, targetValue, KPIs (badges), acciones |
| `web/src/lib/components/goals/WeightIndicator.svelte` | Create | Progress bar + badge numérico: usado dos veces (categorías vs 100, metas vs 100 por categoría) |
| `web/src/lib/components/goals/KpiBadge.svelte` | Create | Pill DaisyUI con nombre de KPI (usado en filas de meta y vista resumen) |
| `web/src/lib/components/goals/RequestChangeModal.svelte` | Create | Modal mock: campo de texto + área de feedback, cierra con `alert-success` |
| `web/src/lib/components/goals/ReadOnlyBanner.svelte` | Create | Banner amarillo arriba del editor cuando entra como jefe: "Estás viendo las metas de {nombre}. Solo puedes solicitar cambios." |
| `web/src/lib/components/goals/AssigneePicker.svelte` | Create | Selector simple (visible para jefes) para cambiar entre "mis evaluados" mock (3 personas seed) |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modify | Reemplazar `EmptyState` por layout completo: header con `WeightIndicator` global, lista de `CategoryCard`, `RequestChangeModal` condicional, `ReadOnlyBanner` condicional |
| `web/src/lib/nav/menuConfig.ts` | Modify | Agregar `'rh'` al array `profiles` del ítem "Asignación anual" (decisión #7) |

## Interfaces / Contracts

```typescript
// web/src/lib/types/goal.ts

export type GoalUnit = 'porcentaje' | 'moneda' | 'numero';

export type KpiUnit = GoalUnit; // same allowed units at KPI level

export interface KPI {
  id: string;
  name: string;
  unit: KpiUnit;
  description: string;
}

export interface GoalKpiLink {
  goalId: string;
  kpiId: string;
}

export interface GoalCategory {
  id: string;
  /** Weight 0..100. All categories for an assignment must sum 100. */
  weight: number;
  name: string;
  description: string;
}

export interface Goal {
  id: string;
  categoryId: string;
  name: string;
  description: string;
  unit: GoalUnit;
  /** Weight 0..100. Goals within a category must sum 100. */
  weight: number;
  /** Numeric target value (interpreted via unit). */
  targetValue: number;
}

export interface ChangeRequest {
  id: string;
  assignmentId: string;
  goalId: string | null;       // null = request about a category
  categoryId: string | null;    // null = request about a goal-less item
  requesterId: string;         // manager profile id
  message: string;
  createdAt: string;           // ISO timestamp
}

export interface EmployeeAssignment {
  id: string;
  employeeId: string;          // simulated employee id
  categoryIds: string[];       // categories assigned to this employee
  goalIds: string[];           // goals assigned to this employee
}

/** Flat mock of org hierarchy: who is the manager of whom. */
export const MANAGER_MAP: Record<string, string> = {
  // profileId-as-employee -> manager profileId-as-employee
  'colaborador': 'jefe',
  'vendedor': 'gerente-tienda',
  'jefe': 'director',
  'gerente-tienda': 'divisional',
  'divisional': 'regional',
  'regional': 'director',
  'director': 'director-general',
  'rh': 'director',
};
```

Store exposes:

- Getters: `getCategories()`, `getGoals()`, `getKpis()`, `getAssignments()`, `getLinks()`, `getChangeRequests()`, `getAssignmentForEmployee(employeeId)`, `getCategoriesForAssignment(assignmentId)`, `getGoalsForCategory(categoryId)`, `getKpisForGoal(goalId)`, `getGoalsForKpi(kpiId)`, `getManagerOf(employeeId)`, `getCategoriesWeightSum(assignmentId)`, `getGoalsWeightSumForCategory(categoryId)`, `isAssignmentValid(assignmentId)`.
- Mutations: `addCategory(c)`, `updateCategory(id, u)`, `deleteCategory(id)` (cascada a metas), `addGoal(g)`, `updateGoal(id, u)`, `deleteGoal(id)`, `addKpi(k)`, `updateKpi(id, u)`, `deleteKpi(id)`, `linkKpiToGoal(goalId, kpiId)`, `unlinkKpiFromGoal(goalId, kpiId)`, `assignCategoryToEmployee(assignmentId, categoryId)`, `unassignCategoryFromEmployee(assignmentId, categoryId)`, `assignGoalToEmployee(assignmentId, goalId)`, `unassignGoalFromEmployee(assignmentId, goalId)`, `recordChangeRequest(req)`.

> **Note:** in this change, the store is seeded with a **fully populated assignment per profile**, so the employee does not pick from a pool — the categories and goals in their assignment are exactly those in the fixtures. `assignCategoryToEmployee` is wired but unused in UI; it exists for B3/C4 forward-compat.

## Data Model Constraints

```typescript
// Validation rules enforced in store getters / mutations

/** All categories assigned to one employee must sum to 100 (tolerance 0.01). */
function isCategoriesWeightValid(assignmentId: string): boolean {
  const cats = getCategoriesForAssignment(assignmentId);
  const sum = cats.reduce((acc, c) => acc + c.weight, 0);
  return Math.abs(sum - 100) < 0.01;
}

/** For each category, the goals assigned to it under the same assignment must sum to 100. */
function isCategoryGoalsWeightValid(categoryId: string, assignmentId: string): boolean {
  const goals = getGoalsForCategoryInAssignment(categoryId, assignmentId);
  const sum = goals.reduce((acc, g) => acc + g.weight, 0);
  return goals.length === 0 || Math.abs(sum - 100) < 0.01;
}
```

- A category with **0 goals** is allowed in the form but flagged as "vacía" (warning) until the user adds at least one goal — or until the user removes the category.
- A goal's `unit` and `Kpi.unit` are independent: a goal can use `porcentaje` while its KPI uses `moneda`. (We do not constrain this in A3; deferred to B3.)
- `targetValue` must be `> 0`. If `unit === 'porcentaje'`, the input caps visually at 100 but accepts decimals (e.g., 95.5).

## Component Architecture

```
+page.svelte (route /objetivos/asignacion)
├── dev persona + assignment lookup
├── mode = viewerIsManagerOf(devProfile, ownerId) ? 'reader' : 'editor'
├── ReadOnlyBanner (if reader)
├── WeightIndicator (categories global sum vs 100)
├── For each category in assignment:
│   └── CategoryCard
│       ├── WeightIndicator (goals-in-category sum vs 100)
│       ├── For each goal in category:
│       │   └── GoalRow
│       │       ├── KpiBadge[] (from getKpisForGoal)
│       │       ├── Edit/Delete actions (editor mode only)
│       │       └── "Solicitar cambio" button (reader mode only)
│       └── "Nueva meta" button (editor mode only)
├── "Nueva categoría" button (editor mode only)
└── Modals (rendered conditionally):
    ├── CategoryFormModal
    ├── GoalFormModal (contains KPI multi-select)
    ├── KpiFormModal (optional, for managing KPI library)
    ├── RequestChangeModal
    └── ConfirmDeleteModal (reuse A2)
```

## Validation Rules (double 100%)

| Rule | Where enforced | UX feedback |
|------|----------------|-------------|
| Suma de pesos de categorías = 100 ± 0.01 | `getCategoriesWeightSum(assignmentId)` | `WeightIndicator` global: verde si OK, ámbar con déficit/exceso y número exacto; botón "Guardar" deshabilitado |
| Suma de pesos de metas por categoría = 100 ± 0.01 (o categoría sin metas) | `isCategoryGoalsWeightValid` | `WeightIndicator` por `CategoryCard`: verde si OK, ámbar si no |
| Nombre de categoría único por asignación | `addCategory` / `updateCategory` | `alert-error` en modal |
| `targetValue > 0` | form submit | `alert-error` en modal |
| `weight 0..100` | form submit (input `min=0 max=100`) | `alert-error` en modal |
| Categoría con 0 metas | `isAssignmentValid` | Badge `badge-warning` "Sin metas"; no bloquea guardado (categoría vacía es válida transitoriamente) |

## Hierarchy Viewing Rules (decision #8)

The `MANAGER_MAP` mock defines who is the manager of whom. The page derives:

```
mode = (devProfile === ownerId) ? 'editor' : 'reader'
```

Editor mode (default for the active profile's own assignment):

- All categories/Goals/KPIs editable.
- Buttons: "Nueva categoría", "Editar", "Eliminar", "Guardar".

Reader mode (any non-owner, including other profile's assignments):

- All read-only inputs (inputs disabled, no edit/delete buttons).
- "Solicitar cambio" button per goal AND per category.
- Clicking opens `RequestChangeModal` with goal/category context prefilled (read-only) and a free-text `textarea`. On submit:
  1. Insert a `ChangeRequest` into the local `changeRequests` array.
  2. Show `alert-success` "Tu solicitud fue registrada".
  3. Close modal. (No email, no backend; mock local.)

> Edge case: a `rh` user opening another profile's assignment is treated as **reader** (RH administers catalog of competencies, not other people's goals in this phase). RH's own assignment is **editor** because RH also has goals (decision #7).

## KPI Linking Model

- A KPI is a **named indicator** (e.g., "Ingresos trimestrales", unit `$`, description "Suma de ingresos por trimestre"). It is independent of any goal.
- A Goal can link to **0..N KPIs** via `GoalKpiLink { goalId, kpiId }`.
- A KPI can be linked to **1..N Goals**.
- In the editor, the goal modal shows a list of checkboxes (KPI library). Selected KPIs become chips inside the goal row.
- In the reader, KPIs are shown as read-only `KpiBadge` chips.

> Note: a KPI used in a goal is purely a **semantic reference** in A3 — the goal's `targetValue` and `unit` are independent of the KPI's. The link only says "this goal contributes to this KPI". Wiring values (e.g., progress queries against the KPI) is deferred to A4 (mid-year) and C4 (API).

## Testing Strategy

| Layer | What to test | Approach |
|-------|--------------|----------|
| Unit | `goalsStore`: CRUD, cascade delete on category, link/unlink KPI, validation getters | Vitest with direct store function calls |
| Unit | Doble 100% validation: tolerances, edge cases (0 goals, 1 goal = 100, sum = 99.99) | Vitest unit tests on `isAssignmentValid` |
| Unit | `managerMap` resolution: editor vs reader for all 8 profiles | Vitest |
| Integration | Route renders editor for owner and reader for non-owner | Vitest + Svelte Testing Library |
| Integration | `WeightIndicator` colors switch on weight changes | Svelte Testing Library + `fireEvent` on form |
| E2E | Full flow: edit a goal weight → see badge turn amber → adjust → save | Playwright (deferred to verify phase) |

## Migration / Rollout

No migration. Greenfield UI module. Rollback: delete the 4 new fixture files, the store, the type file, the 11 component files, revert `+page.svelte` and `menuConfig.ts`. The placeholder page already exists, so rollback is safe.

## Open Questions

- [ ] ¿La validación de doble 100% aplica también a la categoría "sin metas" en A3, o se difiere a A4? (Decisión provisional: categoría sin metas es válida transitoriamente, con warning).
- [ ] `MANAGER_MAP` mock: ¿`director-general` debe aparecer como perfil de UI? (SPEC-ROADMAP l.62 lista 9 perfiles; `evaluation.ts` actual tiene 8 sin `director-general`. Seguir con 8 hasta B4 cierre.)
- [ ] ¿El modal de "Solicitar cambio" debe persistir entre recargas? (Decisión provisional: no, mock local. Persistencia real en change posterior de notificaciones / RBAC).
- [ ] ¿Una meta puede usar la misma unidad que su KPI? (Decisión provisional: no se restringe; cada entidad tiene su propia `unit`).
