# C2: evaluation-lifecycle-api — Implementation Tasks

## Summary

| Phase | Tasks | Estimated Hours | Critical Path |
|-------|-------|-----------------|---------------|
| Phase 1: Infrastructure | 3 | 4h | T1 → T2 → T3 |
| Phase 2: Repository Layer | 3 | 6h | T4, T5, T6 (parallel after T1) |
| Phase 3: Service Layer | 4 | 8h | T7 → T8 → T9; T10 (parallel after T5, T6) |
| Phase 4: Handler Layer | 4 | 6h | T11, T12, T13 (parallel after T7, T8, T9, T10); T14 |
| Phase 5: OpenAPI | 2 | 3h | T15, T16 (parallel after T14) |
| Phase 6: Tests | 6 | 8h | T17, T18, T19, T20, T21, T22 (parallel after T14) |
| Phase 7: Polish | 2 | 2h | T23, T24 (parallel after T22) |
| **Total** | **24** | **~37h** | T1 → T3 → T4/T5/T6 → T7 → T8 → T9 → T11/T12/T13 → T14 → T15/T16/T17-T22 → T23/T24 |

---

## Phase 1: Infrastructure (~4h) [COMPLETE]

### T1 — Scaffold package directories and interfaces

**Description**: Create the Go package structure under `api/internal/` for handlers, services, repositories, middleware, and shared utilities. Define repository and service interfaces.

**Acceptance Criteria**:
- [x] Directories exist: `handler/cycle/`, `service/cycle/`, `repository/cycle/`, `middleware/`, `pkg/cursor/`, `pkg/errors/`
- [x] Repository interface defined in `repository/cycle/` with methods for CRUD, list, and transition
- [x] Service interface defined in `service/cycle/` matching design section 4.1
- [x] `go.mod` includes required dependencies (Chi, Ent)
- [x] Empty Go files compile (`go build ./...` passes without errors)

**Estimated Time**: 1h
**Dependencies**: none
**Layer**: infra

### T2 — Implement shared middleware (rate limit, idempotency, optimistic lock)

**Description**: Build the three middleware components: token-bucket rate limiter per org, Redis-backed idempotency key check, and `If-Match` optimistic locking gate.

**Acceptance Criteria**:
- [x] `RateLimit` middleware returns `429` with `RATE_LIMIT_EXCEEDED` when threshold exceeded; sets `X-RateLimit-*` headers
- [x] `Idempotency` middleware caches `2xx` responses under `idempotency:<key>` for 24h; returns `409` on key reuse with different payload
- [x] `OptimisticLock` middleware extracts `If-Match` header, parses version, injects into context; returns `428` if missing, `400` if malformed
- [x] Middleware chain is unit-testable with `httptest` + `miniredis`
- [x] `README.md` in `middleware/` explains each middleware and its Redis key pattern

**Estimated Time**: 2h
**Dependencies**: T1
**Layer**: middleware

### T3 — Implement error types and cursor pagination helpers

**Description**: Create standard domain error types in `pkg/errors/` and cursor encoder/decoder in `pkg/cursor/`. Ensure all errors map cleanly to HTTP status codes.

**Acceptance Criteria**:
- [x] Error types exist for: `CYCLE_NOT_FOUND`, `INVALID_TRANSITION`, `CYCLE_ALREADY_ACTIVE`, `PHASE_NOT_ADVANCEABLE`, `CONCURRENT_UPDATE`, `IDEMPOTENCY_KEY_CONFLICT`, `RATE_LIMIT_EXCEEDED`, `INVALID_REQUEST`
- [x] `HTTPStatus(err)` function returns correct status code for each error type (e.g., `409` for `CYCLE_ALREADY_ACTIVE`)
- [x] Cursor struct encodes/decodes to base64 JSON with `{id, updated_at}` tuple; returns `INVALID_REQUEST` on malformed input
- [x] `PaginatedList[T]` generic struct supports `data` + `pagination.next_cursor` + `pagination.has_more`
- [x] Error response body follows the standard JSON schema with `code`, `message`, `details`, `trace_id`

