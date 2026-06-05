# C6: evaluations-and-9x9-api — Implementation Tasks

## Summary Table

| Phase | Tasks | Est. Hours | Deliverables |
|-------|-------|-----------|-------------|
| Phase 1: Infrastructure | T1–T3 | ~4h | Package structure, pure helpers |
| Phase 2: Repository Layer | T4–T10 | ~12h | Ent repositories with tx patterns |
| Phase 3: Service Layer | T11–T19 | ~16h | Business logic, concurrency, queue |
| Phase 4: Handler Layer | T20–T23 | ~8h | Chi handlers, route wiring |
| Phase 5: OpenAPI | T24–T25 | ~3h | Contract + TS types |
| Phase 6: Tests | T26–T34 | ~12h | Unit, integration, load, concurrency |
| Phase 7: Polish | T35–T36 | ~2h | Docs + metrics |
| **Total** | **T1–T36** | **~57h** | **Complete year-end closing API** |

---

## Phase 1: Infrastructure (~4h)

### T1: Package Scaffolding
**Layer:** Infrastructure  
**Estimated Time:** 1.5h  
**Dependencies:** None

Create the package skeleton for C6 under `api/internal/`.

- [x] Create directories: `handler/evaluation/`, `service/evaluation/`, `repository/evaluation/`, `pkg/quadrant/`, `pkg/state/`
- [x] Add `.gitkeep` or empty `doc.go` where needed to commit directories
- [x] Add `go.mod` workspace references if required
- [x] Ensure build passes with `go build ./...`

---

### T2: Quadrant Computation Helper (`pkg/quadrant/compute.go`)
**Layer:** Infrastructure  
**Estimated Time:** 1h  
**Dependencies:** T1

Implement the pure, deterministic quadrant calculator.
- [x] Implement `ComputeQuadrant(performance, potential int) int` with tiering logic
- [x] Implement `tier(score int) int` helper (1–3, 4–6, 7–9)
- [x] Add package-level documentation explaining the 3×3 grid mapping
- [x] Add `go:generate` or build tag if needed for test harness
- [x] `go test` passes (even if no tests yet)

---

### T3: State Machine Helper (`pkg/state/machine.go`)

**Layer:** Infrastructure  
**Estimated Time:** 1.5h  
**Dependencies:** T1

Implement a lightweight state transition guard for `Evaluation` states.

- [x] Define valid transitions: `pendiente_evaluacion_final` → `en_progreso`, `en_progreso` → `completada`
- [x] Implement `CanTransition(from, to string) bool`
- [x] Implement `IsTerminal(state string) bool` (only `completada` is terminal)
- [x] Add guard `RequiresPhase(state, phase string) error` that returns `INVALID_PHASE` when `phase != cierre`
- [x] `go test` passes

---

## Phase 2: Repository Layer (~12h)

### T4: Evaluation Repository (CRUD + State Transitions)
**Layer:** Repository  
**Estimated Time:** 2h  
**Dependencies:** T1, C1 (data-model-core applied)

Implement `EvaluationRepo` interface for evaluation lifecycle.

- [x] `GetByID(ctx, id)` returns `*ent.Evaluation` with `NotFound` mapped
- [x] `ListByCycle(ctx, cycleID, state, cursor, limit)` returns cursor-paginated list
- [x] `GetDetail(ctx, id)` returns evaluation with `CompetencyRatings` and `GoalRatings` preloaded
- [x] `SubmitCompetenciesAndGoals(ctx, tx, evalID, comps, goals)` uses `SELECT FOR UPDATE` then bulk upsert
- [x] `Finalize(ctx, evalID)` sets state `completada` and increments `version` inside tx
- [x] `GetSummaryByCycle(ctx, cycleID)` queries materialized view `evaluation_summary`
- [x] All methods return domain-specific errors (`EVALUATION_NOT_FOUND`, etc.)

---

### T5: EvaluationCompetency Repository (Bulk Upsert Ratings)
**Layer:** Repository  
**Estimated Time:** 1.5h  
**Dependencies:** T4

Bulk competency rating operations extracted for reusability.

