# Tasks: C4 — goals-api

## Summary

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| Phase 1: Infrastructure | 3 | ~3h |
| Phase 2: Repository Layer | 6 | ~10h |
| Phase 3: Service Layer | 10 | ~14h |
| Phase 4: Handler Layer | 7 | ~8h |
| Phase 5: OpenAPI | 2 | ~3h |
| Phase 6: Tests | 9 | ~12h |
| Phase 7: Polish | 2 | ~2h |
| **Total** | **39** | **~52h** |

---

## Phase 1: Infrastructure (~3h)

### Task 1: Package scaffolding
- **ID:** 1
- **Title:** Create package directories for goals bounded context
- **Description:** Scaffold the directory structure under `api/internal/` for the goals API: `handler/goal/`, `service/goal/`, `repository/goal/`, `pkg/validation/`, `pkg/batch/`. Add empty `.go` files with package declarations so Go tooling recognizes the packages. Ensure no circular imports with other bounded contexts.
- **Acceptance Criteria:**
  - [ ] All directories exist and are committed to version control.
  - [ ] `go build` passes for the new packages.
  - [ ] `go vet` and `gofmt` report no issues.
  - [ ] Directory structure matches the design.md diagram.
- **Estimated Time:** 1h
- **Dependencies:** C1 (data-model-core schema must exist)
- **Layer:** Infrastructure

### Task 2: Weight validation helper
- **ID:** 2
- **Title:** Implement `pkg/validation/weighting.go`
- **Description:** Build a small, stdlib-only package for Double 100% weight math. Expose `const Epsilon = 0.01`, `func WithinEpsilon(a, b float64) bool`, and `func SumValid(items []float64, expected float64) (sum float64, valid bool)`. Cover nil/empty slice semantics (sum = 0, valid = false if expected > 0).
- **Acceptance Criteria:**
  - [ ] `WithinEpsilon(100.0, 100.0)` returns true.
  - [ ] `WithinEpsilon(99.99, 100.0)` returns true.
  - [ ] `WithinEpsilon(100.01, 100.0)` returns false.
  - [ ] `SumValid(nil, 100.0)` returns `(0, false)`.
  - [ ] Unit tests in `pkg/validation/weighting_test.go` achieve 100% coverage.
- **Estimated Time:** 1h
- **Dependencies:** Task 1
- **Layer:** Infrastructure

### Task 3: Batch helper
- **ID:** 3
- **Title:** Implement `pkg/batch/batch.go`
- **Description:** Create a generic batch operation wrapper: `type Batch[T any] struct { Items []T }` with `MaxSize` constant set to 50, and a helper `func ValidateSize(items []T) error` that returns `ErrBatchSizeExceeded` if len(items) > 50. Keep the package pure stdlib; no DB dependencies.
- **Acceptance Criteria:**
  - [ ] `ValidateSize` returns nil for 50 items.
  - [ ] `ValidateSize` returns `ErrBatchSizeExceeded` for 51 items.
  - [ ] Error message includes the actual and maximum size.
  - [ ] Unit tests cover boundary values (0, 1, 50, 51, 100).
- **Estimated Time:** 1h
- **Dependencies:** Task 1
- **Layer:** Infrastructure

---

## Phase 2: Repository Layer (~10h)

### Task 4: GoalCategory repository (CRUD + weight queries)
- **ID:** 4
- **Title:** Implement GoalCategory repository
- **Description:** Implement `goal_repo.go` methods for categories: `ListCategoriesByEmployee`, `CreateCategory`, `UpdateCategory`, `DeleteCategory`, `LockCategory` (SELECT FOR UPDATE), `SumCategoryWeightsByEmployee`. Ensure uniqueness of category name per employee at the Ent schema level. Return domain-specific errors for not-found and duplicate name.
- **Acceptance Criteria:**
  - [ ] All category CRUD methods are implemented and compile.
  - [ ] `LockCategory` uses `ForUpdate()` to serialize category weight-sum checks.
  - [ ] `SumCategoryWeightsByEmployee` returns accurate sum via Ent aggregate.
  - [ ] Duplicate category name returns `ErrDuplicateCategoryName`.
  - [ ] Unit tests mock the Ent client and assert query behavior.
- **Estimated Time:** 2h
- **Dependencies:** Task 1, C1
- **Layer:** Repository

