# Tasks: org-hierarchy-api (C5)

## Summary

| Phase | Tasks | Est. Hours |
|-------|-------|------------|
| Phase 1: Infrastructure | 1–4 | 4h |
| Phase 2: Repository Layer | 5–9 | 10h |
| Phase 3: Service Layer | 10–17 | 12h |
| Phase 4: Handler Layer | 18–22 | 8h |
| Phase 5: OpenAPI | 23–24 | 3h |
| Phase 6: Tests | 25–32 | 10h |
| Phase 7: Polish | 33–34 | 2h |
| **Total** | **34 tasks** | **~49h** |

---

## Phase 1: Infrastructure (~4h)

### T1. Package scaffolding

**Description:** Create the complete directory layout and empty Go files for the org hierarchy bounded context under `api/internal/` and `api/pkg/`. Add empty interfaces, constructor stubs, and wiring placeholders so later tasks can compile incrementally.

**Acceptance Criteria:**
- [ ] `api/internal/handler/org/`, `api/internal/service/org/`, `api/internal/repository/org/`, `api/pkg/tree/`, `api/pkg/search/`, `api/pkg/cache/` directories exist
- [ ] Every file from the design doc package structure exists with package declaration and minimal interface stubs
- [ ] `go build ./api/...` passes with no errors (stubs compile)
- [ ] `go vet ./api/...` passes

**Estimated Time:** 1h
**Dependencies:** None
**Layer:** Infra

---

### T2. ltree extension setup + migration

**Description:** Add a versioned Ent migration that enables the PostgreSQL `ltree` extension, adds `path` (ltree) and `version` (int) columns to `org_nodes`, and creates GIST and btree indexes. Include a one-time backfill CTE for existing nodes and the `idx_orgnode_path_gist`, `idx_orgnode_path_btree` indexes.

**Acceptance Criteria:**
- [ ] Migration file runs successfully on a fresh PostgreSQL container
- [ ] `CREATE EXTENSION IF NOT EXISTS ltree;` is present in its own migration step
- [ ] `org_nodes` has `path` (ltree, nullable initially) and `version` (int, default 0) columns
- [ ] GIST index on `path` and btree index on `path` are created
- [ ] Backfill CTE correctly populates `path` for existing rows using `parent_id` recursion

**Estimated Time:** 2h
**Dependencies:** T1
**Layer:** Infra

---

### T3. Tree traversal helpers (`pkg/tree/traversal.go`)

**Description:** Implement generic tree-traversal utilities: `Flatten(nodes) []Node` (parent-id based flat list), `ToNested(nodes) *Node` (builds children recursively), `FilterDepth(nodes, maxDepth) []Node`, and `PathString(nodeID, parentPath) string` for `ltree` path generation. Also include `ParsePath`, `IsDescendantOf`, and `IsAncestorOf` helpers.

**Acceptance Criteria:**
- [ ] `Flatten` converts a nested tree into a flat slice with correct `parentId` values
- [ ] `ToNested` builds a deeply nested tree from a flat slice ordered by path
- [ ] `FilterDepth` returns only nodes with `nlevel(path) <= maxDepth`
- [ ] `PathString` concatenates parent path and node ID correctly (e.g., `"1.2"` + `"3"` → `"1.2.3"`)
- [ ] All functions have table-driven unit tests with 80%+ coverage

**Estimated Time:** 1h
**Dependencies:** T1
**Layer:** Infra

---

### T4. Employee search helper (`pkg/search/employee_search.go`)

**Description:** Implement a reusable search builder that constructs a PostgreSQL `tsvector` query over `first_name || ' ' || last_name || ' ' || email`. Support plain queries, limit clamping, and optional trigram fallback if `pg_trgm` is available. Expose `BuildSearchPredicate(query string) sql.Predicate` for use in Ent repositories.

**Acceptance Criteria:**
- [ ] `BuildSearchPredicate` returns a valid Ent `sql.Predicate` for `tsvector @@ plainto_tsquery`
- [ ] Query strings with 2–50 characters are accepted; empty or too-long queries return an error
- [ ] A helper function `NormalizeQuery(q string) string` strips extra whitespace and lowercases input
- [ ] Unit tests verify predicate SQL output for Spanish `to_tsvector`

**Estimated Time:** 1h
**Dependencies:** T1
**Layer:** Infra

---