- [ ] `BulkUpsert(ctx, tx, evalID, []CompetencyUpsert)` uses `CreateBulk` + `OnConflict`
- [ ] `DeleteByEvaluation(ctx, tx, evalID)` clears all competency ratings for an evaluation
- [ ] `GetByEvaluation(ctx, evalID)` returns all ratings for a given evaluation
- [ ] Unit test: 5 competencies upserted in one round-trip

---

### T6: EvaluationGoal Repository (Link Goals to Evaluation)
**Layer:** Repository  
**Estimated Time:** 1.5h  
**Dependencies:** T4, C4 (goals-api applied)

Goal comment updates during evaluation closing.

- [ ] `UpdateComments(ctx, tx, evalID, []GoalCommentUpsert)` updates `EvaluationGoal.final_comments`
- [ ] `GetByEvaluation(ctx, evalID)` returns goals linked to evaluation
- [ ] `VerifyGoalsExist(ctx, evalID, goalIDs)` ensures all referenced goals belong to the evaluation
- [ ] Returns `GOAL_NOT_FOUND` if any goal ID is invalid for this evaluation

---

### T7: NineBoxMatrix Repository (CRUD)
**Layer:** Repository  
**Estimated Time:** 1.5h  
**Dependencies:** T1, C1

Matrix CRUD with evaluator ownership.

- [ ] `CreateMatrix(ctx, cycleID, evaluatorID)` creates matrix and returns it
- [ ] `GetMatrixByID(ctx, id)` returns matrix with entries preloaded
- [ ] `ListMatrices(ctx, cycleID, evaluatorID)` filters by cycle and/or evaluator
- [ ] `DeleteMatrix(ctx, id)` soft-deletes or hard-deletes based on spec (hard for now)
- [ ] Returns `MATRIX_NOT_FOUND` when not found

---

### T8: NineBoxEntry Repository (CRUD + Upsert)
**Layer:** Repository  
**Estimated Time:** 2h  
**Dependencies:** T7

Entry upsert with locking and version support.

- [ ] `UpsertEntry(ctx, tx, matrixID, evaluateeID, perf, pot, quadrant, comments)` uses `OnConflict` update
- [ ] `UpdateEntry(ctx, tx, entryID, perf, pot, quadrant, comments, version)` checks optimistic lock
- [ ] `BatchUpsertEntries(ctx, tx, matrixID, []EntryUpsert)` locks all existing entries with `SELECT FOR UPDATE` then bulk upserts
- [ ] `GetMatrixEntries(ctx, matrixID)` returns entries for a matrix
- [ ] Returns `ENTRY_NOT_FOUND` / `CONCURRENT_UPDATE` appropriately

---

### T9: NineBoxQuadrant Repository (Read-Only Catalog)
**Layer:** Repository  
**Estimated Time:** 1h  
**Dependencies:** T1, C1

Read-only catalog for quadrant metadata.

- [ ] `GetQuadrants(ctx)` returns all 9 quadrant definitions
- [ ] `GetQuadrantByNumber(ctx, quadrant)` returns single quadrant metadata
- [ ] `GetQuadrantByScores(ctx, perf, pot)` maps scores to quadrant via `ComputeQuadrant` then fetches metadata
- [ ] Results cached in memory for 1h (simple map, no external cache needed)

---

### T10: NineBoxScale Repository (Read-Only Catalog)
**Layer:** Repository  
**Estimated Time:** 1h  
**Dependencies:** T1, C1

Read-only catalog for scale definitions.

- [ ] `GetScales(ctx)` returns all 18 scale rows (9 performance + 9 potential)
- [ ] `GetScalesByAxis(ctx, axis)` filters by `performance` or `potential`
- [ ] `GetScaleByAxisAndLevel(ctx, axis, level)` returns single scale definition
- [ ] Results cached in memory for 1h

---

## Phase 3: Service Layer (~16h)

### T11: Evaluation Service — Create/Get/List Evaluations
**Layer:** Service  
**Estimated Time:** 1.5h  
**Dependencies:** T4

Basic read operations and evaluation creation.