### Task 5: Goal repository (CRUD + state transitions)
- **ID:** 5
- **Title:** Implement Goal repository
- **Description:** Implement `CreateGoal`, `UpdateGoal`, `DeleteGoal`, `UpdateGoalCurrentValue`, `SumGoalWeightsByCategory`, `ListGoalsByCategory`. `UpdateGoal` must include `WHERE version = req.Version` for optimistic locking. `UpdateGoalCurrentValue` only updates the `currentValue` field and increments `version`. `DeleteGoal` must cascade-delete `GoalKpiLink` rows.
- **Acceptance Criteria:**
  - [ ] `CreateGoal` inserts a new goal linked to the correct category.
  - [ ] `UpdateGoal` increments `version` and returns `ErrConcurrentModification` if no rows match.
  - [ ] `UpdateGoalCurrentValue` updates only `currentValue` and `version`.
  - [ ] `DeleteGoal` removes the goal and all associated `GoalKpiLink` rows.
  - [ ] `SumGoalWeightsByCategory` returns exact sum via Ent aggregate.
- **Estimated Time:** 2h
- **Dependencies:** Task 4
- **Layer:** Repository

### Task 6: KPI repository (CRUD)
- **ID:** 6
- **Title:** Implement KPI repository
- **Description:** Implement `ListKPIs`, `CreateKPI`, `UpdateKPI`, `DeleteKPI`. `ListKPIs` must support cursor-based pagination (limit, next cursor). `DeleteKPI` must reject if `CountGoalLinksByKPI` returns > 0. Return `ErrKpiLinkedCannotDelete` when linked.
- **Acceptance Criteria:**
  - [ ] `ListKPIs` returns paginated results with `nextCursor`.
  - [ ] `CreateKPI` and `UpdateKPI` return the created/updated KPI.
  - [ ] `DeleteKPI` rejects with `ErrKpiLinkedCannotDelete` when linked.
  - [ ] `CountGoalLinksByKPI` returns correct link count.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 1, C1
- **Layer:** Repository

### Task 7: GoalKpiLink repository (link/unlink)
- **ID:** 7
- **Title:** Implement GoalKpiLink repository
- **Description:** Implement `LinkKPI`, `UnlinkKPI`, `ReplaceGoalKpiLinks`. `LinkKPI` must be idempotent via `OnConflict().DoNothing()`. `ReplaceGoalKpiLinks` deletes all existing links for the goal and re-inserts the provided KPI IDs atomically. Validate that the number of KPI IDs does not exceed 5.
- **Acceptance Criteria:**
  - [ ] `LinkKPI` is idempotent (duplicate call returns no error).
  - [ ] `ReplaceGoalKpiLinks` atomically swaps links.
  - [ ] Attempting to replace with > 5 KPIs returns `ErrKpiLinkLimitExceeded`.
  - [ ] `UnlinkKPI` removes the specific link without affecting others.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 5, Task 6
- **Layer:** Repository

### Task 8: GoalAssignment repository (create/get per cycle)
- **ID:** 8
- **Title:** Implement GoalAssignment repository
- **Description:** Implement `GetAssignment` and `CreateAssignment`. `CreateAssignment` must use a PostgreSQL advisory lock (`pg_advisory_lock`) keyed by `hashtext(employee_id || cycle_id)` to prevent duplicate assignments under concurrency. After acquiring the lock, check for an existing assignment for the same employee+cycle; if found, return the existing one (idempotent). Return `GoalAssignment` with nested categories and goals.
- **Acceptance Criteria:**
  - [ ] `GetAssignment` returns the assignment for the employee's active cycle.
  - [ ] `CreateAssignment` uses advisory lock and is idempotent.
  - [ ] Concurrent duplicate creation attempts result in a single row.
  - [ ] Advisory lock is released after the transaction commits or rolls back.
- **Estimated Time:** 2h
- **Dependencies:** Task 4, Task 5
- **Layer:** Repository

### Task 9: Weight validation queries (SUM aggregates)
- **ID:** 9
- **Title:** Implement aggregate weight queries
- **Description:** Implement `SumGoalWeightsByCategoryID` and `SumCategoryWeightsByEmployee` in the repository. Use Ent aggregate functions (`ent.Sum`) with `Scan` to compute the sums. Ensure nil sums are handled gracefully (return 0.0). These queries are used by both the service layer and the weight validation endpoint.
- **Acceptance Criteria:**
  - [ ] `SumGoalWeightsByCategoryID` returns 0.0 for an empty category.
  - [ ] `SumCategoryWeightsByEmployee` returns 0.0 for an employee with no categories.
  - [ ] Both queries return accurate float values even with large numbers of rows.
  - [ ] Unit tests cover empty and non-empty cases.
- **Estimated Time:** 1h
- **Dependencies:** Task 4, Task 5
- **Layer:** Repository

---

## Phase 3: Service Layer (~14h)