## Phase 2: Repository Layer (~10h)

### T5. OrganizationTree repository (list)

**Description:** Implement `OrgTreeRepo` with `List(ctx, filter)` supporting optional `type` enum filter (`corporate`, `retail`). Return node count per tree via a subquery or join. Use the read replica client when available.

**Acceptance Criteria:**
- [ ] `List` returns all trees when filter is empty
- [ ] `List` with `type = "corporate"` returns only corporate trees
- [ ] Each `OrgTree` in the result includes `nodeCount` (count of org_nodes in that tree)
- [ ] `List` queries the read replica if configured
- [ ] Repository tests pass against testcontainer PostgreSQL

**Estimated Time:** 1h
**Dependencies:** T2
**Layer:** Repository

---

### T6. OrgNode repository (CRUD + tree queries)

**Description:** Implement `OrgNodeRepo` with full CRUD, self-referential edge loading (`WithParent`, `WithChildren`), `GetDescendants(path)`, `GetAncestors(path)`, `CountChildren(nodeID)`, and `UpdatePathAndDescendants(tx, oldPath, newPath)` using raw `ltree` predicates via Ent `sql.ExprP`. Also implement `UpdateWithVersion` for optimistic locking.

**Acceptance Criteria:**
- [ ] `Create` inserts a node and returns it with generated ID and default `version = 0`
- [ ] `UpdateWithVersion` increments version only if the current version matches; otherwise returns `ErrStaleVersion`
- [ ] `GetDescendants` returns all nodes where `path <@ $1`, ordered by path
- [ ] `GetAncestors` returns all nodes where `path @> $1`, ordered by path
- [ ] `UpdatePathAndDescendants` updates paths for a moved subtree in a single query
- [ ] `Delete` performs a hard delete (Ent `DeleteOneID`)

**Estimated Time:** 2h
**Dependencies:** T2, T3
**Layer:** Repository

---

### T7. Employee repository (CRUD + search + batch)

**Description:** Implement `EmployeeRepo` with `List(ctx, filter)` supporting `treeId`, `nodeId`, `profileId`, `isActive`, `q`, cursor-based pagination (`cursor`, `limit`), `GetByID`, `GetByIDs` (batch), `ListByManager`, and `Search(ctx, query, limit)` using the `tsvector` predicate from `pkg/search/`.

**Acceptance Criteria:**
- [ ] `List` applies all filters from `EmployeeFilter` correctly (AND logic)
- [ ] Cursor pagination uses `id > cursor` with `limit`; `nextCursor` is returned in meta
- [ ] `Search` uses the `pkg/search` `tsvector` predicate and clamps `limit` to max 50
- [ ] `GetByIDs` resolves up to 100 UUIDs in a single `IN` query
- [ ] `ListByManager` accepts an `activeOnly` flag and uses `idx_employee_manager_active`
- [ ] Repository tests pass against testcontainer PostgreSQL with realistic seed data

**Estimated Time:** 2h
**Dependencies:** T4
**Layer:** Repository

---

### T8. EvaluatorScope repository (compute + cache)

**Description:** Implement `EvaluatorScopeRepo` with `GetByEvaluatorAndCycle(evaluatorID, cycleID)` and `GetByID(scopeID)`. Return the raw scope row including `scopeData` JSONB. Support lookup by `(evaluator_id, cycle_id)` using `idx_evaluatorscope_eval_cycle`.

**Acceptance Criteria:**
- [ ] `GetByEvaluatorAndCycle` returns a scope when both IDs match
- [ ] `GetByEvaluatorAndCycle` returns `ent.IsNotFound` error when no row exists
- [ ] `GetByID` returns the exact scope row
- [ ] Both methods use the read replica if available
- [ ] Repository tests pass against testcontainer PostgreSQL

**Estimated Time:** 1h
**Dependencies:** T2
**Layer:** Repository

---

### T9. Tree traversal queries (ancestors, descendants, path-to-root)

**Description:** Write dedicated repository methods for high-performance tree traversal: `GetPathToRoot(nodeID)` returns ancestor nodes up to the root, `GetSubtree(nodeID, maxDepth)` returns descendants with optional depth limit, and `GetRootNode(treeID)` returns the single root node for a tree. Ensure all methods use `ltree` indexes.