**Estimated Time**: 1h
**Dependencies**: T1
**Layer**: infra

---

## Phase 2: Repository Layer (~6h)

### T4 — Implement Cycle repository (CRUD + cursor queries with Ent)

**Description**: Write the Ent-backed repository for `Cycle` including create, get-by-ID, list with cursor pagination, and the optimistic-lock update used in transitions.

**Acceptance Criteria**:
- [x] `CreateCycle(ctx, tx, year, orgID)` returns `*ent.Cycle` with `current_phase = asignacion`, `version = 1`
- [x] `GetCycle(ctx, id)` returns `*ent.Cycle` or `CYCLE_NOT_FOUND` error
- [x] `ListCycles(ctx, orgID, year, phase, cursor, limit)` returns `[]*ent.Cycle` ordered by `updated_at DESC, id DESC`, supports `limit + 1` for `has_more`
- [x] `UpdatePhase(ctx, tx, cycleID, nextPhase, expectedVersion)` returns `CONCURRENT_UPDATE` if `RowsAffected == 0`
- [x] `InsertPhaseHistory(ctx, tx, cycleID, from, to, triggeredBy, reason)` writes audit row
- [x] `clientFor(ctx)` helper selects replica client when `dbRole = "replica"` is present in context

**Estimated Time**: 2.5h
**Dependencies**: T1, T3
**Layer**: repository

### T5 — Implement PhaseDefinition repository (read-only)

**Description**: Create a read-only repository for `PhaseDefinition` with a single method that returns all phase definitions. This will later be wrapped by Redis caching in the service layer.

**Acceptance Criteria**:
- [x] `GetPhaseDefinitions(ctx)` returns `[]*ent.PhaseDefinition` ordered by `order` ascending
- [x] `GetPhaseDefinitionByPhase(ctx, phase)` returns a single definition or `sql.ErrNoRows` wrapped as `CYCLE_NOT_FOUND`
- [x] Query uses `clientFor(ctx)` to route to replica when `dbRole = "replica"`
- [x] Repository file compiles and passes `go vet`
- [x] Interface is defined and wired into the service layer DI

**Estimated Time**: 1.5h
**Dependencies**: T1
**Layer**: repository

### T6 — Implement PhaseTransition repository (read-only)

**Description**: Create a read-only repository for `PhaseTransition` that queries transitions by `from_phase` and validates `(from_phase, to_phase, trigger)` tuples.

**Acceptance Criteria**:
- [x] `GetTransitionsByFromPhase(ctx, fromPhase)` returns `[]*ent.PhaseTransition` ordered by `to_phase` ascending
- [x] `ValidateTransition(ctx, fromPhase, toPhase, trigger)` returns `nil` if row exists, `INVALID_TRANSITION` if not found
- [x] Query uses `clientFor(ctx)` to route to replica when `dbRole = "replica"`
- [x] Repository file compiles and passes `go vet`
- [x] Interface is defined and wired into the service layer DI

**Estimated Time**: 2h
**Dependencies**: T1
**Layer**: repository

---

## Phase 3: Service Layer (~8h)

### T7 — Implement Cycle service — CreateCycle with advisory lock

**Description**: Build the `CreateCycle` service method that validates the request, acquires a PostgreSQL advisory lock to prevent duplicate `(org, year)` inserts, and delegates to the repository.

