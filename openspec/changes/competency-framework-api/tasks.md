# Tasks: C3 — Competency Framework API

## Summary

| Phase | Tasks | Hours |
|---|---|---|
| Phase 1: Infrastructure | 1 – 3 | 3 |
| Phase 2: Repository Layer | 4 – 9 | 8 |
| Phase 3: Service Layer | 10 – 15 | 10 |
| Phase 4: Handler Layer | 16 – 20 | 8 |
| Phase 5: OpenAPI | 21 – 22 | 3 |
| Phase 6: Tests | 23 – 29 | 10 |
| Phase 7: Polish | 30 – 31 | 2 |
| **Total** | **31** | **44** |

---

## Phase 1: Infrastructure

### Task 1 — Package scaffolding
**Description:**
Create all directories and placeholder files defined in the design package structure. Define DTO structs in `api/internal/dto/competency/` and stub `interfaces.go` for repository and service layers so downstream tasks compile.

**Acceptance Criteria:**
- [ ] All directories from design section 2 exist and are tracked.
- [ ] DTO files (`pillar_dto.go`, etc.) compile without errors.
- [ ] Placeholder `interfaces.go` files created for `repository/competency/` and `service/competency/`.
- [ ] `go build ./api/...` passes.

**Estimated Time:** 1h  
**Dependencies:** C1 (data-model-core applied)  
**Layer:** infra

---

### Task 2 — Shared middleware
**Description:**
Implement or reuse middleware from C2: rate limiter, idempotency, optimistic lock, read-replica routing, and timeout. Wire the Redis client where required. Integrate `chi` middleware patterns.

**Acceptance Criteria:**
- [ ] Rate limiter distinguishes read vs write limits (50/500 RPS).
- [ ] Idempotency middleware validates UUID keys, caches responses for 24h, and replays on collision.
- [ ] Optimistic lock middleware extracts and validates `If-Match` (RFC3339Nano).
- [ ] Read-replica middleware selects the correct `ent.Client` based on `dbRole`.
- [ ] Timeout middleware applies 3s (GET), 8s (write), 15s (cascade delete).

**Estimated Time:** 1h  
**Dependencies:** 1  
**Layer:** infra

---

### Task 3 — Redis cache helpers & ETag
**Description:**
Implement cache helpers for the competency tree (`competency:tree:v1`, TTL 1h) and invalidation triggers. Build `pkg/etag` and `pkg/cursor` utilities with tests.

**Acceptance Criteria:**
- [ ] `SetCompetencyTree` and `InvalidateCompetencyTree` helpers exist.
- [ ] ETag generation hashes a slice of structs deterministically.
- [ ] Cursor encoding/decoding uses base64 JSON and round-trips correctly.
- [ ] All helper packages have unit tests.

**Estimated Time:** 1h  
**Dependencies:** 1  
**Layer:** infra

---

## Phase 2: Repository Layer

### Task 4 — Pillar repository (CRUD + cascade)
**Description:**
Implement `PillarRepo` interface: `WithTx`, `List`, `Get`, `Create`, `Update`, `Delete`, `CountCompetencies`. Use cursor-based pagination on `name`, eager-load competencies when requested, and `SELECT FOR UPDATE` in delete.

**Acceptance Criteria:**
- [ ] `List` returns paginated results with a `next_cursor` and `has_more` flag.
- [ ] `Get` supports eager loading via `WithCompetencies`.
- [ ] `Delete` acquires a row lock and returns typed errors for `NotFound`.
- [ ] `Create` maps unique-constraint violations to `DUPLICATE_NAME`.

**Estimated Time:** 1.5h  
**Dependencies:** 1  
**Layer:** repository

---

### Task 5 — Competency repository (CRUD + eager loading)
**Description:**
Implement `CompetencyRepo` interface: `WithTx`, `ListByPillar`, `Get`, `Create`, `Update`, `Delete`. Support moving a competency between pillars and cursor pagination.