### Task 10: GoalCategory service — Create/Update/Delete
- **ID:** 10
- **Title:** Implement category service methods
- **Description:** Implement `CreateCategory`, `UpdateCategory`, `DeleteCategory` in `goal_service.go`. Each method must call `enforcePhase` with `phaseAsignacion` (and `phaseAvance` for weight-only updates). `UpdateCategory` must reject non-weight changes in `avance`. `DeleteCategory` must reject if the category contains goals (or cascade delete per business rule, to be defined in spec). Return domain-specific errors.
- **Acceptance Criteria:**
  - [ ] `CreateCategory` is gated to `asignacion` phase.
  - [ ] `UpdateCategory` in `avance` rejects name/description changes.
  - [ ] `DeleteCategory` rejects in `avance` and `cierre`.
  - [ ] Duplicate name per employee returns `ErrDuplicateCategoryName`.
  - [ ] Unit tests cover all phase and validation paths.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 4
- **Layer:** Service

### Task 11: Goal service — Create with weight overflow check
- **ID:** 11
- **Title:** Implement goal creation with weight overflow prevention
- **Description:** Implement `CreateGoal` in the service layer. Acquire the parent category via `LockCategory` inside a transaction, compute the current sum of goal weights, and reject the creation if `sum + req.Weight > 100.0 + epsilon`. Validate `targetValue > 0` and `unit` enum. Validate that all `kpiIds` exist (query KPI repo). On success, return the created goal.
- **Acceptance Criteria:**
  - [ ] Goal creation is rejected with `ErrWeightSumInvalid` if weight exceeds 100%.
  - [ ] `targetValue <= 0` returns `ErrInvalidTargetValue`.
  - [ ] Invalid `unit` returns `ErrInvalidUnit`.
  - [ ] Non-existent `kpiId` returns `ErrKpiNotFound`.
  - [ ] Unit tests cover overflow edge cases (99.99%, 100.0%, 100.01%).
- **Estimated Time:** 1.5h
- **Dependencies:** Task 5, Task 9
- **Layer:** Service

### Task 12: Goal service — Update with optimistic locking
- **ID:** 12
- **Title:** Implement goal update with optimistic locking
- **Description:** Implement `UpdateGoal` in the service layer. Validate the `version` field and reject with `ErrConcurrentModification` if the row was updated between read and write. In `avance` phase, reject updates to `weight` and `targetValue`. Strip disallowed fields before calling the repository. Update KPI links via `ReplaceGoalKpiLinks` if `kpiIds` is provided.
- **Acceptance Criteria:**
  - [ ] Version mismatch returns `ErrConcurrentModification`.
  - [ ] `weight` or `targetValue` update in `avance` returns `ErrPhaseRestricted`.
  - [ ] `name`/`description`/`unit` updates in `avance` are allowed.
  - [ ] KPI links are atomically replaced during update.
  - [ ] Unit tests simulate concurrent update races.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 5, Task 7
- **Layer:** Service

### Task 13: Goal service — Delete (phase-gated)
- **ID:** 13
- **Title:** Implement goal deletion with phase enforcement
- **Description:** Implement `DeleteGoal` in the service layer. Reject deletion in `avance` and `cierre` with `ErrPhaseRestricted` (or `ErrGoalNotDeletableInPhase` if more specific). Verify the goal belongs to the employee before deleting. Cascade-delete KPI links via repository.
- **Acceptance Criteria:**
  - [ ] Deletion in `avance` returns `ErrPhaseRestricted`.
  - [ ] Deletion in `asignacion` succeeds if the goal exists.
  - [ ] Deleting a non-existent goal returns `ErrGoalNotFound`.
  - [ ] KPI links are removed along with the goal.
  - [ ] Unit tests cover all phase and ownership checks.
- **Estimated Time:** 1h
- **Dependencies:** Task 5
- **Layer:** Service

### Task 14: Goal service — UpdateProgress (phase-gated)
- **ID:** 14
- **Title:** Implement progress update
- **Description:** Implement `UpdateGoalProgress` in the service layer. Gate to `avance` phase only. Validate `currentValue >= 0`. Call `UpdateGoalCurrentValue` repository method. Return the updated goal. No weight validation is performed here.
- **Acceptance Criteria:**
  - [ ] Update in `asignacion` returns `ErrPhaseRestricted`.
  - [ ] Negative `currentValue` returns `ErrInvalidCurrentValue`.
  - [ ] Update in `avance` succeeds and returns the goal.
  - [ ] `version` is incremented by the repository.
  - [ ] Unit tests cover phase and validation paths.
- **Estimated Time:** 1h
- **Dependencies:** Task 5
- **Layer:** Service

