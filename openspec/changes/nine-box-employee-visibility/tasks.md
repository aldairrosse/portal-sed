# Tasks: Nine-Box Employee Visibility

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 450–550 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (backend DTO + evaluator + filter): `api/` + OpenAPI; PR 2 (frontend scope + store): `web/` |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Backend: DTO enrichment, evaluator fix, quadrant filter, OpenAPI | PR 1 | ~350 lines; base = `main` |
| 2 | Frontend: scope fix + store enrichment | PR 2 | ~100 lines; base = PR 1 (frontend fields come from backend DTO) |

## Phase 1: Backend DTO Enrichment (NB-001)

- [ ] **TASK-001**: Add `EmployeeName` + `ProfileID` to DTO and OpenAPI
  - `api/internal/dto/evaluation/evaluation_dto.go`: Add `EmployeeName string \`json:"employeeName"\`` and `ProfileID uuid.UUID \`json:"profileId"\`` to `NineBoxEntryDTO`
  - `api/openapi/evaluations-and-9x9.yaml`: Add `employeeName` (type: string) and `profileId` (type: string, format: uuid) to `NineBoxEntryDTO` schema
  - **Verify**: `go build ./...` compiles; OpenAPI validates
  - **Effort**: S | **Depends on**: —

- [ ] **TASK-002**: Add `GetEmployeesByIDs` and enrich `toEntryDTO`
  - `api/internal/service/evaluation/repo_interfaces.go`: Add `GetEmployeesByIDs(ctx, ids []uuid.UUID) (map[uuid.UUID]*EmployeeInfo, error)` to `NineBoxRepo` interface (also define `EmployeeInfo` struct)
  - `api/internal/repository/evaluation/ninebox_repo.go`: Implement `SELECT id, first_name, last_name, profile_id FROM employees WHERE id = ANY($1)`
  - `api/internal/service/evaluation/ninebox_service.go`: In `GetMatrix` and `GetMatrixByPhase`, batch-collect evaluatee IDs, call `GetEmployeesByIDs`, pass map to `toEntryDTO`; `toEntryDTO` populates `EmployeeName` (as `first_name || ' ' || last_name`) and `ProfileID` from map — fallback to `""` / `uuid.Nil` for missing entries
  - **Verify**: `go test ./api/internal/service/evaluation/... -run TestToEntryDTO` (unit) — test with populated map, empty map, and orphan entry
  - **Effort**: M | **Depends on**: TASK-001

## Phase 2: Evaluator Resolution (NB-003)

- [ ] **TASK-003**: Add `GetManagerMapping` to repo
  - `api/internal/service/evaluation/repo_interfaces.go`: Add `GetManagerMapping(ctx, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)` to `NineBoxRepo`
  - `api/internal/repository/evaluation/ninebox_repo.go`: Implement `SELECT id, manager_id FROM employees WHERE id = ANY($1)` — returns map[employeeID]managerID
  - **Verify**: Integration test with sqlmock — verify query shape: `SELECT id, manager_id FROM employees WHERE id IN (...)`
  - **Effort**: M | **Depends on**: TASK-001

- [ ] **TASK-004**: Fix `RecomputeMatrix` evaluator grouping
  - `api/internal/service/evaluation/ninebox_service.go` `RecomputeMatrix`: Replace self-evaluator loop (`evaluatorGroups[empID] = empID`) with `GetManagerMapping` call — group evaluatees by manager ID; skip employees where `managerID == nil` (log warning, continue)
  - **Verify**: Unit test with mock `GetManagerMapping` — verify matrix created per manager, not per employee; employee with NULL manager is skipped
  - **Effort**: M | **Depends on**: TASK-003

## Phase 3: Quadrant Filter (NB-004)

- [ ] **TASK-005**: Add quadrant filter to entries endpoint
  - `api/internal/repository/evaluation/ninebox_repo.go`: Add `GetMatrixEntriesByQuadrant(ctx, matrixID uuid.UUID, quadrant int) ([]*internal.NineBoxEntry, error)` with `WHERE quadrant = $2`
  - `api/internal/service/evaluation/ninebox_service.go`: Add `GetMatrixEntriesFiltered(ctx, matrixID uuid.UUID, quadrant *int)` method — calls `GetMatrixEntriesByQuadrant` when `quadrant != nil`, falls back to `GetMatrixEntries` otherwise
  - `api/internal/handler/evaluation/evaluation_handler.go` `ListMatrixEntries`: Parse `r.URL.Query().Get("quadrant")` → validate 1–9 → pass to service; return 400 with `code: "INVALID_QUADRANT"` for out-of-range
  - `api/openapi/evaluations-and-9x9.yaml`: Add optional `quadrant` query param (integer, 1–9) to `GET /nine-box/matrices/{matrixId}/entries`
  - **Verify**: `curl /entries?quadrant=5` returns only Q5 entries; `?quadrant=0` returns 400; no param returns all
  - **Effort**: M | **Depends on**: TASK-001

## Phase 4: Frontend Fixes (NB-002 + store enrichment)

- [ ] **TASK-006**: Fix scope resolution for jefe profile
  - `web/src/routes/evaluacion/9x9/+page.svelte` line ~92: Change `getChildren(nodeId)` → `getDescendants(nodeId)` for `case 'jefe'`
  - **Verify**: Dev fixture `emp-jefe-01` now resolves all subnodes (not just direct reports); `pnpm run check` passes
  - **Effort**: S | **Depends on**: —

- [ ] **TASK-007**: Use API `employeeName` + `profileId` in store
  - `web/src/lib/stores/nineBoxStore.svelte.ts` `normalizeApiData()`: Replace `employeeName: ''` with `employeeName: dto.employeeName ?? ''` and `profileId: ''` with `profileId: dto.profileId ?? ''`
  - Regenerate OpenAPI TS types if needed: `pnpm run generate-api-types`
  - **Verify**: `pnpm run check` passes; `NineBoxEntryCard` modal renders employee name (no structural changes — component already reads `entry.employeeName`)
  - **Effort**: S | **Depends on**: TASK-005 (needs backend returning the fields)

## Phase 5: Integration Verification

- [ ] **TASK-008**: End-to-end verification
  - Start API + frontend; recompute matrix for a cycle with multi-level hierarchy
  - Verify: modal in matrix shows employee names (not empty); director jefe sees all descendants; entries assigned under real manager; `?quadrant=5` filters correctly
  - Run: `go test ./api/...` and `pnpm run check` — zero regressions
  - **Verify**: All 6 acceptance criteria from spec pass
  - **Effort**: S | **Depends on**: TASK-001 through TASK-007

### Implementation Order

1. PR 1: TASK-001 → TASK-002 → TASK-003 → TASK-004 → TASK-005 (backend + OpenAPI, ~350 lines)
2. PR 2: TASK-006 → TASK-007 → TASK-008 (frontend + verification, ~100 lines)

TASK-006 can run in parallel with backend work (no dependency). TASK-007 depends on backend DTO being deployed so API returns `employeeName`/`profileId`.