- [ ] `GetEvaluation(ctx, id)` returns `EvaluationDetailResponse` with competencies and goals
- [ ] `ListEvaluations(ctx, cycleID, state, cursor, limit)` returns `EvaluationListResponse` with cursor pagination
- [ ] `CreateEvaluation(ctx, employeeID, cycleID)` creates a new evaluation in `pendiente_evaluacion_final` state
- [ ] All read endpoints route to read replicas when replica driver is configured
- [ ] Returns proper error mapping for missing resources

---

### T12: Evaluation Service — SubmitSelfEvaluation (with Phase Check)
**Layer:** Service  
**Estimated Time:** 2h  
**Dependencies:** T11, T3, T5, T6, C2

Self-evaluation submission with full guardrails.

- [ ] `SubmitSelfEvaluation(ctx, evalID, req, idempotencyKey)` checks Redis idempotency first
- [ ] Validates cycle phase is `cierre` via `CycleService.GetPhase` (C2 dependency)
- [ ] Checks `self_evaluation_deadline` has not passed
- [ ] Acquires `SELECT FOR UPDATE` on evaluation row
- [ ] Rejects if state is `completada` (`EVALUATION_ALREADY_FINALIZED`)
- [ ] Bulk upserts `EvaluationCompetency` ratings and `EvaluationGoal` comments in one tx
- [ ] Sets `self_evaluation_completed_at` and increments `version`
- [ ] Stores result in Redis under idempotency key with 24h TTL
- [ ] Returns `EvaluationDetailResponse`

---

### T13: Evaluation Service — SubmitRHEvaluation (with Phase Check)
**Layer:** Service  
**Estimated Time:** 2h  
**Dependencies:** T12

RH evaluation submission parallel to self-evaluation.

- [ ] `SubmitRHEvaluation(ctx, evalID, req, idempotencyKey)` checks idempotency
- [ ] Validates cycle phase is `cierre` (no deadline check)
- [ ] Acquires `SELECT FOR UPDATE` on evaluation row
- [ ] Rejects if state is `completada`
- [ ] Bulk upserts `EvaluationCompetency` ratings in one tx
- [ ] Sets `rh_evaluation_completed_at` and increments `version`
- [ ] Stores result in Redis idempotency cache
- [ ] Returns `EvaluationDetailResponse`

---

### T14: Evaluation Service — FinalizeEvaluation (Advisory Lock)
**Layer:** Service  
**Estimated Time:** 2h  
**Dependencies:** T13

One-way finalization with PostgreSQL advisory lock.

- [ ] `FinalizeEvaluation(ctx, evalID, req)` acquires `pg_advisory_lock(int64(hash(evalID)))`
- [ ] Reads current evaluation; rejects if already `completada`
- [ ] Verifies cycle phase is `cierre`
- [ ] Transaction (`REPEATABLE READ`) updates state to `completada`, increments `version`
- [ ] Refreshes materialized view `evaluation_summary` concurrently
- [ ] Releases advisory lock (defer-safe)
- [ ] Returns finalized `EvaluationDetailResponse`
- [ ] Handles concurrent finalization attempts gracefully (only one succeeds)

---

### T15: NineBox Service — CreateMatrix
**Layer:** Service  
**Estimated Time:** 1.5h  
**Dependencies:** T7, T9

Matrix creation for an evaluator.

- [ ] `CreateMatrix(ctx, cycleID, evaluatorID)` delegates to repository
- [ ] Verifies evaluator is active in the cycle via `OrgService` (C5 dependency, stubbed)
- [ ] Pre-populates matrix with evaluatees from org hierarchy (optional, stubbed)
- [ ] Returns `NineBoxMatrixResponse` with empty `entries` array
- [ ] Returns `UNAUTHORIZED_EVALUATOR` if caller does not own matrix (TODO(auth:C7))

---

### T16: NineBox Service — UpsertEntry (with Quadrant Computation)
**Layer:** Service  
**Estimated Time:** 2h  
**Dependencies:** T15, T8, T2, T9

Single entry upsert with computed quadrant.