**Acceptance Criteria:**
- [ ] `GetPathToRoot` returns nodes ordered from root to self (or self to root — document chosen order)
- [ ] `GetSubtree` with `maxDepth = 0` returns only the starting node; with `maxDepth = -1` returns all descendants
- [ ] `GetRootNode` returns the node with `parent_id IS NULL` for the given tree
- [ ] All three methods execute `EXPLAIN` plans showing GIST index usage on `path`
- [ ] Benchmark tests show < 10ms for 1000-node trees

**Estimated Time:** 2h
**Dependencies:** T6
**Layer:** Repository

---

## Phase 3: Service Layer (~12h)

### T10. OrgTree service — list trees, get tree structure

**Description:** Implement `OrgTreeService` with `GetTrees`, `GetTree`, `GetTreeNodes`, and `ExportTree`. `GetTreeNodes` must support `flat` and `nested` formats and an optional `depth` limit. `ExportTree` writes a chunked JSON stream. Use `pkg/tree` helpers for formatting. Integrate Redis cache and ETag generation.

**Acceptance Criteria:**
- [ ] `GetTrees` delegates to `OrgTreeRepo.List` and returns a typed slice
- [ ] `GetTreeNodes` returns flat nodes with `parentId`, `depth`, `path`, and `employeeCount`
- [ ] `GetTreeNodes` returns a nested tree when `format = nested`
- [ ] `GetTreeNodes` respects `depth` filter when `depth > 0`
- [ ] `ExportTree` writes valid JSON arrays via `io.Writer` without loading entire tree into memory
- [ ] Results are cached in Redis with key `org:tree:{treeId}:v1` and TTL 1h

**Estimated Time:** 1.5h
**Dependencies:** T5, T9, T17
**Layer:** Service

---

### T11. OrgNode service — CRUD with parent validation

**Description:** Implement `OrgNodeService` with `GetNode`, `CreateNode`, `UpdateNode`, `DeleteNode`, and `MoveNode`. `CreateNode` validates `parentId` belongs to the same `organizationId` and tree type. `DeleteNode` refuses if `CountChildren > 0` (error `NODE_HAS_CHILDREN`). `UpdateNode` delegates to `UpdateWithVersion`.

**Acceptance Criteria:**
- [ ] `CreateNode` rejects `INVALID_TREE_TYPE` when type is not `corporate` or `retail`
- [ ] `DeleteNode` returns `NODE_HAS_CHILDREN` 409 when children exist
- [ ] `UpdateNode` returns `STALE_VERSION` 409 on optimistic lock mismatch
- [ ] `MoveNode` acquires an advisory lock before modifying paths
- [ ] `MoveNode` updates the moved node’s path and all descendants’ paths atomically within a transaction

**Estimated Time:** 2h
**Dependencies:** T6, T17
**Layer:** Service

---

### T12. Employee service — CRUD + search

**Description:** Implement `EmployeeService.ListEmployees`, `GetEmployee`, and `SearchEmployees`. `ListEmployees` applies all filters and returns cursor-pagination meta. `GetEmployee` returns the detail view including embedded `orgNode` and `manager`. `SearchEmployees` delegates to `EmployeeRepo.Search` with a 50-result limit.

**Acceptance Criteria:**
- [ ] `ListEmployees` returns correct filtered results and `nextCursor` when more pages exist
- [ ] `GetEmployee` returns `EMPLOYEE_NOT_FOUND` 404 when ID does not exist
- [ ] `GetEmployee` includes nested `orgNode` and `manager` objects
- [ ] `SearchEmployees` rejects queries shorter than 2 characters
- [ ] `SearchEmployees` clamps limit to 50 and defaults to 20

**Estimated Time:** 1.5h
**Dependencies:** T7, T17
**Layer:** Service

---

### T13. Employee service — GetMyEvaluatees (direct reports)

**Description:** Implement `EmployeeService.GetMyEvaluatees(evaluatorID)` returning direct reports only. Use `EmployeeRepo.ListByManager` with `activeOnly = true`. Cache the result in Redis with key `org:evaluatees:{evaluatorID}:v1` and TTL 24h. Invalidate cache on employee transfer or deactivation.

**Acceptance Criteria:**
- [ ] Returns only employees where `manager_id = evaluatorID` and `is_active = true`
- [ ] Cached response is returned when cache key exists and is valid
- [ ] Cache is invalidated when an employee’s `manager_id` changes
- [ ] Cache is invalidated when an employee is deactivated
- [ ] Response time is < 30ms when cached, < 100ms when uncached (benchmark)