**Acceptance Criteria**:
- [x] `CreateCycle` validates `CreateCycleRequest` using `validator/v10`; returns `INVALID_REQUEST` on failure
- [x] Acquires advisory lock `pg_advisory_lock`(hashtext('cycle:create:' + orgID + ':' + year))` and defers `pg_advisory_unlock`
- [x] Inside a transaction, checks existing cycle for `(organization_id, year)`; returns `CYCLE_ALREADY_ACTIVE` if found
- [x] On success, commits transaction and returns `*Cycle` with `version = 1`, `current_phase = "asignacion"`
- [x] `IdempotencyKey` is read from context (injected by middleware) and passed through for deduplication
- [ ] Unit tests with mock repository cover: duplicate detection, validation failure, and happy path

**Estimated Time**: 2h
**Dependencies**: T4
**Layer**: service

### T8 — Implement Cycle service — TransitionPhase with transaction

**Description**: Build the `TransitionPhase` service method that locks the cycle row, validates the optimistic lock, resolves the next phase linearly, validates the transition rule, updates the cycle, and writes the audit log.

**Acceptance Criteria**:
- [x] `TransitionPhase` starts a `READ COMMITTED` transaction, locks cycle with `SELECT FOR UPDATE`
- [x] Compares `req.ExpectedVersion` with row `version`; returns `CONCURRENT_UPDATE` on mismatch
- [x] Resolves next phase linearly (`asignacion` → `avance` → `cierre`); returns `INVALID_TRANSITION` if no next phase exists
- [x] Validates transition against `PhaseTransition` repository; returns `INVALID_TRANSITION` if not defined
- [x] Updates cycle with `version + 1`; returns `CONCURRENT_UPDATE` if `RowsAffected == 0` (fallback raw SQL)
- [x] Inserts audit log into `cycle_phase_history` with `from_phase`, `to_phase`, `triggered_by`, `reason`
- [ ] Unit tests with mock repository cover: version mismatch, invalid transition, last phase (no next), and happy path

**Estimated Time**: 2.5h
**Dependencies**: T4, T6
**Layer**: service

### T9 — Implement Cycle service — ListCycles with cursor pagination

**Description**: Build the `ListCycles` service method that decodes the cursor, delegates to the repository, and assembles the `PaginatedList[Cycle]` response.

**Acceptance Criteria**:
- [x] `ListCycles` validates `ListCyclesRequest` (required `organization_id`, optional `year`, `current_phase`, `limit` clamped to `[1, 100]`)
- [x] Decodes cursor into `{id, updated_at}`; returns `INVALID_REQUEST` on malformed base64 or JSON
- [x] Queries with `limit + 1` to detect `has_more`; strips last row if over limit
- [x] Encodes next cursor from the last row returned; sets `has_more = true` when `limit + 1` rows exist
- [x] Returns `PaginatedList[Cycle]` with `data` and `pagination` fields
- [ ] Unit tests cover: empty list, cursor decode/encode, `has_more` toggle, and invalid limit

**Estimated Time**: 1.5h
**Dependencies**: T4, T3
**Layer**: service

### T10 — Implement Phase service — GetPhaseDefinitions and GetAvailableTransitions

**Description**: Build the two read-only service methods: `GetPhaseDefinitions` (with Redis caching) and `GetAvailableTransitions` (filtering by current cycle phase).

**Acceptance Criteria**:
- [x] `GetPhaseDefinitions` queries repository; computes ETag under `phases:definitions:v1` with TTL 1h
- [x] ETag computed as `SHA256(jsonPayload)[:16]`; stored in Redis alongside cached data
- [x] `GetAvailableTransitions` fetches cycle by ID, then queries `PhaseTransition` where `from_phase = cycle.current_phase`
- [x] Returns `CYCLE_NOT_FOUND` if cycle does not exist
- [x] Returns transitions ordered by `to_phase` ascending
- [ ] Unit tests with mock repository cover: cache hit, cache miss, cycle not found, and transitions from each phase

**Estimated Time**: 2h
**Dependencies**: T5, T6, T4
**Layer**: service

---

## Phase 4: Handler Layer (~6h)

### T11 — Implement Cycle handler — GET /cycles and POST /cycles

**Description**: Build the HTTP handlers for listing and creating cycles. `GET` decodes query params, calls `ListCycles`, and returns `200` with pagination. `POST` decodes body, calls `CreateCycle`, and returns `201`.

**Acceptance Criteria**:
- [x] `GET /api/v1/cycles` accepts `organization_id` (required), `year`, `current_phase`, `cursor`, `limit`; returns `200` with `PaginatedList[Cycle]` JSON
- [x] `POST /api/v1/cycles` accepts `CreateCycleRequest` body; validates via `validator/v10`; returns `201` with full `Cycle` JSON
- [x] `POST` reads `Idempotency-Key` from header (injected by middleware); returns `409` on conflict
- [x] `GET` sets `dbRole = "replica"` in context before calling service
- [x] Handler maps domain errors to correct HTTP status via `pkg/errors.HTTPStatus`
- [ ] Unit tests with mock service cover: happy path, validation errors, 409 duplicate, and missing org param

**Estimated Time**: 2h
**Dependencies**: T9, T7
**Layer**: handler

### T12 — Implement Cycle handler — GET /cycles/:id and PUT /cycles/:id/transition

**Description**: Build the HTTP handlers for retrieving a single cycle and advancing its phase. `GET` returns the full cycle. `PUT` requires `If-Match` and `Idempotency-Key`.

**Acceptance Criteria**:
- [x] `GET /api/v1/cycles/:id` returns `200` with full `Cycle` JSON or `404` with `CYCLE_NOT_FOUND`
- [x] `PUT /api/v1/cycles/:id/transition` requires `If-Match` (optimistic lock) and `Idempotency-Key`
- [x] `PUT` accepts optional body with `trigger` and `reason`; returns `200` with updated `Cycle` JSON
- [x] `PUT` returns `409` for `CONCURRENT_UPDATE`, `INVALID_TRANSITION`, `PHASE_NOT_ADVANCEABLE`, or `IDEMPOTENCY_KEY_CONFLICT`
- [x] `GET` sets `dbRole = "replica"` in context before calling service
- [ ] Unit tests with mock service cover: 404, 409 concurrency, 409 invalid transition, missing `If-Match`, and happy path

**Estimated Time**: 2h
**Dependencies**: T8, T3
**Layer**: handler

### T13 — Implement Phase handler — GET /phases and GET /cycles/:id/transitions

**Description**: Build the HTTP handlers for the phase catalog and available transitions. `GET /phases` supports ETag caching. `GET /cycles/:id/transitions` returns transitions from the current phase.

**Acceptance Criteria**:
- [x] `GET /api/v1/phases` returns `200` with `data` array of `PhaseDefinition`; supports `ETag` / `If-None-Match` and returns `304` when unchanged
- [x] `GET /api/v1/cycles/:id/transitions` returns `200` with `data` array of `PhaseTransition`; returns `404` if cycle not found
- [x] Both endpoints set `dbRole = "replica"` in context before calling service
- [x] `GET /phases` sets `Cache-Control: max-age=3600` in response
- [x] Handler maps domain errors to correct HTTP status via `pkg/errors.HTTPStatus`
- [ ] Unit tests with mock service cover: 304, 200, 404, and ETag generation

**Estimated Time**: 1.5h
**Dependencies**: T10
**Layer**: handler

### T14 — Register routes and wire middleware stack

**Description**: Wire all handlers into the Chi router with the correct middleware stack per endpoint: auth placeholder, rate limit, idempotency, optimistic lock, and replica routing.

**Acceptance Criteria**:
- [ ] `GET /api/v1/cycles` stack: `AuthPlaceholder` → `RateLimit(read)` → `ReadReplica`
- [ ] `POST /api/v1/cycles` stack: `AuthPlaceholder` → `RateLimit(write)` → `Idempotency`
- [ ] `GET /api/v1/cycles/:id` stack: `AuthPlaceholder` → `RateLimit(read)` → `ReadReplica`
- [ ] `PUT /api/v1/cycles/:id/transition` stack: `AuthPlaceholder` → `RateLimit(write)` → `Idempotency` → `OptimisticLock`
- [ ] `GET /api/v1/phases` stack: `AuthPlaceholder` → `RateLimit(read)` → `ETagCache` → `ReadReplica`
- [ ] `GET /api/v1/cycles/:id/transitions` stack: `AuthPlaceholder` → `RateLimit(read)` → `ReadReplica`
- [ ] `routes.go` is unit-tested with `httptest` to verify middleware order and 404/405 responses

**Estimated Time**: 0.5h
**Dependencies**: T11, T12, T13, T2
**Layer**: handler

---

## Phase 5: OpenAPI (~3h)

### T15 — Generate and validate OpenAPI 3.1 specification

**Description**: Write the complete `api/openapi/cycle.yaml` covering all six endpoints, request/response schemas, and error components. Validate against OpenAPI 3.1 schema.

**Acceptance Criteria**:
- [x] `cycle.yaml` contains paths for all 6 endpoints matching design section 3.1
- [x] Schemas defined: `Cycle`, `CycleLight`, `CreateCycleRequest`, `TransitionPhaseRequest`, `PhaseDefinition`, `PhaseTransition`, `CursorPagination`, `Error`
- [x] Error responses reference `#/components/responses/BadRequest`, `NotFound`, `Conflict`, `RateLimit`
- [ ] `openapi-generator` or `swagger-codegen` validates the spec without errors
- [ ] `README.md` in `openapi/` explains how to validate the spec (`make validate-openapi`)

