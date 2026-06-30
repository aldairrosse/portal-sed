# Design: Resultados de competencias — Paginación, búsqueda y filtro de equipo

## Technical Approach

Mover el cálculo de promedios de competencias del cliente al servidor, creando un endpoint paginado `GET /api/v1/evaluations/competency-results` con OFFSET/LIMIT, búsqueda ILIKE y filtro de equipo. El frontend reutiliza el patrón de `rhEvaluadosStore` + `rh/evaluaciones/+page.svelte` para la UI paginada.

**Prerequisito crítico**: La columna `rating` en `evaluation_competencies` es única (se sobreescribe por RH). Se requieren columnas `self_rating` y `rh_rating` para promedios independientes. Ver sección de Riesgos.

## Architecture Decisions

| Decisión | Opción | Tradeoff | Elección |
|----------|--------|----------|----------|
| Paginación | OFFSET/LIMIT vs cursor | OFFSET permite saltar a página N; cursor es mejor para datasets muy volátiles. Los resultados de competencias son snapshot-style. | **OFFSET/LIMIT** — consistente con `rhEvaluadosStore` y `ListByManagerPaginated` |
| Cálculo de promedios | AVG() en SQL vs agregar en app | SQL AVG es O(1) en memoria, aprovecha índice. App requiere traer todos los rows. | **AVG() en SQL** — GROUP BY employee_id |
| Scope team | Filtrar en SQL con subquery org_nodes vs resolver en service | SQL es más eficiente pero requiere JOIN a org hierarchy. Service ya resuelve team en `evaluatee_service`. | **SQL con JOIN a org hierarchy** — un solo query, sin N+1 |
| Ubicación del repo | `EvaluationRepo` existente vs nuevo `CompetencyResultsRepo` | EvaluationRepo ya tiene `db` y el patrón raw SQL. Crear nuevo repo es over-engineering. | **EvaluationRepo** — agregar métodos al repo existente |
| Tipo de respuesta | DTO con `data` + `meta` vs array plano | `data`+`meta` es el patrón establecido en `EmployeeListResponse`. | **`data` + `meta`** — consistente con el resto de la API |

## Data Flow

```
+page.svelte
    │ onMount → init(cycleId) → load()
    ▼
competencyResultsStore.svelte.ts
    │ GET /api/v1/evaluations/competency-results
    │   ?cycle_id=&offset=&limit=&q=&scope=
    ▼
EvaluationHandler.GetCompetencyResults
    │ parsea query params, valida cycle_id UUID
    │ obtiene currentUserID de auth context
    ▼
EvaluationService.GetCompetencyResults
    │ llama repo.List + repo.Count
    │ computa hasMore = offset + len(rows) < total
    ▼
EvaluationRepo (raw SQL)
    │ SELECT e.id, concat(e.first_name,' ',e.last_name),
    │        ep.name, AVG(ec.self_rating), AVG(ec.rh_rating),
    │        CASE WHEN ... END as status
    │ FROM evaluations ev
    │ JOIN employees e ON e.id = ev.employee_id
    │ JOIN evaluation_profiles ep ON ep.id = e.profile_id
    │ LEFT JOIN evaluation_competencies ec ON ec.evaluation_id = ev.id
    │ WHERE ev.cycle_id = $1
    │   [AND scope=team → exclude current user]
    │   [AND q → ILIKE filter]
    │ GROUP BY e.id, ep.name
    │ ORDER BY e.last_name, e.first_name
    │ LIMIT $x OFFSET $y
    ▼
PostgreSQL → rows → DTO → JSON → client → store → UI
```

## Backend Design

### Migración: `000008_add_self_rh_rating_columns`

```sql
-- Up
ALTER TABLE evaluation_competencies ADD COLUMN self_rating INTEGER NULL CHECK (self_rating >= 1 AND self_rating <= 5);
ALTER TABLE evaluation_competencies ADD COLUMN rh_rating INTEGER NULL CHECK (rh_rating >= 1 AND rh_rating <= 5);
-- Backfill: existing rating → self_rating (asumiendo que rating actual es autoevaluación si no hay rh_evaluation_completed_at)
UPDATE evaluation_competencies ec
SET self_rating = ec.rating
FROM evaluations ev
WHERE ec.evaluation_id = ev.id AND ev.self_evaluation_completed_at IS NOT NULL;
-- La columna rating se mantiene para retrocompatibilidad durante la transición.

-- Down
ALTER TABLE evaluation_competencies DROP COLUMN IF EXISTS self_rating;
ALTER TABLE evaluation_competencies DROP COLUMN IF EXISTS rh_rating;
```