**Estimated Time:** 1h
**Dependencies:** T12
**Layer:** Service

---

### T14. Employee service — GetChainOfCommand (ancestors)

**Description:** Implement `EmployeeService.GetChainOfCommand(empID)` returning the ancestor chain from employee to root. Use the employee’s `org_node.path` to query ancestors via `ltree @>`. Map each node to an `Ancestor` with a computed `relation` label (`self`, `direct_manager`, `director`, `vp`, `ceo`). Cache with key `org:ancestors:{empId}:v1` and TTL 12h.

**Acceptance Criteria:**
- [ ] Returns ancestors ordered from root to self (or self to root — be consistent)
- [ ] Each ancestor has a `relation` field mapped by depth from the employee
- [ ] The first element in the chain is the employee (`relation: self`)
- [ ] The last element is the root node (`relation: ceo` or equivalent)
- [ ] Cache invalidation triggers when `org_node_id` changes or tree is restructured

**Estimated Time:** 1.5h
**Dependencies:** T12, T9
**Layer:** Service

---

### T15. Employee service — BatchResolve

**Description:** Implement `EmployeeService.BatchLookup(ids []uuid.UUID)` resolving up to 100 employee IDs in a single call. Return a stable order matching input order. Missing IDs are omitted silently (document behavior). Integrate a short-lived Redis cache (`org:employee:{empId}:v1`, TTL 1h) for single-employee lookups.

**Acceptance Criteria:**
- [ ] Resolves up to 100 IDs in a single database query
- [ ] Returns employees in the same order as the input `ids` slice
- [ ] Missing IDs do not cause errors; they are simply omitted from results
- [ ] Response time is < 50ms p99 for 100 IDs (benchmark)
- [ ] Each resolved employee is cached individually for 1h

**Estimated Time:** 1h
**Dependencies:** T12
**Layer:** Service

---

### T16. EvaluatorScope service — compute and cache

**Description:** Implement `EvaluatorService.GetEvaluatorScope(evaluatorID, cycleID)`. Try Redis first (`org:scope:{evaluatorId}:{cycleId}:v1`, TTL 24h). On miss, query `EvaluatorScopeRepo.GetByEvaluatorAndCycle`. If `scopeType = department`, resolve all employees in `scopeData.orgNodeIds` using `path <@ ANY(...)`. Return `evaluateeCount`. Cache the result.

**Acceptance Criteria:**
- [ ] Returns cached scope when Redis key exists
- [ ] Computes `evaluateeCount` correctly for `department`, `team`, and `individual` scope types
- [ ] For `department` scope, resolves employees under the specified org node IDs
- [ ] Returns `EMPLOYEE_NOT_FOUND` 404 when evaluator does not exist
- [ ] Cache invalidation triggers on `evaluator_scopes` table changes

**Estimated Time:** 1.5h
**Dependencies:** T8, T17
**Layer:** Service

---

### T17. Cache service — Redis tree cache with generation-based invalidation

**Description:** Implement `OrgCache` in `pkg/cache/org_cache.go` with methods `GetTree`, `SetTree`, `GetETag`, `SetETag`, `InvalidateTreeCache`, `GetEvaluatees`, `SetEvaluatees`, `GetScope`, `SetScope`, `GetAncestors`, `SetAncestors`. Use a generation counter per tree (`org:tree:{treeId}:gen`) so dependent keys embed the generation. `InvalidateTreeCache` increments the generation instead of scanning and deleting keys.

**Acceptance Criteria:**
- [ ] `InvalidateTreeCache` increments `org:tree:{treeId}:gen` via a Redis pipeline
- [ ] `GetTree` and `SetTree` include the current generation in the cache key
- [ ] `GetETag` returns a stable string derived from tree ID and last modification time
- [ ] All cache methods gracefully handle Redis unavailability (fallback to DB, log warning)
- [ ] Cache warming on startup pre-loads all active trees

**Estimated Time:** 1.5h
**Dependencies:** T1
**Layer:** Service

---

## Phase 4: Handler Layer (~8h)

### T18. Tree handler — tree endpoints