**Acceptance Criteria:**
- [ ] `ListByPillar` uses cursor pagination (name-based) and limit.
- [ ] `Create` validates the target pillar FK exists.
- [ ] `Update` handles `pillar_id` moves inside a transaction.
- [ ] `Delete` cascades to scale criteria and acceptance levels via Ent edges.

**Estimated Time:** 1.5h  
**Dependencies:** 1, 4  
**Layer:** repository

---

### Task 6 — ScaleCriterion repository (bulk write, version increment)
**Description:**
Implement `ScaleRepo` interface: `WithTx`, `GetByCompetency`, `ReplaceAll`. `ReplaceAll` must bulk-delete existing criteria, bulk-insert new ones, and bump the competency `version` in the same transaction.

**Acceptance Criteria:**
- [ ] `ReplaceAll` deletes all existing criteria for the competency then inserts new ones.
- [ ] Competency `version` is incremented atomically within the same tx.
- [ ] `GetByCompetency` returns criteria ordered by `level`.
- [ ] Transaction rolls back on any step failure.

**Estimated Time:** 1.5h  
**Dependencies:** 1, 5  
**Layer:** repository

---

### Task 7 — LevelDefinition repository (read-only)
**Description:**
Implement `CatalogRepo.ListLevels`. Route to the read-replica client. Return all 5 fixed level definitions in stable order.

**Acceptance Criteria:**
- [ ] Returns all 5 level definitions (levels 1–5).
- [ ] Uses the read-replica `ent.Client`.
- [ ] Ordering is stable (by `level` ASC).
- [ ] No N+1 queries.

**Estimated Time:** 1h  
**Dependencies:** 1  
**Layer:** repository

---

### Task 8 — EvaluationProfile repository (read-only)
**Description:**
Implement `CatalogRepo.ListProfiles`. Route to the read-replica client. Return all 8 fixed evaluation profiles in stable order.

**Acceptance Criteria:**
- [ ] Returns all 8 evaluation profiles.
- [ ] Uses the read-replica `ent.Client`.
- [ ] Ordering is stable (by `name` ASC).
- [ ] No N+1 queries.

**Estimated Time:** 1h  
**Dependencies:** 1  
**Layer:** repository

---

### Task 9 — CompetencyAcceptanceLevel repository (upsert)
**Description:**
Implement `AcceptanceRepo` interface: `WithTx`, `List`, `Upsert`. `List` supports optional filtering by `profile_id` and `competency_id`. `Upsert` must create or update on the unique key `(competency_id, profile_id)`.

**Acceptance Criteria:**
- [ ] `List` filters correctly when zero, one, or both query params are provided.
- [ ] `Upsert` creates a new row when the pair does not exist.
- [ ] `Upsert` updates `level` and `updated_at` when the pair already exists.
- [ ] Uses transaction wrapper `WithTx` for atomicity.

**Estimated Time:** 1.5h  
**Dependencies:** 1, 5  
**Layer:** repository

---

## Phase 3: Service Layer

### Task 10 — Pillar service: Create/Update with validation
**Description:**
Implement `PillarService.List`, `Get`, `Create`, and `Update`. Map `ent` entities to DTOs. Enforce validation rules (name length, description length) and optimistic locking on `Update`.

**Acceptance Criteria:**
- [ ] `Create` returns `PillarDetail`; maps `DUPLICATE_NAME` to 409.
- [ ] `Update` checks `If-Match` against `updated_at`; returns `CONCURRENT_UPDATE` on mismatch.
- [ ] Validation rejects empty names or descriptions > 2000 chars.
- [ ] `List` items include `competency_count` without N+1.

**Estimated Time:** 1.5h  
**Dependencies:** 4  
**Layer:** service

---