- [ ] `UpsertEntry(ctx, matrixID, req)` validates scores 1–9
- [ ] Verifies caller owns the matrix (TODO(auth:C7))
- [ ] Transaction: `SELECT FOR UPDATE` on existing entry
- [ ] Calls `ComputeQuadrant(perf, pot)` to determine quadrant
- [ ] Fetches `NineBoxQuadrant` metadata for denormalized response fields
- [ ] Upserts entry with `version++`
- [ ] Returns `NineBoxEntryDTO` with `quadrant`, `quadrantLabel`, `quadrantColor`
- [ ] Returns `QUADRANT_OUT_OF_RANGE` for invalid scores

---

### T17: NineBox Service — BatchSubmitEntries
**Layer:** Service  
**Estimated Time:** 2h  
**Dependencies:** T16

Atomic batch submission for up to 20 entries.

- [ ] `BatchSubmitEntries(ctx, matrixID, req)` validates max 20 entries
- [ ] Validates unique `evaluateeId`s within the batch
- [ ] Transaction (`REPEATABLE READ`): locks all existing entries for the matrix
- [ ] For each entry: compute quadrant, build upsert
- [ ] Executes all upserts atomically; any failure rolls back everything
- [ ] Re-fetches persisted rows to return complete DTOs
- [ ] Target: p95 < 300ms for 20 entries (local Postgres)
- [ ] Returns `[]NineBoxEntryDTO`

---

### T18: Dashboard Service — GetEvaluationSummary (Counts by State)
**Layer:** Service  
**Estimated Time:** 1.5h  
**Dependencies:** T4

Lightweight dashboard read using materialized view.

- [ ] `GetSummary(ctx, cycleID)` queries `evaluation_summary` materialized view
- [ ] Returns `EvaluationSummaryResponse` with `counts` map keyed by state
- [ ] Returns `0` counts for missing states rather than omitting keys
- [ ] Routes to read replica
- [ ] Refresh materialized view after finalization (called by T14)

---

### T19: Year-End Burst Queue (Redis + Workers)
**Layer:** Service  
**Estimated Time:** 3.5h  
**Dependencies:** T12, T13

Queue-based write smoothing for peak load.

- [ ] Middleware counts in-flight writes via Redis atomic counter (`writes:inflight:{orgID}`)
- [ ] If count > 500, serialize request + idempotency key and push to `queue:evaluations:{orgID}` (list) or priority sorted set
- [ ] Returns `202 Accepted` with `Retry-After: 5` header when queued
- [ ] Worker pool (10 goroutines per org) drains queue:
  - Batches up to 10 submissions per DB transaction
  - Priority: deadline < 24h processed first
- [ ] Decrements in-flight counter after worker completion
- [ ] Graceful shutdown: wait for in-flight workers to finish
- [ ] Unit test: 500+ submissions queued, workers drain within 10s

---

## Phase 4: Handler Layer (~8h)

### T20: Evaluation Handler — CRUD + Submission Endpoints
**Layer:** Handler  
**Estimated Time:** 2.5h  
**Dependencies:** T11, T12, T13, T14

Chi handlers for evaluation lifecycle.

- [ ] `GET /api/v1/evaluations` — `ListEvaluations` (query: cycle_id, state, cursor, limit)
- [ ] `GET /api/v1/evaluations/{id}` — `GetEvaluation`
- [ ] `POST /api/v1/evaluations/{id}/self-evaluation` — `SubmitSelfEvaluation` (header: `Idempotency-Key`)
- [ ] `PUT /api/v1/evaluations/{id}/self-evaluation` — `UpdateSelfEvaluation` (header: `If-Match`)
- [ ] `POST /api/v1/evaluations/{id}/rh-evaluation` — `SubmitRHEvaluation` (header: `Idempotency-Key`)
- [ ] `PUT /api/v1/evaluations/{id}/rh-evaluation` — `UpdateRHEvaluation` (header: `If-Match`)
- [ ] `POST /api/v1/evaluations/{id}/finalize` — `FinalizeEvaluation`
- [ ] JSON decode, DTO validation, service call, JSON encode
- [ ] Proper error mapping: `404`, `409`, `429`, `503`

---