**Description:** Implement `OrgTreeHandler` in `api/internal/handler/org/orgtree_handler.go` for `GET /api/v1/org-trees`, `GET /api/v1/org-trees/:treeId`, `GET /api/v1/org-trees/:treeId/nodes`, and `GET /api/v1/org-trees/:treeId/export`. Support query params `type`, `format`, `depth`. Implement ETag and `If-None-Match` 304 handling for node reads.

**Acceptance Criteria:**
- [ ] `GET /org-trees` returns `OrgTreeListResponse` with `data` array
- [ ] `GET /org-trees/:treeId/nodes` returns flat list by default; nested when `format=nested`
- [ ] `GET /org-trees/:treeId/nodes` returns 404 `TREE_NOT_FOUND` for unknown tree
- [ ] `GET /org-trees/:treeId/nodes` returns 304 when `If-None-Match` matches current ETag
- [ ] `GET /org-trees/:treeId/export` streams JSON with `Transfer-Encoding: chunked`
- [ ] All endpoints have table-driven HTTP tests with 80%+ coverage

**Estimated Time:** 2h
**Dependencies:** T10
**Layer:** Handler

---

### T19. Node handler — CRUD endpoints

**Description:** Implement `OrgNodeHandler` in `api/internal/handler/org/orgnode_handler.go` for `GET /api/v1/org-nodes/:nodeId`, `POST /api/v1/org-nodes`, `PUT /api/v1/org-nodes/:nodeId`, `DELETE /api/v1/org-nodes/:nodeId`, and `POST /api/v1/org-nodes/:nodeId/move`. Parse and validate request bodies using DTOs. Map service errors to HTTP status codes per the design error table.

**Acceptance Criteria:**
- [ ] `POST /org-nodes` returns 201 with `OrgNodeDetailResponse`; 400 for invalid parent or tree type
- [ ] `PUT /org-nodes/:nodeId` returns 200 or 409 `STALE_VERSION`
- [ ] `DELETE /org-nodes/:nodeId` returns 204 or 409 `NODE_HAS_CHILDREN`
- [ ] `POST /org-nodes/:nodeId/move` returns 200 or 400 `INVALID_PARENT` when cycle detected
- [ ] All error responses include `code`, `message`, `details`, and `traceId`

**Estimated Time:** 2h
**Dependencies:** T11
**Layer:** Handler

---

### T20. Employee handler — CRUD + search + batch endpoints

**Description:** Implement `EmployeeHandler` in `api/internal/handler/org/employee_handler.go` for `GET /api/v1/employees`, `GET /api/v1/employees/:empId`, `GET /api/v1/employees/search`, and `POST /api/v1/employees/batch`. Parse query filters (`treeId`, `nodeId`, `profileId`, `isActive`, `q`, `cursor`, `limit`). Validate `limit` max 200. Return `EmployeeListResponse` or `EmployeeDetailResponse`.

**Acceptance Criteria:**
- [ ] `GET /employees` applies all query filters and returns cursor-pagination meta
- [ ] `GET /employees/:empId` returns 404 `EMPLOYEE_NOT_FOUND` when ID does not exist
- [ ] `GET /employees/search` rejects `q` shorter than 2 characters with 400
- [ ] `POST /employees/batch` accepts up to 100 IDs and returns `EmployeeListResponse`
- [ ] Batch endpoint returns 200 even if some IDs are missing (omit missing)

**Estimated Time:** 1.5h
**Dependencies:** T12, T15
**Layer:** Handler

---

### T21. Evaluatee handler — my-evaluatees + ancestors endpoints

**Description:** Implement the evaluatee and chain-of-command endpoints in a dedicated handler (or extend `EmployeeHandler`): `GET /api/v1/employees/:empId/evaluatees`, `GET /api/v1/employees/:empId/manager`, and `GET /api/v1/employees/:empId/ancestors`. Return `EmployeeListResponse`, `EmployeeDetailResponse`, and `AncestorChainResponse` respectively.

**Acceptance Criteria:**
- [ ] `GET /employees/:empId/evaluatees` returns direct reports with `EmployeeListResponse`
- [ ] `GET /employees/:empId/manager` returns manager detail or 404 if no manager
- [ ] `GET /employees/:empId/ancestors` returns `AncestorChainResponse` ordered from self to root
- [ ] Each ancestor in the chain has the correct `relation` enum value
- [ ] All endpoints return `EMPLOYEE_NOT_FOUND` 404 for unknown `empId`