### Task 11 — Pillar service: Delete with cascade + advisory lock
**Description:**
Implement `PillarService.Delete`. Acquire a PostgreSQL advisory lock on the pillar ID to guard against concurrent deletes. Enforce `force` logic: reject if `force=false` and competencies exist; otherwise cascade-delete within a transaction and invalidate the Redis tree cache.

**Acceptance Criteria:**
- [ ] Advisory lock acquired before transaction starts.
- [ ] Returns `PILLAR_HAS_COMPETENCIES` (409) when `force=false` and count > 0.
- [ ] Cascade delete removes competencies, scale criteria, and acceptance levels in one tx.
- [ ] Redis competency tree cache is invalidated on successful delete.

**Estimated Time:** 1.5h  
**Dependencies:** 4, 10  
**Layer:** service

---

### Task 12 — Competency service: CRUD with cascade
**Description:**
Implement `CompetencyService`: `ListByPillar`, `Get`, `Create`, `Update`, `Delete`. Support moving a competency between pillars. `Delete` must check for scale criteria and reject if `force=false`.

**Acceptance Criteria:**
- [ ] `Create` validates the target pillar exists.
- [ ] `Update` validates the new pillar exists when `pillar_id` is changed.
- [ ] `Delete` returns `COMPETENCY_HAS_CRITERIA` (409) when `force=false`.
- [ ] `Get` returns `CompetencyDetail` with `scale_criteria` grouped by level.

**Estimated Time:** 2h  
**Dependencies:** 5, 10  
**Layer:** service

---

### Task 13 — ScaleCriterion service: Bulk write for criteria
**Description:**
Implement `ScaleService.GetByCompetency` and `Upsert`. Validate that levels are 1–5 and that no duplicate levels exist in the request. Use the repository `ReplaceAll` method, then refresh and map the result.

**Acceptance Criteria:**
- [ ] `Upsert` returns `INVALID_LEVEL` (400) for any level outside 1–5.
- [ ] `Upsert` returns `DUPLICATE_LEVEL` (400) if the request contains duplicate levels.
- [ ] Replaces all existing criteria in a single transaction.
- [ ] Returns updated criteria, new version, and `updated_at`.

**Estimated Time:** 1.5h  
**Dependencies:** 6, 12  
**Layer:** service

---

### Task 14 — AcceptanceLevel service: Upsert
**Description:**
Implement `AcceptanceService.List` and `Upsert`. Validate FK existence (competency and profile) and level range 1–5. Map `ent` results to `AcceptanceLevelItem` DTOs.

**Acceptance Criteria:**
- [ ] `Upsert` validates competency and profile exist; returns `RESOURCE_NOT_FOUND` (404) if not.
- [ ] `Upsert` validates `level` is 1–5; returns `INVALID_LEVEL` (400) if not.
- [ ] `List` supports optional `profile_id` and `competency_id` filters.
- [ ] Returns `AcceptanceLevelItem` with correct `created_at` / `updated_at`.

**Estimated Time:** 1.5h  
**Dependencies:** 9, 12  
**Layer:** service

---

### Task 15 — Catalog service: GetLevels, GetProfiles (with ETag)
**Description:**
Implement `CatalogService.ListLevels` and `ListProfiles`. Compute an ETag from the payload hash. Support `If-None-Match` by returning a 304 when the hash matches. Cache results in Redis with a long TTL.

**Acceptance Criteria:**
- [ ] `GET /levels` response includes `ETag: "levels:v1:<hash>"`.
- [ ] Returns `304 Not Modified` when `If-None-Match` matches current ETag.
- [ ] `GET /profiles` follows the same ETag pattern.
- [ ] Uses read-replica client and Redis cache.

**Estimated Time:** 2h  
**Dependencies:** 7, 8, 3  
**Layer:** service

---

## Phase 4: Handler Layer

### Task 16 — Pillar handler: CRUD endpoints
**Description:**
Implement `ListPillars`, `CreatePillar`, `GetPillar`, `UpdatePillar`, `DeletePillar`. Parse query parameters (`include`, `cursor`, `limit`), path parameters, and map service errors to the correct HTTP status codes per the design.