### Task 15: KPI service — CRUD
- **ID:** 15
- **Title:** Implement KPI service methods
- **Description:** Implement `ListKPIs`, `CreateKPI`, `UpdateKPI`, `DeleteKPI` in `kpi_service.go`. `ListKPIs` must support cursor pagination. `DeleteKPI` must reject if the KPI is linked to any goals. No phase enforcement for KPI catalog (RBAC handled by C7).
- **Acceptance Criteria:**
  - [ ] `ListKPIs` returns paginated KPI list with `nextCursor`.
  - [ ] `CreateKPI` and `UpdateKPI` return the KPI.
  - [ ] `DeleteKPI` rejects linked KPIs with `ErrKpiLinkedCannotDelete`.
  - [ ] Unit tests cover all CRUD paths.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 6
- **Layer:** Service

### Task 16: KPI linking service — Link/Unlink
- **ID:** 16
- **Title:** Implement KPI link/unlink service
- **Description:** Implement `LinkKPI` and `UnlinkKPI` in the service layer. Both are gated to `asignacion` phase. `LinkKPI` must validate the goal exists and belongs to the employee, and that the KPI exists. `UnlinkKPI` must validate the link exists. Enforce the max 5 KPIs per goal limit.
- **Acceptance Criteria:**
  - [ ] Linking in `avance` returns `ErrPhaseRestricted`.
  - [ ] Linking a 6th KPI returns `ErrKpiLinkLimitExceeded`.
  - [ ] Linking non-existent goal/KPI returns `ErrGoalNotFound`/`ErrKpiNotFound`.
  - [ ] Unlinking non-existent link returns `ErrKpiNotFound`.
  - [ ] Unit tests cover phase, limit, and not-found cases.
- **Estimated Time:** 1h
- **Dependencies:** Task 7, Task 15
- **Layer:** Service

### Task 17: Weight validation service — ValidateDoubleWeighting
- **ID:** 17
- **Title:** Implement double 100% weight validation
- **Description:** Implement `ValidateDoubleWeighting` in the service layer. Query all categories for the employee, sum their weights. For each category, query all goals and sum their weights. Use `epsilon = 0.01` from `pkg/validation`. Return a `WeightValidationResult` with `valid`, `categorySum`, `goalSums`, and `deficit` values. This is a read-only operation; no phase enforcement.
- **Acceptance Criteria:**
  - [ ] Returns `valid = true` when both levels sum to 100%.
  - [ ] Returns `valid = false` when category sum is 99.99% or 100.01%.
  - [ ] Returns `valid = false` when any goal sum is 99.99% or 100.01%.
  - [ ] Empty employee (no categories) returns `valid = false`.
  - [ ] Unit tests cover all edge cases.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 2, Task 9
- **Layer:** Service

### Task 18: Batch service — BatchCreateUpdateGoals
- **ID:** 18
- **Title:** Implement atomic batch create/update
- **Description:** Implement `BatchCreateUpdateGoals` in the service layer. Validate batch size <= 50 using `pkg/batch`. Gate to `asignacion` phase. Run the entire batch inside a single transaction via `WithTx`. For each item, call `CreateGoal` or `UpdateGoal`. After all items, run `validateWeightInvariantTx` to ensure the final projected state satisfies Double 100%. If any step fails, the entire transaction rolls back.
- **Acceptance Criteria:**
  - [ ] Batch of 50 items completes in <500ms in CI.
  - [ ] Batch > 50 items returns `ErrBatchSizeExceeded`.
  - [ ] Single failure in the batch rolls back all changes.
  - [ ] Post-batch weight validation rejects invalid final state.
  - [ ] Unit tests verify atomicity and weight validation.
- **Estimated Time:** 2h
- **Dependencies:** Task 3, Task 11, Task 12
- **Layer:** Service

### Task 19: Phase enforcement middleware/helper
- **ID:** 19
- **Title:** Implement phase enforcement helper
- **Description:** Implement `enforcePhase` in the service layer. It queries C2 (`evaluation-lifecycle-api`) via `PhaseChecker` interface to get the current phase for the employee. Compare against the allowed phases. Return `ErrPhaseRestricted` if the current phase is not in the allowed set. C2 integration can be mocked for unit testing.
- **Acceptance Criteria:**
  - [ ] `enforcePhase` returns nil when phase is allowed.
  - [ ] `enforcePhase` returns `ErrPhaseRestricted` when phase is disallowed.
  - [ ] C2 lookup failure is propagated as a service error.
  - [ ] Unit tests mock `PhaseChecker` and cover all phase combinations.
  - [ ] Integration test verifies real C2 call (or stub) behavior.
- **Estimated Time:** 1.5h
- **Dependencies:** C2
- **Layer:** Service

---