### T21: NineBox Handler — Matrix + Entry Endpoints
**Layer:** Handler  
**Estimated Time:** 2.5h  
**Dependencies:** T15, T16, T17

Chi handlers for 9×9 matrix operations.

- [ ] `GET /api/v1/nine-box/matrices` — `ListMatrices` (query: cycle_id, evaluator_id)
- [ ] `GET /api/v1/nine-box/matrices/{matrixId}` — `GetMatrix`
- [ ] `POST /api/v1/nine-box/matrices` — `CreateMatrix`
- [ ] `GET /api/v1/nine-box/matrices/{matrixId}/entries` — `ListMatrixEntries`
- [ ] `POST /api/v1/nine-box/matrices/{matrixId}/entries` — `UpsertMatrixEntry`
- [ ] `PUT /api/v1/nine-box/entries/{entryId}` — `UpdateEntry` (header: `If-Match`)
- [ ] `POST /api/v1/nine-box/batch` — `BatchSubmitEntries`
- [ ] JSON decode, DTO validation, service call, JSON encode
- [ ] Proper error mapping: `400`, `404`, `409`

---

### T22: Dashboard Handler — Summary Endpoint
**Layer:** Handler  
**Estimated Time:** 1h  
**Dependencies:** T18

Single dashboard endpoint.

- [ ] `GET /api/v1/evaluations/summary` — `GetEvaluationSummary` (query: cycle_id)
- [ ] Returns `EvaluationSummaryResponse` with counts map
- [ ] Routes through read-replica middleware
- [ ] `304` handling if ETag matches (optional, can be stubbed)

---

### T23: Route Registration + Middleware
**Layer:** Handler  
**Estimated Time:** 2h  
**Dependencies:** T20, T21, T22

Wire all handlers into the Chi router.

- [ ] `routes.go` registers all 17 endpoints under `/api/v1/`
- [ ] Mounts evaluation handlers with `TODO(auth:C7)` middleware stubs
- [ ] Mounts nine-box handlers with `TODO(auth:C7)` middleware stubs
- [ ] Mounts rate-limiting middleware (Redis-backed, 100 req/s writes, 2000 req/s reads)
- [ ] Mounts circuit-breaker middleware (503 when pool > 90% for 10s)
- [ ] Mounts timeout middleware (read 5s, write 20s, batch 30s, finalize 30s)
- [ ] Mounts CORS and recovery middleware
- [ ] `go test ./handler/evaluation/...` passes

---

## Phase 5: OpenAPI (~3h)

### T24: OpenAPI 3.1 Spec
**Layer:** Contract  
**Estimated Time:** 2h  
**Dependencies:** T20, T21, T22

Complete OpenAPI specification as source of truth.

- [ ] YAML file generated/updated in `api/openapi/` or `openspec/specs/`
- [ ] All 17 endpoints documented with operationIds, parameters, request/response schemas
- [ ] All DTOs from design.md Section 4.2 represented as components/schemas
- [ ] Error responses (`404`, `409`, `429`, `503`) defined as reusable components/responses
- [ ] Validation annotations: `min=1,max=5`, `min=1,max=9`, `maxItems=20`, `format=uuid`
- [ ] `openapi-generator` or `swagger-codegen` validates without errors

---

### T25: TypeScript Type Generation
**Layer:** Contract  
**Estimated Time:** 1h  
**Dependencies:** T24

Generate TS types from OpenAPI for the frontend.

- [ ] Run `openapi-typescript` against the spec to produce `web/src/lib/api/types.ts`
- [ ] All request/response types exported (`EvaluationDetailResponse`, `NineBoxEntryDTO`, etc.)
- [ ] `tsc --noEmit` passes in `web/` without type errors
- [ ] Add npm script: `pnpm run generate:api` to regenerate types

---

## Phase 6: Tests (~12h)

### T26: Unit Tests — Repositories
**Layer:** Test  
**Estimated Time:** 2h  
**Dependencies:** T4–T10

Table-driven tests for all repository methods.