**Estimated Time**: 2h
**Dependencies**: T14
**Layer**: infra

### T16 — Generate TypeScript types from OpenAPI spec

**Description**: Run `openapi-typescript` against `cycle.yaml` to generate TypeScript types for the frontend. Output to `web/src/lib/api/cycle.types.ts`.

**Acceptance Criteria**:
- [ ] `openapi-typescript api/openapi/cycle.yaml --output web/src/lib/api/cycle.types.ts` runs without errors
- [ ] Generated file includes types for all request/response bodies and pagination objects
- [ ] `web/src/lib/api/cycle.types.ts` compiles in a Vite + TypeScript project
- [ ] No manual edits required to the generated file; it is auto-generated via `pnpm run gen:api` or similar
- [ ] `package.json` script added: `"gen:api": "openapi-typescript api/openapi/cycle.yaml --output web/src/lib/api/cycle.types.ts"`

**Estimated Time**: 1h
**Dependencies**: T15
**Layer**: infra

---

## Phase 6: Tests (~8h)

### T17 — Write unit tests for repository layer

**Description**: Unit tests for `CycleRepository`, `PhaseDefinitionRepository`, and `PhaseTransitionRepository` using an in-memory PostgreSQL container via `testcontainers-go`. Verify Ent queries, index coverage, and optimistic-lock behavior.