## Phase 4: Handler Layer (~8h)

### Task 20: Category handler — CRUD endpoints
- **ID:** 20
- **Title:** Implement category HTTP handler
- **Description:** Implement `category_handler.go` with `ListCategories`, `CreateCategory`, `UpdateCategory`, `DeleteCategory` handlers. Parse `empId` from Chi URL params. Decode JSON requests, validate basic constraints, and call the service layer. Return proper HTTP status codes (200, 201, 204, 400, 403, 409). Wire ETag header for list response.
- **Acceptance Criteria:**
  - [ ] `GET /api/v1/employees/{empId}/categories` returns `200` with `CategoryListResponse`.
  - [ ] `POST` returns `201` with `CategoryResponse`.
  - [ ] `PUT` returns `200` with `CategoryResponse`.
  - [ ] `DELETE` returns `204`.
  - [ ] Duplicate name returns `409` with `DUPLICATE_CATEGORY_NAME`.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 10
- **Layer:** Handler

### Task 21: Goal handler — CRUD + progress endpoints
- **ID:** 21
- **Title:** Implement goal HTTP handler
- **Description:** Implement `goal_handler.go` with `CreateGoal`, `UpdateGoal`, `DeleteGoal`, `UpdateGoalProgress` handlers. Parse `empId` and `catId`/`goalId` from URL params. Decode requests and call service layer. Map service errors to HTTP status codes: `ErrPhaseRestricted` → `403`, `ErrConcurrentModification` → `409`, `ErrWeightSumInvalid` → `422`, `ErrGoalNotFound` → `404`.
- **Acceptance Criteria:**
  - [ ] `POST /api/v1/employees/{empId}/categories/{catId}/goals` returns `201`.
  - [ ] `PUT /api/v1/goals/{goalId}` returns `200`.
  - [ ] `DELETE /api/v1/goals/{goalId}` returns `204`.
  - [ ] `PATCH /api/v1/goals/{goalId}/progress` returns `200`.
  - [ ] Version mismatch returns `409` with `CONCURRENT_MODIFICATION`.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 11, Task 12, Task 13, Task 14
- **Layer:** Handler

### Task 22: KPI handler — CRUD + linking endpoints
- **ID:** 22
- **Title:** Implement KPI HTTP handler
- **Description:** Implement `kpi_handler.go` with `ListKPIs`, `CreateKPI`, `UpdateKPI`, `DeleteKPI` handlers, and `LinkKPI` / `UnlinkKPI` handlers in the same file or a separate linking handler. Return proper HTTP status codes. `ListKPIs` must support `cursor` and `limit` query params.
- **Acceptance Criteria:**
  - [ ] `GET /api/v1/kpis` returns `200` with `KpiListResponse`.
  - [ ] `POST /api/v1/kpis` returns `201`.
  - [ ] `DELETE /api/v1/kpis/{kpiId}` returns `409` if linked.
  - [ ] `POST /api/v1/goals/{goalId}/kpis` returns `200`.
  - [ ] `DELETE /api/v1/goals/{goalId}/kpis/{kpiId}` returns `204`.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 15, Task 16
- **Layer:** Handler

### Task 23: Validation handler — validate-weights endpoint
- **ID:** 23
- **Title:** Implement weight validation HTTP handler
- **Description:** Implement `weight_handler.go` with `ValidateWeights` handler. `POST /api/v1/employees/{empId}/validate-weights` calls `ValidateDoubleWeighting` and returns `WeightValidationResponse`. No request body. This endpoint is read-only and allowed in all phases.
- **Acceptance Criteria:**
  - [ ] Returns `200` with `WeightValidationResponse`.
  - [ ] `valid` field reflects actual Double 100% check.
  - [ ] `details` include category sums and per-category goal sums.
  - [ ] Works in all phases (`asignacion`, `avance`, `cierre`).
- **Estimated Time:** 1h
- **Dependencies:** Task 17
- **Layer:** Handler

### Task 24: Assignment handler — assignment endpoints
- **ID:** 24
- **Title:** Implement assignment HTTP handler
- **Description:** Implement `assignment_handler.go` with `GetAssignment` and `CreateAssignment` handlers. `GET /api/v1/employees/{empId}/assignments` returns the current assignment. `POST` creates an assignment gated to `asignacion` phase. Return `AssignmentResponse` with nested categories and goals.
- **Acceptance Criteria:**
  - [ ] `GET` returns `200` with `AssignmentResponse`.
  - [ ] `POST` returns `201` with `AssignmentResponse`.
  - [ ] `POST` in `avance` returns `403`.
  - [ ] Duplicate creation returns the existing assignment (idempotent).