**Estimated Time:** 1.5h
**Dependencies:** T13, T14
**Layer:** Handler

---

### T22. Route registration + middleware

**Description:** Register all org hierarchy routes in `api/internal/handler/org/routes.go` using Chi. Attach rate-limiting middleware (100 req/s writes, 2000 req/s reads), context timeout middleware (3s simple reads, 10s tree traversal), and ETag middleware for tree endpoints. Ensure routes are mounted under `/api/v1/`.

**Acceptance Criteria:**
- [ ] All 14 endpoints from the design are registered and reachable
- [ ] Rate limiting returns 429 `RATE_LIMIT_EXCEEDED` with `Retry-After` header
- [ ] Context timeouts return 503 when exceeded
- [ ] ETag middleware is applied to `GET /org-trees/:treeId/nodes` and `/org-trees/:treeId`
- [ ] Routes are grouped and documented with Chi route comments

**Estimated Time:** 1.5h
**Dependencies:** T18, T19, T20, T21
**Layer:** Handler

---

## Phase 5: OpenAPI (~3h)

### T23. OpenAPI 3.1 spec

**Description:** Write the complete OpenAPI 3.1 YAML spec in `api/openapi/org-hierarchy.yaml` covering all 14 endpoints, request/response schemas, and error responses. Include reusable components for `OrgTree`, `OrgNode`, `Employee`, `Ancestor`, `EvaluatorScope`, `PaginationMeta`, and `Error`. Ensure the spec validates cleanly with `swagger-cli validate` or `openapi-generator-cli validate`.

**Acceptance Criteria:**
- [ ] All 14 endpoints are documented with operation IDs, parameters, request bodies, and responses
- [ ] All schemas from the design are defined under `components/schemas`
- [ ] All error responses reference reusable `components/responses` or `Error` schema
- [ ] `swagger-cli validate api/openapi/org-hierarchy.yaml` passes with no errors
- [ ] Spec includes `servers` section with `/api/v1` base path

**Estimated Time:** 2h
**Dependencies:** T22
**Layer:** Contract

---

### T24. TypeScript type generation

**Description:** Generate TypeScript types from the OpenAPI spec into `web/src/lib/api/org-hierarchy.ts` using `openapi-typescript`. Verify the generated types compile in the Vite + Svelte project. Add a `pnpm` script `generate-api` that regenerates types on demand.

**Acceptance Criteria:**
- [ ] `openapi-typescript api/openapi/org-hierarchy.yaml --output web/src/lib/api/org-hierarchy.ts` runs successfully
- [ ] Generated file exports types for all request bodies, responses, and schemas
- [ ] `pnpm run typecheck` (or `tsc --noEmit`) in `web/` passes with generated types
- [ ] `package.json` includes a `generate-api` script
- [ ] No manual edits are required to the generated file for it to compile

**Estimated Time:** 1h
**Dependencies:** T23
**Layer:** Contract

---

## Phase 6: Tests (~10h)

### T25. Unit tests — repositories

**Description:** Write comprehensive unit tests for `OrgTreeRepo`, `OrgNodeRepo`, `EmployeeRepo`, and `EvaluatorScopeRepo` using testcontainers PostgreSQL. Cover CRUD, edge cases (empty results, not found), `UpdateWithVersion` concurrency, and `ltree` predicate correctness. Use `internal/testfixtures/` for reusable seed data.

**Acceptance Criteria:**
- [ ] `OrgTreeRepo.List` tests cover filtering by type and node count accuracy
- [ ] `OrgNodeRepo` tests cover descendants, ancestors, path updates, and version locking
- [ ] `EmployeeRepo` tests cover all filter combinations, search, pagination, and batch lookup
- [ ] `EvaluatorScopeRepo` tests cover lookup by evaluator + cycle and missing rows
- [ ] Overall repository test coverage is >= 80%

**Estimated Time:** 2h
**Dependencies:** T5, T6, T7, T8
**Layer:** Test

---

### T26. Unit tests — services (tree traversal, cache)

**Description:** Write unit tests for all services using mocked repositories (mockery or hand-rolled). Focus on business logic: cycle detection in `MoveNode`, cache invalidation on tree changes, ETag generation, scope computation, and `BatchLookup` ordering. Use table-driven tests.