- [ ] `evaluation_repo_test.go`: GetByID, ListByCycle, GetDetail, SubmitCompetenciesAndGoals, Finalize, GetSummaryByCycle
- [ ] `ninebox_repo_test.go`: CreateMatrix, GetMatrixByID, ListMatrices, UpsertEntry, UpdateEntry, BatchUpsertEntries, GetScales, GetQuadrants
- [ ] `evaluation_goal_repo_test.go`: UpdateComments, VerifyGoalsExist
- [ ] `evaluation_competency_repo_test.go`: BulkUpsert, DeleteByEvaluation, GetByEvaluation
- [ ] All tests use `enttest` with SQLite or testcontainers Postgres
- [ ] `go test -race ./repository/...` passes

---

### T27: Unit Tests — Quadrant Computation (All 81 Combinations)
**Layer:** Test  
**Estimated Time:** 1h  
**Dependencies:** T2

Exhaustive table-driven test for `ComputeQuadrant`.

- [ ] Test all 81 combinations of `performance (1..9) × potential (1..9)`
- [ ] Assert deterministic quadrant mapping matches design.md grid
- [ ] Assert panic on out-of-range inputs: `0`, `10`, `-1`, `100`
- [ ] Coverage target: 100% of `pkg/quadrant`
- [ ] `go test -coverprofile=quadrant.out ./pkg/quadrant/...` shows 100%

---

### T28: Unit Tests — State Machine Transitions
**Layer:** Test  
**Estimated Time:** 1h  
**Dependencies:** T3

Test every valid and invalid state transition.

- [ ] Valid: `pendiente_evaluacion_final` → `en_progreso`, `en_progreso` → `completada`
- [ ] Invalid: `completada` → any, `pendiente_evaluacion_final` → `completada`
- [ ] `IsTerminal` returns true only for `completada`
- [ ] `RequiresPhase` returns `INVALID_PHASE` when phase != `cierre`
- [ ] `RequiresPhase` returns nil when phase == `cierre`

---

### T29: Unit Tests — Services
**Layer:** Test  
**Estimated Time:** 2h  
**Dependencies:** T11–T19

Mock-based tests for all service methods.

- [ ] `evaluation_service_test.go`: SubmitSelfEvaluation, UpdateSelfEvaluation, SubmitRHEvaluation, UpdateRHEvaluation, FinalizeEvaluation, GetEvaluation, ListEvaluations, GetSummary
- [ ] `ninebox_service_test.go`: CreateMatrix, GetMatrix, ListMatrices, UpsertEntry, UpdateEntry, BatchSubmitEntries, GetScales, GetQuadrants
- [ ] Mock repositories, Redis client, and `CycleService` (C2)
- [ ] Test error paths: `EVALUATION_ALREADY_FINALIZED`, `SELF_EVAL_DEADLINE_PASSED`, `INVALID_PHASE`, `CONCURRENT_UPDATE`
- [ ] `go test -race ./service/evaluation/...` passes

---

### T30: Unit Tests — Handlers
**Layer:** Test  
**Estimated Time:** 2h  
**Dependencies:** T20–T23

HTTP-level tests using `httptest` and mocked services.

- [ ] `evaluation_handler_test.go`: happy paths + validation errors for 8 evaluation endpoints
- [ ] `ninebox_handler_test.go`: happy paths + validation errors for 9 nine-box endpoints
- [ ] Assert `404` when resource missing
- [ ] Assert `409` when evaluation finalized or phase invalid
- [ ] Assert `409` on optimistic lock mismatch (`If-Match`)
- [ ] Assert `429` when rate limit exceeded
- [ ] Assert `503` when circuit breaker open
- [ ] `go test -race ./handler/evaluation/...` passes

---

### T31: Integration Tests
**Layer:** Test  
**Estimated Time:** 1.5h  
**Dependencies:** T26, T29

End-to-end flows with real database.

- [ ] Self-evaluation flow: create evaluation → submit self-eval → get detail → verify competencies and goals
- [ ] RH evaluation flow: submit RH eval → finalize → verify state `completada`
- [ ] Nine-box flow: create matrix → upsert entry → verify quadrant → batch submit → verify all entries
- [ ] Dashboard flow: finalize 3 evaluations → get summary → verify counts
- [ ] Uses testcontainers Postgres or local Docker instance
- [ ] `go test -tags=integration ./...` passes