- **Estimated Time:** 1h
- **Dependencies:** Task 8
- **Layer:** Handler

### Task 25: Batch handler — batch endpoint
- **ID:** 25
- **Title:** Implement batch HTTP handler
- **Description:** Implement `batch_handler.go` (or add to `goal_handler.go`) with `BatchGoals` handler. `POST /api/v1/goals/batch` decodes `BatchGoalRequest`, validates size <= 50, and calls `BatchCreateUpdateGoals`. Return `BatchGoalResponse` with the resulting goals. Map errors to appropriate HTTP status codes.
- **Acceptance Criteria:**
  - [ ] `POST` returns `200` with `BatchGoalResponse`.
  - [ ] Batch > 50 returns `400` with `BATCH_SIZE_EXCEEDED`.
  - [ ] Weight failure in batch returns `422` with `WEIGHT_SUM_INVALID`.
  - [ ] Phase failure returns `403`.
- **Estimated Time:** 1h
- **Dependencies:** Task 18
- **Layer:** Handler

### Task 26: Route registration + phase middleware
- **ID:** 26
- **Title:** Register Chi routes and wire middleware
- **Description:** Implement `routes.go` in `handler/goal/`. Register all endpoints with Chi router. Group routes under `/api/v1`. Apply existing auth middleware (injected by C7) to all routes. Wire idempotency key middleware (if provided by C7 or a shared package). Add request logging and request ID injection. Ensure no handler leaks Ent or service implementation details.
- **Acceptance Criteria:**
  - [ ] All 18 endpoints are registered and reachable.
  - [ ] Auth middleware rejects unauthenticated requests with `401`.
  - [ ] Request ID is present in all response headers.
  - [ ] `go test` for the handler package passes.
  - [ ] Route registration is isolated and testable.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 20, Task 21, Task 22, Task 23, Task 24, Task 25
- **Layer:** Handler

---

## Phase 5: OpenAPI (~3h)

### Task 27: OpenAPI 3.1 spec
- **ID:** 27
- **Title:** Write OpenAPI 3.1 specification
- **Description:** Create the complete OpenAPI 3.1 YAML file for all goals-api endpoints. Define all request/response schemas, parameters, and error responses as documented in design.md. Include `PhaseRestricted`, `WeightInvalid`, `DuplicateCategoryName`, `ConcurrentModification`, `KpiLinked`, and `RateLimit` response components. Save to `api/openapi/goals-api.yaml` (or project-specific path).
- **Acceptance Criteria:**
  - [ ] All 18 endpoints are documented with correct methods, paths, and schemas.
  - [ ] All request/response schemas match the design.md definitions.
  - [ ] All error response components are defined.
  - [ ] `openapi-generator-cli validate` passes with zero errors.
  - [ ] Spec is reviewed and committed.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 26
- **Layer:** Contract

### Task 28: TypeScript type generation
- **ID:** 28
- **Title:** Generate TypeScript types from OpenAPI
- **Description:** Run `openapi-typescript` (or equivalent) against the OpenAPI spec to generate TypeScript interfaces in `web/src/lib/api/goals-api.ts` (or project-specific path). Ensure zero type errors when imported into the Svelte frontend. Add a `pnpm` script to regenerate types.
- **Acceptance Criteria:**
  - [ ] `pnpm run generate-api-types` (or equivalent) produces the TS file.
  - [ ] `pnpm run check` (or `tsc --noEmit`) passes with zero errors.
  - [ ] All request/response types are importable from the generated file.
  - [ ] CI validates that generated types are up-to-date.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 27
- **Layer:** Contract

---

## Phase 6: Tests (~12h)

### Task 29: Unit tests — repositories
- **ID:** 29
- **Title:** Write unit tests for all repositories
- **Description:** Write `goal_repo_test.go`, `kpi_repo_test.go`, and `assignment_repo_test.go` using mocked Ent clients (gomock/mockery). Cover all CRUD paths, aggregate queries, and transaction wrappers. Verify that `WithTx` correctly commits on success and rolls back on failure.
- **Acceptance Criteria:**
  - [ ] All repository methods have at least one unit test.
  - [ ] `WithTx` rollback on error is tested.
  - [ ] `LockCategory` is mocked and tested.
  - [ ] Aggregate queries (`SumGoalWeightsByCategory`, `SumCategoryWeightsByEmployee`) are tested.
  - [ ] Coverage for repository package >= 80%.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 4, Task 5, Task 6, Task 7, Task 8, Task 9
- **Layer:** Test