**Acceptance Criteria:**
- [ ] `OrgNodeService.MoveNode` tests cover valid moves, self-parent, and descendant cycles
- [ ] `OrgTreeService.GetTreeNodes` tests cover flat vs nested formatting and depth filtering
- [ ] `EmployeeService.GetChainOfCommand` tests verify ancestor ordering and relation labels
- [ ] `OrgCache` tests verify generation-based invalidation and fallback on Redis failure
- [ ] Overall service test coverage is >= 85%

**Estimated Time:** 2h
**Dependencies:** T10, T11, T12, T16, T17
**Layer:** Test

---

### T27. Unit tests — handlers

**Description:** Write table-driven HTTP tests for all handlers using `net/http/httptest` + Chi. Test every endpoint for success, 404, 400, 409, and 304 cases. Verify response bodies match OpenAPI schemas (manual assertions). Mock services to isolate handler logic.

**Acceptance Criteria:**
- [ ] Every endpoint has at least one success and one error test
- [ ] 304 Not Modified is tested for tree endpoints with ETag matching
- [ ] Error responses assert exact `code`, `message`, and `details` fields
- [ ] Request body validation is tested (missing required fields, enum violations)
- [ ] Overall handler test coverage is >= 80%

**Estimated Time:** 1.5h
**Dependencies:** T18, T19, T20, T21
**Layer:** Test

---

### T28. Integration tests

**Description:** Write end-to-end integration tests using testcontainers PostgreSQL + Redis (if available). Test full flows: create tree → create nodes → move node → query evaluatees → search employees. Verify cache invalidation, transaction rollback on errors, and advisory lock serialization.

**Acceptance Criteria:**
- [ ] Test flow: create org tree → add nodes → add employees → query tree nodes → verify structure
- [ ] Test flow: move node → verify descendant paths updated → verify old cache invalidated
- [ ] Test flow: employee transfer → verify old and new manager evaluatee caches invalidated
- [ ] Test flow: evaluator scope query → verify correct employee list for department scope
- [ ] Integration test suite runs in CI with `go test -tags=integration ./api/...`

**Estimated Time:** 1.5h
**Dependencies:** T25, T26
**Layer:** Test

---

### T29. Tree traversal correctness test (deep trees)

**Description:** Write a dedicated correctness test that seeds a deep tree (CEO → VP → Director → Manager → Employee, 5+ levels) and asserts exact ancestor/descendant chains. Test both `ltree` and closure table (if implemented) paths. Ensure no off-by-one depth errors.

**Acceptance Criteria:**
- [ ] A 5-level deep tree is seeded with deterministic IDs
- [ ] `GetChainOfCommand` for the deepest employee returns exactly 5 ancestors in correct order
- [ ] `GetDescendants` for the root returns every node in the tree
- [ ] `FilterDepth` with `maxDepth = 2` returns only nodes at depth 0, 1, and 2
- [ ] Test passes for both `ltree` and closure table implementations

**Estimated Time:** 1h
**Dependencies:** T28
**Layer:** Test

---

### T30. Performance test (1000+ node tree)

**Description:** Write Go benchmarks for `GetTreeNodes`, `GetChainOfCommand`, `GetMyEvaluatees`, and `SearchEmployees` using trees of 1,000 and 10,000 nodes. Assert p99 latencies: < 50ms for tree nodes (flat), < 100ms for nested, < 50ms for chain of command, < 50ms for search.

**Acceptance Criteria:**
- [ ] `BenchmarkGetTreeNodes_1000Nodes_Flat` averages < 50ms per call
- [ ] `BenchmarkGetTreeNodes_10000Nodes_Flat` averages < 100ms per call
- [ ] `BenchmarkGetChainOfCommand_1000Nodes` averages < 50ms per call
- [ ] `BenchmarkSearchEmployees_1000Rows` averages < 50ms per call
- [ ] Benchmark results are printed with `-benchmem` and saved to `docs/performance/bench-C5.txt`

**Estimated Time:** 1h
**Dependencies:** T29
**Layer:** Test

---

### T31. Concurrent tree modification test

**Description:** Write a concurrency test that spawns multiple goroutines attempting to move different subtrees simultaneously. Verify that advisory locks prevent path corruption, no nodes become orphaned, and all `ltree` paths remain valid after all goroutines complete.