---

### T32: Concurrency Test (100+ Goroutines Submitting Evaluations)
**Layer:** Test  
**Estimated Time:** 1h  
**Dependencies:** T29

Race-condition test for optimistic locking and idempotency.

- [ ] Spawn 100 goroutines calling `SubmitSelfEvaluation` on same evaluation ID with **same** idempotency key
- [ ] Assert exactly 1 transaction succeeds; 99 return cached result (200)
- [ ] Spawn 50 goroutines with **different** idempotency keys
- [ ] Assert exactly 1 wins `SELECT FOR UPDATE`; rest receive `409 CONCURRENT_UPDATE`
- [ ] Run with `go test -race`; no data races detected
- [ ] Final state is `en_progreso` (or `completada` if mocked)

---

### T33: Batch Test (20 Entries < 300ms)
**Layer:** Test  
**Estimated Time:** 0.5h  
**Dependencies:** T29

Latency and atomicity test for batch endpoint.

- [ ] Submit 20 valid entries via `BatchSubmitEntries`; assert all exist in DB
- [ ] Assert all quadrants computed correctly
- [ ] Submit 20 entries with one invalid score; assert transaction rolls back, zero entries modified
- [ ] Submit duplicate `evaluateeId`s in same batch; assert `400` validation error before DB touch
- [ ] Assert p95 latency < 300ms (local Postgres, warmed pool)

---

### T34: Load Test (1000 Concurrent Self-Evaluations)
**Layer:** Test  
**Estimated Time:** 1h  
**Dependencies:** T32

High-volume simulation.

- [ ] `POST /api/v1/evaluations/{id}/self-evaluation` with 1000 concurrent requests
- [ ] Target: all complete in < 30s, p95 latency < 500ms
- [ ] Use `k6` script or `go test` with `golang.org/x/sync/errgroup`
- [ ] Real Postgres instance in Docker (not SQLite)
- [ ] Report: throughput (req/s), p50/p95/p99 latency, error rate
- [ ] No `500` errors; acceptable `429` / `503` if rate limiting triggers

---

## Phase 7: Polish (~2h)

### T35: Documentation
**Layer:** Documentation  
**Estimated Time:** 1h  
**Dependencies:** All above

Developer-facing READMEs for the change.

- [ ] `api/internal/service/evaluation/README.md`: self-eval flow, RH-eval flow, state transitions, locking
- [ ] `api/internal/service/evaluation/README.md` (ninebox section or separate file): quadrant computation, batch logic
- [ ] `api/internal/handler/evaluation/README.md`: endpoint summary, auth TODOs, error codes
- [ ] `api/internal/handler/evaluation/README.md` (ninebox section): matrix endpoints, batch usage
- [ ] `api/internal/repository/evaluation/README.md`: transaction patterns, materialized view usage
- [ ] Include mermaid sequence diagrams for self-eval and finalize flows

---

### T36: Prometheus Metrics
**Layer:** Observability  
**Estimated Time:** 1h  
**Dependencies:** T23

Expose key metrics for monitoring the year-end burst.

- [ ] `pgx_pool_conns_busy` — gauge of active connections
- [ ] `pgx_pool_conns_idle` — gauge of idle connections
- [ ] `pgx_pool_wait_duration_ms` — histogram of wait time for connection
- [ ] `evaluation_submissions_total` — counter (labels: `type=[self|rh]`, `status=[success|error]`)
- [ ] `evaluation_finalizations_total` — counter (labels: `status=[success|error]`)
- [ ] `ninebox_entries_upserted_total` — counter (labels: `batch=[true|false]`)
- [ ] `queue_depth` — gauge of Redis queue depth per org
- [ ] `request_duration_seconds` — histogram for all endpoints (labels: `method`, `path`, `status`)
- [ ] Metrics endpoint `/metrics` registered in Chi router
- [ ] `go test` passes; metrics do not break existing tests

---

*Tasks generated for change C6: evaluations-and-9x9-api.*
*Source: `design.md` + `proposal.md` — SED Evaluación de Desempeño.*