### Task 30: Unit tests — weight validation (edge cases)
- **ID:** 30
- **Title:** Write unit tests for weight validation edge cases
- **Description:** Write tests for `pkg/validation/weighting.go` and the service-layer `ValidateDoubleWeighting`. Test edge cases: 99.99% (should fail), 100.0% (pass), 100.01% (fail), empty categories (fail), single category/goal at 100% (pass), three categories at 33.33% each (pass with epsilon). Ensure float precision issues are handled.
- **Acceptance Criteria:**
  - [ ] 99.99% category sum returns `valid = false`.
  - [ ] 100.01% category sum returns `valid = false`.
  - [ ] 100.0% category sum returns `valid = true`.
  - [ ] Empty categories return `valid = false`.
  - [ ] Three items at 33.33% return `valid = true` (epsilon tolerance).
- **Estimated Time:** 1h
- **Dependencies:** Task 2, Task 17
- **Layer:** Test

### Task 31: Unit tests — services
- **ID:** 31
- **Title:** Write unit tests for all service methods
- **Description:** Write `goal_service_test.go` and `kpi_service_test.go` using mocked repositories and `PhaseChecker`. Cover all phase gating paths, weight validation, optimistic locking, batch processing, and error mapping. Mock C2 phase checker to return `asignacion`, `avance`, and `cierre`.
- **Acceptance Criteria:**
  - [ ] All service methods have at least one unit test.
  - [ ] Phase gating is tested for every mutating operation.
  - [ ] Weight overflow prevention is tested.
  - [ ] Optimistic locking failure is tested.
  - [ ] Coverage for service package >= 80%.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 10, Task 11, Task 12, Task 13, Task 14, Task 15, Task 16, Task 17, Task 18, Task 19
- **Layer:** Test

### Task 32: Unit tests — handlers
- **ID:** 32
- **Title:** Write unit tests for all HTTP handlers
- **Description:** Write handler tests using `httptest` and mocked services. Cover request decoding, response encoding, error code mapping, and middleware interaction. Test that `empId` and `goalId` are correctly extracted from Chi URL params. Verify ETag generation for list endpoints.
- **Acceptance Criteria:**
  - [ ] All handlers have at least one unit test.
  - [ ] `400` is returned for invalid JSON.
  - [ ] `404` is returned for missing resources.
  - [ ] `409` is returned for concurrent modification.
  - [ ] `403` is returned for phase restriction.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 20, Task 21, Task 22, Task 23, Task 24, Task 25
- **Layer:** Test

### Task 33: Integration tests
- **ID:** 33
- **Title:** Write integration tests with PostgreSQL
- **Description:** Write integration tests using `testcontainers` with PostgreSQL 15+. Spin up a real DB, run Ent migrations, and test the full stack: handler → service → repository → DB. Cover category CRUD, goal CRUD, KPI linking, and assignment creation.
- **Acceptance Criteria:**
  - [ ] Integration tests run in CI with `go test -tags=integration`.
  - [ ] A real PostgreSQL container is created and cleaned up per test.
  - [ ] All major write operations are exercised end-to-end.
  - [ ] Tests run under `go test -race` with no races detected.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 29, Task 31
- **Layer:** Test

### Task 34: Phase restriction tests
- **ID:** 34
- **Title:** Write phase restriction integration tests
- **Description:** Write integration tests that verify phase restrictions using a stubbed or real C2 phase provider. Test that writes in `avance` and `cierre` return `PHASE_RESTRICTED` (or `GOAL_NOT_DELETABLE_IN_PHASE`). Test that `PATCH /progress` is rejected in `asignacion`. Test that `UpdateCategory` in `avance` only allows weight changes.
- **Acceptance Criteria:**
  - [ ] Write in `avance` returns `403` for create, delete, and update (weight/target).
  - [ ] Write in `cierre` returns `403` for all mutating operations.
  - [ ] `PATCH /progress` in `asignacion` returns `403`.
  - [ ] `UpdateCategory` name change in `avance` returns `403`.
  - [ ] `UpdateCategory` weight change in `avance` returns `200` (if allowed by rules).
- **Estimated Time:** 1h
- **Dependencies:** Task 33
- **Layer:** Test

### Task 35: Concurrency test (50 goroutines)
- **ID:** 35
- **Title:** Write concurrency test for goal creation
- **Description:** Write a test that spawns 50 goroutines, each attempting to create a goal with weight=50 in the same category. Only 2 should succeed; the rest must receive `WEIGHT_SUM_INVALID` (or equivalent). Run under `go test -race` to ensure no race conditions. Use `sync.WaitGroup` to coordinate goroutines.
- **Acceptance Criteria:**
  - [ ] Exactly 2 goals succeed.
  - [ ] 48 requests fail with weight overflow error.
  - [ ] Final sum of goal weights in the category is exactly 100.0.
  - [ ] `go test -race` passes with no race warnings.
  - [ ] Test runs under `testcontainers` PostgreSQL.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 33