**Acceptance Criteria:**
- [ ] 10 goroutines concurrently move random subtrees for 100 iterations each
- [ ] Zero `ltree` path violations (all paths remain valid ancestor/descendant relationships)
- [ ] Zero orphaned nodes (every node’s `parent_id` still exists or is null for roots)
- [ ] Advisory locks are acquired and released without deadlock
- [ ] Test completes in < 30 seconds

**Estimated Time:** 1h
**Dependencies:** T30
**Layer:** Test

---

### T32. Load test (2000 req/s)

**Description:** Write a Go-based load test (or k6 script in `tests/load/`) targeting `GET /api/v1/employees` and `GET /api/v1/org-trees/:treeId/nodes`. Assert < 1% error rate and p99 latency < 100ms at 2000 req/s sustained for 60 seconds. Document results in `docs/performance/load-C5.md`.

**Acceptance Criteria:**
- [ ] Load test targets `GET /employees` with randomized cursor pagination
- [ ] Load test targets `GET /org-trees/:treeId/nodes` flat format
- [ ] Sustained 2000 req/s for 60s results in < 1% errors
- [ ] p99 latency is < 100ms for both endpoints
- [ ] Results document includes throughput, latency percentiles, and error rate

**Estimated Time:** 1h
**Dependencies:** T30
**Layer:** Test

---

## Phase 7: Polish (~2h)

### T33. Documentation

**Description:** Write internal documentation: `api/internal/handler/org/README.md` describing endpoint overview, `api/internal/service/org/README.md` describing business rules and caching strategy, and `api/internal/repository/org/README.md` describing index usage and `ltree` query patterns. Update `docs/get-started/` with any new environment variables (Redis URL, read replica DSN).

**Acceptance Criteria:**
- [ ] Handler README lists all 14 endpoints with method, path, and brief description
- [ ] Service README documents cycle detection, cache invalidation, and optimistic locking
- [ ] Repository README documents `ltree` index usage and fallback closure table strategy
- [ ] `docs/get-started/` includes `REDIS_URL` and `DATABASE_REPLICA_DSN` in env example
- [ ] All READMEs are linked from the main project README or `docs/` index

**Estimated Time:** 1h
**Dependencies:** T22, T32
**Layer:** Docs

---

### T34. Prometheus metrics

**Description:** Add Prometheus metrics for all org hierarchy endpoints: request count + duration histograms per endpoint, cache hit/miss counters (`org_cache_hits_total`, `org_cache_misses_total`), tree depth gauge (`org_tree_depth`), and evaluatee count gauge (`org_evaluatee_count`). Register metrics in handler middleware and expose on `/metrics`.

**Acceptance Criteria:**
- [ ] `http_request_duration_seconds` histogram includes labels for route and method
- [ ] `org_cache_hits_total` and `org_cache_misses_total` are incremented per cache key pattern
- [ ] `org_tree_nodes_total` gauge records the number of nodes per tree ID
- [ ] `org_evaluatee_count` gauge records the number of evaluatees per evaluator
- [ ] Metrics endpoint `/metrics` returns valid Prometheus exposition format

**Estimated Time:** 1h
**Dependencies:** T22
**Layer:** Observability

---

## Dependency Graph (Simplified)

```
T1  → T2  → T5  → T10 → T18 → T23 → T24
  ↘ T3  → T6  → T9  → T11 → T19
  ↘ T4  → T7  → T12 → T13 → T21
              ↘ T15 → T20
        ↘ T8  → T16
  ↘ T17 ─────────────────────────────→ T10, T11, T12, T16

T5–T9  → T25
T10–T17 → T26
T18–T22 → T27
T25–T27 → T28 → T29 → T30 → T31 → T32
T22, T32 → T33
T22     → T34
```

---

## Notes

- All task IDs are prefixed with `T` and map directly to the numbered list in the requirements.
- Estimated times assume a senior Go developer with Ent/Chi experience. Junior developers should multiply by 1.5x.
- Parallelization opportunity: T5–T9 (Repository) can be worked on in parallel once T2–T4 are done. T10–T17 (Service) can be worked on in parallel once their respective repositories are done.
- The critical path is: **T1 → T2 → T6 → T9 → T11 → T19 → T22 → T23 → T24 → T28 → T29 → T30 → T31 → T32 → T33** (~31h).
- Cache implementation (T17) is a prerequisite for most service tasks; prioritize it early in Phase 3.