**Acceptance Criteria:**
- [ ] `ListPillars` returns 200 with pagination object.
- [ ] `CreatePillar` returns 201 with `PillarDetail`.
- [ ] `GetPillar` returns 404 `PILLAR_NOT_FOUND` when missing.
- [ ] `UpdatePillar` returns 409 for `CONCURRENT_UPDATE` or `DUPLICATE_NAME`.
- [ ] `DeletePillar` returns 204 or 409 `PILLAR_HAS_COMPETENCIES`.

**Estimated Time:** 2h  
**Dependencies:** 10, 11, 2  
**Layer:** handler

---

### Task 17 — Competency handler: CRUD endpoints
**Description:**
Implement `ListCompetenciesByPillar`, `CreateCompetency`, `GetCompetency`, `UpdateCompetency`, `DeleteCompetency`. Parse `pillarId` path param and wire middleware as specified in the design.

**Acceptance Criteria:**
- [ ] `ListCompetenciesByPillar` returns paginated `CompetencyLite` items.
- [ ] `CreateCompetency` returns 201; validates pillar exists (404 if not).
- [ ] `GetCompetency` returns `CompetencyDetail` with `scale_criteria` grouped by level.
- [ ] `UpdateCompetency` supports pillar move and optimistic locking.
- [ ] `DeleteCompetency` handles `force` query parameter.

**Estimated Time:** 2h  
**Dependencies:** 12, 13, 2  
**Layer:** handler

---

### Task 18 — ScaleCriterion handler: Bulk endpoint
**Description:**
Implement `GetScaleCriteria` and `UpsertScaleCriteria`. Parse competency `id` from the path. Wire idempotency middleware for the POST handler.

**Acceptance Criteria:**
- [ ] `GetScaleCriteria` returns criteria grouped by level and current `version`.
- [ ] `UpsertScaleCriteria` returns 200 with updated criteria and new `version`.
- [ ] Returns 400 for `INVALID_LEVEL` or `DUPLICATE_LEVEL`.
- [ ] Returns 404 `COMPETENCY_NOT_FOUND` when the competency does not exist.

**Estimated Time:** 1.5h  
**Dependencies:** 13, 2  
**Layer:** handler

---

### Task 19 — Catalog handler: Levels, Profiles, Acceptance Levels
**Description:**
Implement `ListLevels`, `ListProfiles`, `ListAcceptanceLevels`, `UpsertAcceptanceLevel`. Add ETag support for Levels and Profiles. Parse query filters for acceptance levels.

**Acceptance Criteria:**
- [ ] `ListLevels` returns array and sets `ETag` header.
- [ ] `ListProfiles` returns array and sets `ETag` header.
- [ ] `ListAcceptanceLevels` filters by `profile_id` and/or `competency_id`.
- [ ] `UpsertAcceptanceLevel` returns 200 with `AcceptanceLevelItem`.

**Estimated Time:** 1h  
**Dependencies:** 14, 15, 2  
**Layer:** handler

---

### Task 20 — Route registration + middleware wiring
**Description:**
Create `routes.go` in `handler/competency`. Register all 16 endpoints under `/api/v1/`. Apply Chi middleware groups (`Timeout`, `RateLimiter`, `Idempotency`, `OptimisticLock`) and inject the dependency container (repos, services, Redis, ent clients).

**Acceptance Criteria:**
- [ ] All 16 endpoints registered with correct HTTP methods and paths.
- [ ] POST writes protected by `Idempotency` middleware where required.
- [ ] PUT endpoints protected by `OptimisticLock` middleware.
- [ ] GET endpoints use read-replica and timeout middleware.
- [ ] Dependency container is passed cleanly to handlers.

**Estimated Time:** 1.5h  
**Dependencies:** 16, 17, 18, 19  
**Layer:** handler