### Repository: `evaluation_repo.go`

Agregar a `EvaluationRepo`:

```go
type CompetencyResultRow struct {
    ID            uuid.UUID
    Name          string
    ProfileName   string
    SelfRatingAvg *float64  // NULL si sin self_ratings
    RHRatingAvg   *float64  // NULL si sin rh_ratings
    Status        string    // completada | autoevaluacion | pendiente | sin-datos
}

func (r *EvaluationRepo) ListCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, scope string, currentUserID uuid.UUID, offset, limit int) ([]*CompetencyResultRow, error)

func (r *EvaluationRepo) CountCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, scope string, currentUserID uuid.UUID) (int, error)
```

**SQL base (List)**:
```sql
SELECT e.id,
       e.first_name || ' ' || e.last_name AS name,
       COALESCE(ep.name, '') AS profile_name,
       AVG(ec.self_rating) AS self_rating_avg,
       AVG(ec.rh_rating) AS rh_rating_avg,
       CASE
         WHEN AVG(ec.self_rating) IS NOT NULL AND AVG(ec.rh_rating) IS NOT NULL THEN 'completada'
         WHEN AVG(ec.self_rating) IS NOT NULL THEN 'autoevaluacion'
         WHEN AVG(ec.rh_rating) IS NOT NULL THEN 'pendiente'
         ELSE 'sin-datos'
       END AS status
FROM evaluations ev
JOIN employees e ON e.id = ev.employee_id
LEFT JOIN evaluation_profiles ep ON ep.id = e.profile_id
LEFT JOIN evaluation_competencies ec ON ec.evaluation_id = ev.id
WHERE ev.cycle_id = $1
```

**Filtros dinámicos**:
- `scope=team`: `AND e.id = ANY($teamIDs) AND e.id != $currentUserUUID` — teamIDs se resuelve en service vía `evaluatee_service.GetTeamMembers` lógica (ListByHeadEmployee → ListByOrgNodeIDs → ListChildren heads)
- `q != ""`: `AND (e.first_name ILIKE $q OR e.last_name ILIKE $q)`
- ORDER BY: `e.last_name, e.first_name, e.id`
- LIMIT/OFFSET al final

**Count**: Mismo query sin SELECT de columnas de rating, sin GROUP BY, solo `COUNT(DISTINCT e.id)`.

### Service: `evaluation_service.go`

```go
func (s *EvaluationService) GetCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, scope string, currentUserID uuid.UUID, offset, limit int) (*dto.CompetencyResultsResponse, error)
```

Clampa `limit` 1–200 (default 50), `offset` >= 0. Llama `ListCompetencyResults` + `CountCompetencyResults`. Computa `hasMore = offset + len(rows) < total`.

### Handler: `evaluation_handler.go`

`GetCompetencyResults(w, r)`:
1. Parsea `cycle_id` (requerido, UUID)
2. Parsea `offset` (default 0), `limit` (default 50, clamp 1–200)
3. Parsea `q` (opcional), `scope` (default "all", valida "all"|"team")
4. Obtiene `currentUserID` de `auth.GetEmployeeID(r.Context())`
5. Llama service, escribe JSON

### Interfaces: `interfaces.go`

Agregar a `EvalService`:
```go
GetCompetencyResults(ctx context.Context, cycleID uuid.UUID, query string, scope string, currentUserID uuid.UUID, offset, limit int) (*dto.CompetencyResultsResponse, error)
```

### DTO: `evaluation_dto.go`

```go
type CompetencyResultItem struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`
    ProfileName string   `json:"profileName"`
    SelfRatingAvg *float64 `json:"selfRatingAvg"`
    RHRatingAvg   *float64 `json:"rhRatingAvg"`
    Status      string   `json:"status"`
}

type CompetencyResultsResponse struct {
    Data []CompetencyResultItem `json:"data"`
    Meta PaginationMeta         `json:"meta"`
}