- **Layer:** Test

### Task 36: Batch atomicity test
- **Description:** Write a test that sends a batch of 3 creates + 2 updates, where the last update would violate the weight limit. Assert that the entire batch fails and no rows are modified in the database. Verify that the transaction is rolled back and the pre-batch state is intact.
- **Acceptance Criteria:**
  - [ ] Batch with a weight violation fails with `422 WEIGHT_SUM_INVALID`.
  - [ ] No goals are created or updated after the failed batch.
  - [ ] Pre-batch goal weights remain unchanged.
  - [ ] Test is repeatable and runs under `testcontainers`.
- **Estimated Time:** 1h
- **Dependencies:** Task 33
- **Layer:** Test

### Task 37: Load test (200 req/s reads, 50 req/s writes)
- **ID:** 37
- **Title:** Write and run load tests
- **Description:** Write a load test using `k6` or a Go benchmark script. Target 200 req/s sustained reads and 50 req/s writes for 5 minutes. Measure p50, p95, p99 latencies. Assert p95 < 300ms and 0% weight validation errors due to race conditions. Run against a staging environment or a local PostgreSQL with connection pool tuned to `MaxConns=30`.
- **Acceptance Criteria:**
  - [ ] 200 req/s read throughput sustained for 5 minutes.
  - [ ] 50 req/s write throughput sustained for 5 minutes.
  - [ ] p95 latency < 300ms.
  - [ ] 0% weight validation errors due to race conditions.
  - [ ] Load test script is committed and documented.
- **Estimated Time:** 1.5h
- **Dependencies:** Task 33
- **Layer:** Test

### Task 38: OpenAPI validation test
- **ID:** 38
- **Title:** Validate OpenAPI spec and generated types in CI
- **Description:** Add a CI step that runs `openapi-generator-cli validate` against the spec. Add a step that runs `openapi-typescript` and `tsc --noEmit` to verify generated TypeScript types compile cleanly. Fail the build if the spec is invalid or types are stale.
- **Acceptance Criteria:**
  - [ ] CI step validates OpenAPI spec with zero errors.
  - [ ] CI step generates TypeScript types and compiles them.
  - [ ] CI fails if the spec is invalid.
  - [ ] CI fails if generated types are stale (checksum comparison).
- **Estimated Time:** 1h
- **Dependencies:** Task 27, Task 28
- **Layer:** Test

---

## Phase 7: Polish (~2h)

### Task 39: Documentation
- **ID:** 39
- **Title:** Write developer documentation
- **Description:** Write a README in `api/internal/handler/goal/` or `docs/goals-api.md` explaining the package structure, how to run tests, how to regenerate OpenAPI types, and the phase enforcement matrix. Include a mermaid diagram of the request flow (handler → service → repository → DB). Document the idempotency key and rate limit behavior.
- **Acceptance Criteria:**
  - [ ] README explains the 3-layer architecture (handler/service/repository).
  - [ ] Phase enforcement matrix is documented in a table.
  - [ ] Instructions for running unit and integration tests are present.
  - [ ] Instructions for regenerating OpenAPI types are present.
  - [ ] Document is reviewed and committed.
- **Estimated Time:** 1h
- **Dependencies:** Task 38
- **Layer:** Documentation

### Task 40: Prometheus metrics
- **ID:** 40
- **Title:** Instrument goals API with Prometheus metrics
- **Description:** Add Prometheus counters and histograms for all goals-api endpoints: `goals_api_requests_total` (labeled by method, endpoint, status), `goals_api_request_duration_seconds` (labeled by endpoint), `goals_api_weight_validation_failures_total`, `goals_api_phase_restrictions_total`, and `goals_api_concurrent_modifications_total`. Ensure metrics are registered and exposed on the standard metrics endpoint (provided by shared observability package).
- **Acceptance Criteria:**
  - [ ] All endpoints emit request count and duration metrics.
  - [ ] Weight validation failures are counted.
  - [ ] Phase restrictions are counted.
  - [ ] Concurrent modifications are counted.
  - [ ] Metrics are visible on `/metrics` and tested in CI.
- **Estimated Time:** 1h
- **Dependencies:** Task 26
- **Layer:** Observability

---

## Notes

- **Order of execution:** Follow phase order. Tasks within a phase can be parallelized where dependencies allow.
- **CI gating:** No task in Phase 4+ should be considered complete until its corresponding unit tests in Phase 6 pass in CI.
- **Dependencies on C1/C2/C7:** These are external changes. If they are not ready, mock their interfaces and write TODOs for integration points.