**Acceptance Criteria**:
- [ ] `TestCreateCycle` verifies row creation and default values (`version = 1`, `phase = asignacion`)
- [ ] `TestGetCycle` verifies retrieval by UUID and `CYCLE_NOT_FOUND` for missing ID
- [ ] `TestListCycles` verifies cursor pagination, ordering, and `has_more` detection
- [ ] `TestUpdatePhase` verifies optimistic-lock success and `CONCURRENT_UPDATE` on stale version
- [ ] `TestInsertPhaseHistory` verifies audit row creation with correct fields
- [ ] `TestReplicaRouting` verifies `clientFor(ctx)` returns replica client when `dbRole = "replica"`
- [ ] `TestMain` spins up Postgres container, runs `atlas migrate apply`, and tears down
- [ ] All tests pass with `go test ./repository/cycle/...`

**Estimated Time**: 2h
**Dependencies**: T14
**Layer**: test

### T18 — Write unit tests for service layer (including concurrency)

**Description**: Unit tests for `CycleService` and `PhaseService` using `testify/mock` to stub repositories. Cover business rules, validation, and simulated concurrency.

**Acceptance Criteria**:
- [ ] `TestCreateCycle` covers: duplicate cycle returns `CYCLE_ALREADY_ACTIVE`, validation failure, and happy path
- [ ] `TestTransitionPhase` covers: version mismatch → `CONCURRENT_UPDATE`, invalid transition → `INVALID_TRANSITION`, last phase → `INVALID_TRANSITION`, and happy path
- [ ] `TestListCycles` covers: cursor decode/encode, limit clamping, empty list, and `has_more`
- [ ] `TestGetPhaseDefinitions` covers: repository call and cache interaction (mock Redis)
- [ ] `TestGetAvailableTransitions` covers: cycle not found → `CYCLE_NOT_FOUND`, and transitions from each phase
- [ ] `TestConcurrentTransitionSimulation` mocks 5 simultaneous calls and verifies exactly 1 success + 4 failures
- [ ] All tests pass with `go test ./service/cycle/...`