type PaginationMeta struct {
    HasMore bool `json:"hasMore"`
    Total   int  `json:"total"`
    Offset  int  `json:"offset"`
    Limit   int  `json:"limit"`
}
```

### Routes: `routes.go`

Registrar dentro del grupo auth + readRateLimit + readReplica:
```go
r.Get("/evaluations/competency-results", handler.GetCompetencyResults)
```

**Importante**: Debe registrarse ANTES de `/evaluations/{id}` para que Chi no interprete "competency-results" como un `{id}`.

## Frontend Store Design

**File**: `web/src/lib/stores/competencyResultsStore.svelte.ts`

Estructura idéntica a `rhEvaluadosStore.svelte.ts`:

```ts
// State
let items = $state<CompetencyResultItem[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);
let currentPage = $state(0);
let apiTotal = $state(0);
let hasMore = $state(false);
let hasPrev = $state(false);
let currentQ: string | undefined;
let scopeFilter = $state<'all' | 'team'>('all');
let cycleId: string | undefined;

const PAGE_SIZE = 50;

// Getters (mismo naming que rhEvaluadosStore)
export function getItems(), isLoading(), getError(), hasMoreItems(),
       hasPrevItems(), getCurrentPage(), getTotalCount(), getSearchQuery()

// Acciones
export async function init(id: string)     // guarda cycleId
export async function load()               // GET con offset=currentPage*50
export async function next()               // currentPage++ si hasMore
export async function prev()               // currentPage-- si currentPage > 0
export function search(query: string)      // debounce 300ms, reset page
export function setScope(s: 'all'|'team')  // cambia scopeFilter, reset page, load()
export function getScope(): 'all' | 'team' // getter para el toggle
```

**API call**: `client.GET('/evaluations/competency-results' as never, { params: { query: { cycle_id: cycleId, offset, limit: 50, q: currentQ, scope: scopeFilter } } })`

## Frontend UI Design

**File**: `web/src/routes/evaluacion/9x9/competencias/+page.svelte`

Reescritura completa siguiendo `rh/evaluaciones/+page.svelte`:

```
┌─────────────────────────────────────────────────────────────┐
│ ★ Resultados de competencias                                │
│   Promedios de autoevaluación y evaluación RH...           │
├─────────────────────────────────────────────────────────────┤
│ [Buscar empleado...]  ☐ Mi equipo  Viendo X de Y  [◀ Pág N ▶]│
├─────────────────────────────────────────────────────────────┤
│ Empleado │ Perfil │ Autoeval │ RH │ Estado │ Acción         │
│ ─────────┼────────┼──────────┼────┼────────┼────────        │
│ J. Doe   │ Ger... │   3.5    │4.0 │ ● Comp │ Ver →          │
│ ...      │        │          │    │        │                │
└─────────────────────────────────────────────────────────────┘
```

- **Header**: `BarChart3` icon + "Resultados de competencias" + descripción
- **Search**: `input.input-bordered.input-sm.max-w-xs`, SIN `disabled={loading}`
- **Toggle "Mi equipo"**: `input.toggle.toggle-sm` + label + tooltip "Excluye tu propio registro de la lista"
- **Counter**: "Viendo {items.length} de {total} empleados" / "Viendo {items.length} resultado(s)" si hay búsqueda
- **Paginación**: `btn-outline btn-xs` + `ChevronLeft`/`ChevronRight` + "Pág. {currentPage+1}" — siempre visibles
- **Skeleton**: SOLO `{#if loading && items.length === 0}` → `PageSkeleton variant="table" rows={5}`
- **Error**: `ErrorState` con retry → `load()`
- **Empty**: "Sin resultados de competencias para mostrar"
- **Tabla**: 6 columnas (Empleado, Perfil, Autoevaluación, RH, Estado, Acción)
  - Perfil: `titleCase(item.profileName)`
  - Autoevaluación: `item.selfRatingAvg?.toFixed(1) ?? '—'`
  - RH: `item.rhRatingAvg?.toFixed(1) ?? '—'`
  - Estado: badge con clase condicional (success/info/warning/ghost)
  - Acción: `<a href="/evaluacion/9x9/competencias/{item.id}">Ver detalle</a>`
- **onMount**: `init(getActiveCycle()?.id)` → `load()`

## Team Filter Design

Cuando `scope=team`:
- Backend obtiene `currentUserID` de `auth.GetEmployeeID(r.Context())`
- Service resuelve team IDs replicando la lógica de `evaluatee_service.GetTeamMembers`:
  1. `nodeRepo.ListByHeadEmployee(ctx, currentUserID)` → nodos donde el usuario es head
  2. `empRepo.ListByOrgNodeIDs(ctx, nodeIDs, true)` → empleados directos del nodo
  3. `nodeRepo.ListChildren(ctx, nodeID)` para cada nodo → head employees de hijos directos
  4. Merge + dedup
- SQL agrega: `AND e.id = ANY($teamIDs) AND e.id != $currentUserUUID`
- **Perfil `rh`/`director-general`**: ven todos (scope=all por defecto). El toggle "Mi equipo" filtra por su equipo orgánico.
- **Perfil `jefe`**: solo ven su equipo (nodos donde es head + heads de hijos)

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/migrations/000008_add_self_rh_rating_columns.up.sql` | Create | Migración: agrega `self_rating` y `rh_rating` a `evaluation_competencies` |
| `api/migrations/000008_add_self_rh_rating_columns.down.sql` | Create | Rollback de la migración |
| `api/internal/repository/evaluation/evaluation_repo.go` | Modify | Agregar `ListCompetencyResults` + `CountCompetencyResults` con SQL raw |
| `api/internal/service/evaluation/evaluation_service.go` | Modify | Agregar `GetCompetencyResults`, importar `evaluateeService` lógica para team IDs |
| `api/internal/handler/evaluation/evaluation_handler.go` | Modify | Agregar `GetCompetencyResults` handler + actualizar `SubmitEval` para dual-write |
| `api/internal/handler/evaluation/interfaces.go` | Modify | Agregar método a `EvalService` interface |
| `api/internal/handler/evaluation/routes.go` | Modify | Registrar `GET /evaluations/competency-results` |
| `api/internal/dto/evaluation/evaluation_dto.go` | Modify | Agregar `CompetencyResultItem`, `CompetencyResultsResponse`, `PaginationMeta` |
| `web/src/lib/stores/competencyResultsStore.svelte.ts` | Create | Store reactivo con paginación, búsqueda, scope |
| `web/src/routes/evaluacion/9x9/competencias/+page.svelte` | Modify | Reescritura completa con paginación server-side |

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | `ListCompetencyResults` / `CountCompetencyResults` | Table-driven tests con `sqlmock` — verificar SQL generado, filtros, paginación |
| Unit (Go) | `GetCompetencyResults` service | Mock repo, verificar `hasMore` computation |
| Unit (Go) | Handler param parsing | Test `limit` clamp, `scope` validation, `cycle_id` required |
| Unit (TS) | Store: `load`, `next`, `prev`, `search`, `setScope` | Viest + mock `client.GET` |
| Integration | Endpoint completo con BD test | Verificar AVG, COUNT, ILIKE, scope exclusion |
| E2E | N/A para este change | Fuera de scope |

## Risks and Mitigations

| Riesgo | Impacto | Probabilidad | Mitigación |
|--------|---------|-------------|------------|
| **Columnas `self_rating`/`rh_rating` no existen en el schema actual** | Alto | Cierta | Migración `000008` como prerequisito. Backfill desde `rating` existente. Sin esto, los promedios separados no son posibles. |
| `AVG()` sobre muchos rows por empleado puede ser lento | Medio | Media | Índice en `evaluation_competencies(evaluation_id)` ya existe. El JOIN a `evaluations` filtra por `cycle_id` indexado. |
| Route conflict: `/evaluations/competency-results` vs `/evaluations/{id}` | Alto | Media | Chi matchea rutas estáticas antes que params. Registrar `competency-results` ANTES de `{id}`. Verificar con test. |
| Migración incluida + dual-write en SubmitEval | Medio | Alta | Se implementa dual-write en este change. `SubmitEval` escribe en `rating` (legacy) + `self_rating`/`rh_rating` (nuevos). Backfill para datos existentes. |
| `evaluation_competencies.rating` se sigue sobreescribiendo | Medio | Alta | Con dual-write en este change, `self_rating` y `rh_rating` se escriben correctamente desde el inicio. La columna `rating` se mantiene para retrocompatibilidad. |

## Decisions (resolved)

1. **Migración incluida en este change**: La migración `000008` + backfill + dual-write en `SubmitEval` se implementa aquí. No se crea change separado.
2. **Filtro de equipo por jerarquía**: `scope=team` usa la misma lógica que `GetTeamMembers` en `evaluatee_service.go`:
   - Empleados en los nodos org donde el usuario es `head_employee` (direct team)
   - Head employees de nodos hijos directos (indirect reports)
   - Excluir al usuario actual del listado
   - Implementación: el backend resuelve team IDs vía `nodeRepo.ListByHeadEmployee` + `empRepo.ListByOrgNodeIDs` + `nodeRepo.ListChildren`, luego filtra `WHERE e.id IN (team_ids) AND e.id != $currentUserID`
