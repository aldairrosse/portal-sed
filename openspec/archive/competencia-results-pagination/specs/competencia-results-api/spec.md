# competencia-results-api Specification

## Purpose

Endpoint paginado server-side que retorna los resultados agregados de competencias por empleado dentro de un ciclo, con búsqueda y filtro de equipo. Reemplaza la carga completa en cliente por queries con `OFFSET`/`LIMIT`.

## API Contract

| Field | Type | Details |
|-------|------|---------|
| Method | `GET` | |
| Path | `/api/v1/evaluations/competency-results` | |
| Query `cycle_id` | string (UUID) | Requerido; identifica el ciclo de evaluación |
| Query `offset` | int | Default 0, clamped ≥ 0 |
| Query `limit` | int | 1–200, default 50 |
| Query `q` | string | ILIKE search sobre nombre de empleado; opcional |
| Query `scope` | string | `all` (default) o `team` (excluye al usuario actual) |
| Response `200` | `CompetencyResultsResponse` | `{ data: CompetencyResultItem[], meta: { hasMore, limit, offset, total } }` |
| Response `400` | `APIErrorResponse` | `cycle_id` faltante, UUID inválido, limit fuera de rango |
| Response `404` | `APIErrorResponse` | Ciclo no encontrado |

`CompetencyResultItem`:

| Campo | Tipo | Origen BD |
|-------|------|-----------|
| `id` | string (UUID) | `employees.id` |
| `name` | string | `employees.first_name \|\| ' ' \|\| employees.last_name` |
| `profileName` | string | `evaluation_profiles.name` |
| `selfRatingAvg` | number \| null | `AVG(ec.self_rating)` sobre `evaluation_competencies` JOIN `evaluations` |
| `rhRatingAvg` | number \| null | `AVG(ec.rh_rating)` |
| `status` | string | Derivado: `completada` \| `autoevaluacion` \| `pendiente` \| `sin-datos` |

## Requirements

### Requirement: Paginated competency results with search and team filter

El endpoint SHALL retornar resultados agregados de competencias para empleados en el ciclo dado, paginados con `OFFSET`/`LIMIT`. SHALL soportar búsqueda por nombre (`q`) y filtro de equipo (`scope=team`). Los promedios SHALL calcularse en BD con `AVG()` sobre `evaluation_competencies` filtradas por `evaluations.cycle_id`.

#### Scenario: Default page load with all scope

- GIVEN un ciclo con 120 empleados evaluados en competencias
- WHEN `GET /evaluations/competency-results?cycle_id={id}` (sin offset, limit, q, scope)
- THEN response tiene `meta.limit:50`, `meta.offset:0`, `meta.total:120`, `meta.hasMore:true`
- AND `data` contiene 50 items con `selfRatingAvg`, `rhRatingAvg`, `profileName` poblados

#### Scenario: Team scope excludes current user

- GIVEN usuario actual `emp-001` con 5 miembros en su equipo y `scope=team`
- WHEN endpoint es llamado
- THEN `data` no contiene ningún item con `id=emp-001`
- AND `meta.total` refleja el conteo sin el usuario actual

#### Scenario: Search filters by employee name

- GIVEN ciclo con empleados "María Gómez" y "Carlos Pérez"
- WHEN `GET ...?q=María`
- THEN `data` contiene solo a María
- AND `meta.total` refleja el conteo filtrado

#### Scenario: Last page sets hasMore false

- GIVEN 95 empleados, `offset=50`, `limit=50`
- WHEN endpoint es llamado
- THEN `meta.hasMore` es `false`, `meta.total` es 95

#### Scenario: Invalid limit returns 400

- GIVEN `limit=300`
- WHEN endpoint es llamado
- THEN response es 400 con mensaje "limit must be between 1 and 200"

#### Scenario: Missing cycle_id returns 400

- GIVEN request sin `cycle_id`
- WHEN endpoint es llamado
- THEN response es 400 con mensaje indicando que `cycle_id` es requerido

### Requirement: Repository layer with separate COUNT method

`EvaluationRepo` SHALL proveer `ListCompetencyResults(ctx, cycleID, query, scope, currentUserID, offset, limit)` con JOIN a `evaluation_competencies`, `evaluations`, `employees`, `evaluation_profiles`; y `CountCompetencyResults(ctx, cycleID, query, scope, currentUserID)` para el total.

**Files**: `api/internal/repository/evaluation/evaluation_repo.go`

### Requirement: Service layer hasMore computation

`EvaluationService` SHALL proveer `GetCompetencyResults(ctx, cycleID, query, scope, currentUserID, offset, limit)` que llama al repositorio y computa `hasMore = offset + len(rows) < total`.

**Files**: `api/internal/service/evaluation/evaluation_service.go`

### Requirement: Handler query param parsing

`EvaluationHandler.GetCompetencyResults` SHALL parsear `cycle_id`, `offset`, `limit`, `q`, `scope` del query string. SHALL validar `cycle_id` como UUID obligatorio. SHALL clampear `limit` a 1–200 y `offset` a ≥ 0. `scope` SHALL aceptar solo `all` o `team`, default `all`.

**Files**: `api/internal/handler/evaluation/evaluation_handler.go`