**Estimated Time**: 2h
**Dependencies**: T14
**Layer**: test

### T19 — Write unit tests for handler layer

**Description**: Unit tests for all HTTP handlers using `httptest.ResponseRecorder` + mock service interface. Verify status codes, header parsing, error bodies, and JSON serialization.

**Acceptance Criteria**:
- [ ] `TestListCyclesHandler` covers: 200 with pagination, 400 missing `organization_id`, 429 rate limit
- [ ] `TestCreateCycleHandler` covers: 201, 409 duplicate, 409 idempotency conflict, 400 validation
- [ ] `TestGetCycleHandler` covers: 200, 404, 400 bad UUID
- [ ] `TestTransitionPhaseHandler` covers: 200, 404, 409 concurrency, 409 invalid transition, 428 missing `If-Match`, 400 bad `If-Match`
- [ ] `TestGetPhaseDefinitionsHandler` covers: 200, 304 ETag match
- [ ] `TestGetAvailableTransitionsHandler` covers: 200, 404 cycle not found
- [ ] All tests pass with `go test ./handler/cycle/...`

**Estimated Time**: 1.5h
**Dependencies**: T14
**Layer**: test

### T20 — Write integration tests with testcontainers

**Description**: End-to-end integration tests that spin up PostgreSQL and Redis containers, apply migrations, and test the full request → handler → service → repository → DB flow.

**Acceptance Criteria**:
- [ ] `TestIntegrationCreateAndGetCycle` creates a cycle via `POST` and retrieves it via `GET`
- [ ] `TestIntegrationListCycles` creates multiple cycles and verifies cursor pagination across pages
- [ ] `TestIntegrationTransitionPhase` creates a cycle, transitions it, and verifies `version` increment + audit log
- [ ] `TestIntegrationIdempotency` resends `POST` with same key and receives `201` with identical body; different payload returns `409`
- [ ] `TestIntegrationRateLimit` exceeds read threshold and receives `429`
- [ ] `TestIntegrationReplicaRouting` verifies GET queries execute against the replica pool (mocked via separate container or connection string)
- [ ] `TestMain` starts Postgres + Redis, applies Atlas migrations, and stops containers on exit
- [ ] All tests pass with `go test ./integration/...`

**Estimated Time**: 1.5h
**Dependencies**: T14, T17
**Layer**: test

### T21 — Write concurrency test (50 goroutines)

**Description**: A dedicated race-condition test that launches 50 goroutines attempting to transition the same cycle simultaneously. Expect exactly one success, the rest returning `409`.

**Acceptance Criteria**:
- [ ] `TestTransitionPhaseConcurrent` creates a single cycle, launches 50 goroutines calling `PUT /cycles/:id/transition`
- [ ] Exactly 1 goroutine receives `200`; the rest receive `409` (`CONCURRENT_UPDATE` or `PHASE_NOT_ADVANCEABLE`)
- [ ] After all goroutines finish, the cycle's `current_phase` is `avance` (or `cierre` if run twice)
- [ ] `cycle_phase_history` contains exactly 1 audit row for the transition (or 2 if two sequential runs)
- [ ] No duplicate transitions (same cycle, same from/to, same timestamp) — no race-induced duplicates
- [ ] Test passes with `go test -race ./...` (no data races detected)

**Estimated Time**: 0.5h
**Dependencies**: T20
**Layer**: test

### T22 — Write load test script (k6)

**Description**: A k6 script for load testing the three critical scenarios: list cycles at 1000 req/s, create cycles at 100 req/s, and transition cycles at 50 req/s burst.

