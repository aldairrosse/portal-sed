# Tasks: Resultados de competencias — Paginación, búsqueda y filtro de equipo

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~550–600 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (backend: ~270) → PR 2 (store: ~120) → PR 3 (UI: ~180) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend: migración, endpoint, dual-write | PR 1 | base: `main`. Autónomo — verificable vía curl/BD. |
| 2 | Frontend store: `competencyResultsStore.svelte.ts` | PR 2 | base: PR 1 branch. `init()` + `load()` verificable con API corriendo. |
| 3 | Frontend UI: reescribir `+page.svelte` | PR 3 | base: PR 2 branch. Depende del store del PR 2. |

## Phase 1: Backend — Migración y schema

- [x] 1.1 Crear migración `000008_add_self_rh_rating_columns.up.sql`: agregar columnas `self_rating` y `rh_rating` (INTEGER NULL, CHECK 1–5) a `evaluation_competencies`; backfill desde `rating` para `self_rating`. Crear `.down.sql` para rollback.
- [x] 1.2 Actualizar `SubmitEval` en `api/internal/repository/evaluation/evaluation_repo.go`: dual-write al upsert de competencias — escribe `self_rating` cuando `setSelfCompleted`, `rh_rating` cuando `setRHCompleted`; mantiene `rating` para retrocompatibilidad.

## Phase 2: Backend — Repositorio

- [x] 2.1 Agregar `ListCompetencyResults(ctx, cycleID, query, currentUserID, offset, limit)` a `evaluation_repo.go`: SQL con JOINs a `evaluations`, `employees`, `evaluation_profiles`, `evaluation_competencies`; GROUP BY `e.id, ep.name`; AVG de `self_rating`/`rh_rating`; CASE para `status`; filtro ILIKE (`q`); filtro de exclusión del usuario actual; ORDER BY `last_name, first_name, e.id`; LIMIT/OFFSET.
- [x] 2.2 Agregar `CountCompetencyResults(ctx, cycleID, query, currentUserID)` a `evaluation_repo.go`: mismo filtrado pero `COUNT(DISTINCT e.id)` sin GROUP BY ni columnas de rating.

## Phase 3: Backend — Service, Handler, DTO y rutas

- [x] 3.1 Agregar `CompetencyResultItem`, `CompetencyResultsResponse` y `PaginationMeta` a `api/internal/dto/evaluation/evaluation_dto.go`.
- [x] 3.2 Agregar `GetCompetencyResults` a `api/internal/service/evaluation/evaluation_service.go`: clampa limit 1–200 (default 50), llama repo `List` + `Count`, computa `hasMore = offset + len(rows) < total`.
- [x] 3.3 Agregar `GetCompetencyResults` handler a `api/internal/handler/evaluation/evaluation_handler.go`: parsea `cycle_id` (UUID obligatorio), `offset` (≥0, default 0), `limit` (default 50, max 200), `q`, `scope` (`all`|`team`, default `all`); obtiene `currentUserID` de `auth.GetEmployeeID`; llama service y escribe JSON.
- [x] 3.4 Agregar `GetCompetencyResults` a la interfaz `EvalService` en `api/internal/handler/evaluation/interfaces.go`.
- [x] 3.5 Registrar `GET /evaluations/competency-results` en `api/internal/handler/evaluation/routes.go` dentro del grupo auth + readRateLimit + readReplica, **ANTES** de `GET /evaluations/{id}` para evitar conflicto de ruta en Chi.

## Phase 4: Backend — Tests

- [x] 4.1 Test unitario de `ListCompetencyResults` y `CountCompetencyResults` con mocks: cobertura de repo + service con mock repos; service tests verifican hasMore.
- [x] 4.2 Test unitario del handler: validar clamp de `limit`, rechazo de `scope` inválido, 400 si falta `cycle_id`, UUID inválido.
- [x] 4.3 Test unitario del service: mock repos, verificar cálculo de `hasMore` en borde (exacto, una página extra, última página).

## Phase 5: Frontend — Store

- [x] 5.1 Crear `web/src/lib/stores/competencyResultsStore.svelte.ts` con estructura idéntica a `rhEvaluadosStore.svelte.ts`: `$state` con `items`, `loading`, `error`, `currentPage`, `apiTotal`, `hasMore`, `hasPrev`, `currentQ`, `scopeFilter`. Exportar getters (`getItems`, `isLoading`, `getError`, `hasMoreItems`, `hasPrevItems`, `getCurrentPage`, `getTotalCount`, `getScope`). Implementar `init(cycleId)`, `load()` (GET con OFFSET=currentPage*50), `next()`, `prev()`, `search(q)` con debounce 300ms y reset de página, `setScope(s)`. `PAGE_SIZE = 50`.

## Phase 6: Frontend — UI

- [x] 6.1 Reescribir `web/src/routes/evaluacion/9x9/competencias/+page.svelte` siguiendo el patrón de `rh/evaluaciones/+page.svelte`: header con `BarChart3` + "Resultados de competencias", input de búsqueda **sin** `disabled={loading}`, toggle "Mi equipo" (`toggle toggle-sm` + tooltip), contador condicional, controles de paginación **siempre visibles**, skeleton solo en `{#if loading && items.length === 0}`, `ErrorState` con retry, estado vacío.
- [x] 6.2 Implementar tabla de 6 columnas: Empleado (`name`), Perfil (`titleCase(profileName)`), Autoevaluación (`selfRatingAvg?.toFixed(1) ?? '—'`), RH (`rhRatingAvg?.toFixed(1) ?? '—'`), Estado (badge condicional: `success`/`info`/`warning`/`ghost`), Acción (link a `/evaluacion/9x9/competencias/{id}`).
