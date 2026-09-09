# Tasks: Fix Competency Network Data Loading

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~180–220 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Backend — Repo, DTO, Service

- [x] 1.1 Add `GetCompetencyRatingsByEmployee` to `evaluation_repo.go`
  — Raw SQL: `SELECT ec.competency_id, ec.rating, ec.comments FROM evaluations e JOIN evaluation_competencies ec ON ec.evaluation_id = e.id WHERE e.employee_id = $1 AND e.cycle_id = $2 ORDER BY ec.competency_id`
  — Use `db.QueryContext` pattern (same as `ListByCycle`)
  — New struct `EmployeeCompetencyRatingRow { CompetencyID uuid.UUID; Rating int; Comments string }`
  — Return empty slice (not error) when no rows found
  — AC: table test with sqlmock covers populated and empty results

- [x] 1.2 Add `EmployeeCompetencyRatingsResponse` DTO and `EmployeeCompetencyRatingDTO` to `evaluation_dto.go`
  — Response: `{ EmployeeID uuid, CycleID uuid, Ratings []EmployeeCompetencyRatingDTO }`
  — Rating DTO: `{ CompetencyID uuid, Rating int, Comments string }`
  — JSON tags: `employeeId`, `cycleId`, `ratings`, `competencyId`, `rating`, `comments`
  — AC: fields match design doc response shape

- [x] 1.3 Add `GetEmployeeCompetencyRatings(ctx, employeeID, cycleID) → (*EmployeeCompetencyRatingsResponse, error)` to `evaluation_service.go`
  — Call repo query, map rows to DTO ratings slice
  — Return `EVALUATION_NOT_FOUND` domain error when no evaluation exists for employee+cycle
  — No phase validation (read-only endpoint)
  — AC: mock repo test verifies mapping and not-found → error

## Phase 2: Backend — Handler & Route

- [x] 2.1 Add `GetEmployeeCompetencies` handler method to `evaluation_handler.go`
  — Parse `{employeeId}` from Chi URL param + `cycle_id` from query string
  — Validate both are valid UUIDs; `cycle_id` missing → 400 `MISSING_PARAM`
  — Call `evalSvc.GetEmployeeCompetencyRatings(ctx, employeeID, cycleID)`
  — Return 200 with response DTO or error status
  — AC: httptest: missing cycle_id → 400; valid → 200; invalid UUID → 400

- [x] 2.2 Register `GET /evaluations/employee/{employeeId}` in `routes.go`
  — Inside `RequireAuth` group, before `GET /evaluations/{id}`
  — Reuse `readRateLimit` + `readReplicaMiddleware` pattern (same as other GETs)
  — For now: no RBAC middleware (proposal scopes out new permissions)
  — AC: route responds at `GET /api/v1/evaluations/employee/{uuid}`

## Phase 3: OpenAPI Spec

- [x] 3.1 Add endpoint schema to `api/openapi/evaluations-and-9x9.yaml`
  — Path: `/evaluations/employee/{employeeId}` with `GET` operation
  — Query param: `cycle_id` (required, UUID)
  — Response 200: `EmployeeCompetencyRatingsResponse` schema (employeeId, cycleId, ratings[])
  — Response 400: reuse `BadRequest`
  — Response 404: reuse `NotFound`
  — AC: spec validates with `openspec validate`

## Phase 4: Frontend — Store

- [x] 4.1 Modify `evaluationStore.load()` to accept optional `employeeId` param
  — Signature: `export async function load(employeeId?: string): Promise<void>`
  — When `employeeId` provided:
    - Get `cycleId` from `getActiveCycle()?.id`; throw if null
    - Call `client.GET('/evaluations/employee/{employeeId}', { params: { query: { cycle_id: cycleId } } })`
    - Map `response.ratings[]` → `CompetencyRating[]` with `selfRating: r.rating` (single rating column, view uses `selfRating`)
  — When `employeeId` absent: keep existing path (GET `/evaluations/{id}` with session user)
  — AC: Vitest: mock `client.GET`, verify correct URL + cycle_id when employeeId provided; verify existing path unchanged when absent

## Phase 5: Frontend — Route & Skeleton

- [x] 5.1 Modify `+page.svelte` to call `evaluationStore.load(employeeId)` on mount
  — Import `load, isLoading` from `$lib/stores/evaluationStore.svelte`
  — Add `$effect` block: `if (employeeId) load(employeeId)`
  — Show `<PageSkeleton variant="card" rows={3} avatar />` while `isLoading()` is true and no data
  — Show `CompetencyNetworkView` when loading completes
  — AC: page triggers fetch on mount; skeleton visible during loading; view renders on completion
  — AC: existing pages NOT broken (no-arg `load()` still works)