**Acceptance Criteria**:
- [ ] `loadtest/cycles.js` contains three scenarios: `list` (1000 req/s, 60s), `create` (100 req/s, 60s), `transition` (50 req/s burst)
- [ ] `list` scenario asserts p95 latency < 200ms and error rate < 0.1%
- [ ] `create` scenario asserts zero `CYCLE_ALREADY_ACTIVE` collisions beyond expected (only 1 per org/year)
- [ ] `transition` scenario asserts exactly 1 success per cycle, rest 409
- [ ] Script reports HTTP status distribution, PostgreSQL `pg_stat_statements` average latency, and Redis memory usage
- [ ] `README.md` in `loadtest/` explains how to run: `k6 run loadtest/cycles.js`

**Estimated Time**: 1h
**Dependencies**: T20
**Layer**: test

---

## Phase 7: Polish (~2h)

### T23 — Write documentation (README for each package)

**Description**: Create concise `README.md` files in `internal/service/cycle/`, `internal/handler/cycle/`, `internal/repository/cycle/`, and `middleware/` explaining the package's role, key flows, and how to run tests.

**Acceptance Criteria**:
- [ ] `internal/service/cycle/README.md` explains the transition flow (lock → validate → update → audit) and advisory-lock rationale
- [ ] `internal/handler/cycle/README.md` lists all endpoints, middleware stacks, and error codes
- [ ] `internal/repository/cycle/README.md` explains Ent client lifecycle, transaction wrapper, and read-replica routing
- [ ] `middleware/README.md` documents rate limit, idempotency, and optimistic-lock middleware with Redis key patterns
- [ ] Each README includes a `go test` command snippet for the package
- [ ] Docs are reviewed for spelling and consistent formatting (Markdown lint clean)

**Estimated Time**: 1h
**Dependencies**: T22
**Layer**: infra

### T24 — Instrument Prometheus metrics

**Description**: Add Prometheus metrics instrumentation to handlers, services, and middleware. Expose a `/metrics` endpoint (or prepare for it) tracking request latency, rate-limit hits, and pool health.

**Acceptance Criteria**:
- [ ] `http_request_duration_seconds` histogram tracked per endpoint and status code
- [ ] `rate_limit_hits_total` counter per endpoint and `organization_id`
- [ ] `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms` gauges exposed from pgx pool stats
- [ ] `cycle_transition_total` counter with labels `from_phase`, `to_phase`, `trigger`, `status` (success | conflict)
- [ ] `idempotency_cache_hit_total` counter with labels `endpoint` and `hit` (true | false)
- [ ] Metrics registration is compatible with `prometheus/client_golang` and can be served via `/metrics` in a future change
- [ ] Unit tests verify metrics are incremented on the correct code paths

**Estimated Time**: 1h
**Dependencies**: T22
**Layer**: infra

---

## Dependency Graph

```
T1 ──────────────────────┐
   ├── T2 ───────────────┼── T14 ───────────────────────────┐
   │                     │                                   │
   └── T3 ───────────────┼── T4 ─── T7 ─── T11 ─────────────┤
                         │             └── T8 ─── T12 ──────┤
                         │             └── T9 ──────────────┤
                         ├── T5 ─── T10 ─── T13 ─────────────┤
                         └── T6 ─── T8 ─────────────────────┤
                                                            │
                                                            ├── T15 ─── T16
                                                            │
                                                            ├── T17 ─── T21
                                                            ├── T18 ─── T21
                                                            ├── T19 ─── T21
                                                            └── T20 ─── T22 ─── T23
                                                                                └── T24
```

## Notes

- **TODO(auth:C7)**: All auth and RBAC enforcement is deferred to change C7. Middleware and handlers contain placeholders that inject mock identity (`rh` role) for development.
- **Atlas migrations**: Ensure `atlas migrate apply` is run in `TestMain` before integration and concurrency tests.
- **Redis**: Integration tests use `miniredis` or a `testcontainers` Redis container. Load tests assume a real Redis instance.
- **Replica routing**: If no read replica is configured in the test environment, `clientFor(ctx)` should fall back to the primary client to avoid test failures.