---

## Phase 5: OpenAPI

### Task 21 — OpenAPI 3.1 spec
**Description:**
Write `api/openapi/competency-framework-api.yaml` covering all 16 endpoints, schemas, reusable responses, and security markers (`TODO(auth:C7)`). Validate the spec with an OpenAPI linter/generator.

**Acceptance Criteria:**
- [ ] YAML covers all 16 endpoints with correct paths, methods, and parameters.
- [ ] All request/response schemas defined and referenced.
- [ ] Reusable error responses (`BadRequest`, `NotFound`, `Conflict`, etc.) live in `components/responses`.
- [ ] Validation passes without errors or warnings.

**Estimated Time:** 1.5h  
**Dependencies:** 20  
**Layer:** openapi

---

### Task 22 — TypeScript type generation
**Description:**
Generate TypeScript types from the OpenAPI spec into `web/src/lib/api/` using `openapi-typescript`. Verify that generated interfaces match the Go DTOs and that the frontend build passes.

**Acceptance Criteria:**
- [ ] Types generated in `web/src/lib/api/` from the spec.
- [ ] All DTOs present as TypeScript interfaces (no missing fields).
- [ ] Frontend build (`pnpm run check` or `tsc`) passes without type errors.
- [ ] No manual duplication of types between Go and TS.

**Estimated Time:** 1.5h  
**Dependencies:** 21  
**Layer:** openapi

---

## Phase 6: Tests

### Task 23 — Unit tests: repositories
**Description:**
Write unit tests for `PillarRepo`, `CompetencyRepo`, `ScaleRepo`, `CatalogRepo`, and `AcceptanceRepo`. Use `enttest` (SQLite in-memory) or mocked clients. Cover pagination, eager loading, and transaction rollback.

**Acceptance Criteria:**
- [ ] Every repository method has at least one happy-path test.
- [ ] Cursor pagination tested for correct `next_cursor` and `has_more`.
- [ ] Transaction rollback is verified on forced errors.
- [ ] Eager loading verified (assert no N+1 via query logging or count).

**Estimated Time:** 1.5h  
**Dependencies:** 4, 5, 6, 7, 8, 9  
**Layer:** test

---

### Task 24 — Unit tests: services (cascade, bulk, upsert)
**Description:**
Write unit tests for service-layer business rules: cascade delete logic, advisory lock behavior, bulk criteria replacement, version increment, and acceptance-level upsert create-vs-update paths.

**Acceptance Criteria:**
- [ ] `PillarService.Delete` tested with `force=true` and `force=false`.
- [ ] `CompetencyService.Update` returns `CONCURRENT_UPDATE` on stale `If-Match`.
- [ ] `ScaleService.Upsert` increments version correctly after bulk replace.
- [ ] `AcceptanceService.Upsert` creates when absent and updates when present.

**Estimated Time:** 1.5h  
**Dependencies:** 10, 11, 12, 13, 14  
**Layer:** test

---

### Task 25 — Unit tests: handlers
**Description:**
Write table-driven HTTP tests for all 16 endpoints using `httptest`. Cover validation failures, not-found scenarios, conflicts, and query-param edge cases.

**Acceptance Criteria:**
- [ ] Each handler has table-driven tests covering happy path and error paths.
- [ ] HTTP status codes match the design error mapping exactly.
- [ ] Invalid `cursor`/`limit` returns 400 `INVALID_CURSOR` / `INVALID_PARAMETER`.
- [ ] Missing `If-Match` on PUT returns 428 `PRECONDITION_REQUIRED`.

**Estimated Time:** 2h  
**Dependencies:** 16, 17, 18, 19  
**Layer:** test

---

### Task 26 — Integration tests with testcontainers
**Description:**
Spin up a PostgreSQL container via testcontainers. Run end-to-end CRUD flows through the real HTTP router. Verify cascade delete side effects in the database and acceptance-level upsert integrity.

