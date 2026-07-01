# Design: Nine-Box Employee Visibility

## Technical Approach

Five targeted fixes across backend DTO, scope resolution, evaluator mapping, API filtering, and frontend display. No schema migrations — all changes are query/view-layer.

## Architecture Decisions

### Decision: DTO Enrichment Strategy

| Option | Tradeoff | Decision |
|--------|----------|----------|
| JOIN in repo query | Single query, consistent with existing `ListByManagerPaginated` pattern | **Chosen** |
| Batch lookup in service (`GetEmployeesByIDs`) | Two queries, extra interface method, service owns mapping logic | Rejected |
| N+1 per entry in `toEntryDTO()` | Simple but O(n) queries — rejected immediately | Rejected |

**Choice**: Add `GetMatrixEntriesWithEmployee(ctx, matrixID)` to `NineBoxRepo` that JOINs `nine_box_entries` with `employees` to return `employeeName` (`first_name || ' ' || last_name`) and `profileId`. The service's `toEntryDTO()` populates the new fields from the enriched row.

### Decision: Evaluator Resolution in RecomputeMatrix

| Option | Tradeoff | Decision |
|--------|----------|----------|
| `employees.manager_id` direct lookup | Simple, single query, indexed column | **Chosen** |
| `org_chart_nodes` hierarchy traversal | More complex, unnecessary — `manager_id` already stores the reporting line | Rejected |

**Choice**: Add `GetManagerMapping(ctx, employeeIDs []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)` to `NineBoxRepo` — a single `SELECT id, manager_id FROM employees WHERE id IN (...)` query. Replace the self-evaluator loop in `RecomputeMatrix` with this mapping. Employees with NULL `manager_id` are skipped (no manager → no matrix placement).

### Decision: Quadrant Filter Location

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Repo-level SQL WHERE | Efficient, reduces data transfer | **Chosen** |
| Service-level filter post-fetch | Simpler code but fetches all entries unnecessarily | Rejected |

**Choice**: Add optional `quadrant *int` parameter to `GetMatrixEntries` (or a new `GetMatrixEntriesByQuadrant` method). Apply as `WHERE quadrant = $N` in SQL. Handler passes query param through service to repo.

## Data Flow

### DTO Enrichment (read path)

```
Frontend GET /nine-box/matrices/{id}
    → Handler → Service.GetMatrix()
        → Repo.GetMatrixByID() [Ent, with entries]
        → For each entry: Repo.GetEmployeeInfo(entry.EvaluateeID)
            → SELECT first_name, last_name, profile_id FROM employees WHERE id = $1
        → toEntryDTO() populates employeeName, profileId
    → JSON response with enriched fields
```

**Optimization**: Replace per-entry lookup with a single batch query. Collect all `evaluateeID`s, call `GetEmployeesByIDs(ids)` once, build `map[uuid.UUID]*EmployeeInfo`, then populate DTOs from the map.

### Evaluator Resolution (RecomputeMatrix)

```
POST /nine-box/recompute/{cycleId}/{phaseId}
    → Service.RecomputeMatrix()
        → Repo.GetGoalAssigneesByCycle(cycleID) → []employeeID
        → Repo.GetManagerMapping(employeeIDs) → map[evaluatee]manager
        → Group evaluatees by manager (evaluator)
        → For each (manager, evaluatees) group:
            → Get/create matrix for (cycle, manager, phase)
            → Compute tiers + quadrant per evaluatee
            → Upsert entries
```

### Scope Resolution (frontend fix)

```
+page.svelte scopeIds derived:
    jefe: getDescendants(nodeId)  ← was getChildren(nodeId)
    director: getDescendants(nodeId)  ← unchanged
    director-general/rh: all entries  ← unchanged
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/internal/dto/evaluation/evaluation_dto.go` | Modify | Add `EmployeeName string` and `ProfileID uuid.UUID` to `NineBoxEntryDTO` |
| `api/internal/service/evaluation/repo_interfaces.go` | Modify | Add `GetEmployeesByIDs` and `GetManagerMapping` to `NineBoxRepo` interface |
| `api/internal/service/evaluation/ninebox_service.go` | Modify | Batch employee lookup in `GetMatrix`/`GetMatrixByPhase`; fix `RecomputeMatrix` evaluator grouping |
| `api/internal/repository/evaluation/ninebox_repo.go` | Modify | Add `GetEmployeesByIDs`, `GetManagerMapping`, `GetMatrixEntriesByQuadrant` methods |
| `api/internal/handler/evaluation/evaluation_handler.go` | Modify | Parse `quadrant` query param in entries list handler |
| `api/openapi/evaluations-and-9x9.yaml` | Modify | Add `employeeName`, `profileId` to `NineBoxEntryDTO` schema; add `quadrant` query param to entries endpoint |
| `web/src/lib/types/nine-box.ts` | Modify | No change needed — `employeeName` and `profileId` already in `NineBoxEntry` interface |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Modify | Use `dto.employeeName` and `dto.profileId` from API instead of hardcoded empty strings |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Modify | Change `getChildren` → `getDescendants` for `jefe` profile scope |

## Interfaces / Contracts

### DTO addition (Go)

```go
type NineBoxEntryDTO struct {
    // ... existing fields ...
    EmployeeName string    `json:"employeeName"`
    ProfileID    uuid.UUID `json:"profileId"`
}
```

### OpenAPI schema addition

```yaml
NineBoxEntryDTO:
  properties:
    # ... existing ...
    employeeName:
      type: string
      description: "Employee full name (first + last)"
    profileId:
      type: string
      format: uuid
      description: "Employee's evaluation profile ID"
```

### New repo methods

```go
// Batch employee lookup for DTO enrichment
GetEmployeesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*EmployeeInfo, error)

// Manager mapping for evaluator resolution
GetManagerMapping(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]uuid.UUID, error)

// Quadrant-filtered entries
GetMatrixEntriesByQuadrant(ctx context.Context, matrixID uuid.UUID, quadrant int) ([]*internal.NineBoxEntry, error)
```

Where `EmployeeInfo` is:

```go
type EmployeeInfo struct {
    ID        uuid.UUID
    FirstName string
    LastName  string
    ProfileID uuid.UUID
}
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `toEntryDTO()` with enriched data | Table test: entry + EmployeeInfo map → expected DTO |
| Unit | `RecomputeMatrix` evaluator grouping | Mock `GetManagerMapping` → verify matrix created per manager |
| Integration | `GetEmployeesByIDs` with sqlmock | Verify query shape and scanning |
| Integration | `GetMatrixEntriesByQuadrant` | Verify SQL WHERE clause applied |
| E2E | Modal shows employee names | Playwright: navigate to 9x9, click cell, verify names visible |

## Migration / Rollout

No database migration required. Changes are additive (new DTO fields, new query methods). Frontend already has `employeeName` and `profileId` in its `NineBoxEntry` type — it just receives empty strings today. Once the API starts returning values, the frontend store's `normalizeApiData()` will pick them up with a one-line change.

**Rollback**: Revert DTO fields (remove from JSON), restore `getChildren` in frontend. Quadrant filter is additive — safe to keep.

## Open Questions

- [ ] Should `GetManagerMapping` skip employees with `manager_id = NULL` silently, or log a warning for data-quality visibility?