**Acceptance Criteria:**
- [ ] PostgreSQL testcontainer starts before the test suite.
- [ ] Full flow tested: create pillar → create competency → upsert criteria → delete pillar with cascade.
- [ ] Post-delete DB query confirms orphaned competencies/criteria/acceptance levels are gone.
- [ ] Acceptance-level upsert flow tested end-to-end.

**Estimated Time:** 2h  
**Dependencies:** 20, 25  
**Layer:** test

---

### Task 27 — Concurrency test (50 goroutines on criteria)
**Description:**
Launch 50 goroutines concurrently calling `UpsertScaleCriteria` against the same competency. Assert no lost updates, correct final version, and no internal server errors.

**Acceptance Criteria:**
- [ ] 50 concurrent upserts complete without panics or 500s.
- [ ] Final competency version equals initial version + 50.
- [ ] No duplicate or orphaned criteria remain in the database.
- [ ] All goroutines receive 200 or 409 (never 500).

**Estimated Time:** 1h  
**Dependencies:** 26  
**Layer:** test

---

### Task 28 — Cache invalidation test
**Description:**
Populate the Redis competency tree cache with a GET request, mutate the catalog via PUT/POST/DELETE, and assert that subsequent GETs do not return stale data. Also verify ETag behavior for static catalogs.

**Acceptance Criteria:**
- [ ] Redis key `competency:tree:v1` exists after the first GET.
- [ ] Write operation invalidates the cache key.
- [ ] Next GET repopulates the cache with fresh data.
- [ ] ETag for `/levels` and `/profiles` changes after mutations (if applicable) or returns 304 on repeat requests.

**Estimated Time:** 1h  
**Dependencies:** 26  
**Layer:** test

---

### Task 29 — Load test (500 req/s)
**Description:**
Write and run a load test script (k6 or vegeta) against `GET /api/v1/pillars` at 500 req/s for 60 seconds. Assert p95 latency < 200 ms and error rate < 0.1%.

**Acceptance Criteria:**
- [ ] Load test script committed under `tests/load/`.
- [ ] p95 latency remains under 200 ms for the full 60s run.
- [ ] Error rate is < 0.1% (no 500s).
- [ ] Prometheus `pgx_pool_*` metrics are collected and visible during the run.

**Estimated Time:** 1.5h  
**Dependencies:** 26  
**Layer:** test

---

## Phase 7: Polish

### Task 30 — Documentation
**Description:**
Write `README.md` files in `internal/service/competency/` and `internal/handler/competency/` explaining the cascade flow, locking strategy, middleware usage, and how to run tests. Add a short README in `internal/repository/competency/` for eager-loading and read-replica patterns.

**Acceptance Criteria:**
- [ ] Service README explains tx boundaries, cascade logic, and advisory locks.
- [ ] Handler README summarizes endpoints and `TODO(auth:C7)` markers.
- [ ] Repository README explains eager loading and read-replica selection.
- [ ] Docs are reviewed for accuracy against the implementation.

**Estimated Time:** 1h  
**Dependencies:** 20  
**Layer:** docs

---

### Task 31 — Prometheus metrics
**Description:**
Expose `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, and `pgx_pool_wait_duration_ms`. Add application-level counters/histograms (`requests_total`, `request_duration_seconds`) with method/path/status labels. Wire `/metrics` endpoint.

**Acceptance Criteria:**
- [ ] `/metrics` returns pgx pool metrics in Prometheus format.
- [ ] `requests_total` counter includes `method`, `path`, and `status` labels.
- [ ] `request_duration_seconds` histogram covers handler latency.
- [ ] Metrics are verifiable via an integration test hit to `/metrics`.

**Estimated Time:** 1h  
**Dependencies:** 20  
**Layer:** observability

---

*Generated for OpenSpec change C3 — competency-framework-api.*
