# mis-evaluados-api Specification

## Purpose

Paginated server-side endpoint returning a manager's direct evaluatees with profile info, search, and self-exclusion. Extends existing `GET /employees/{empId}/evaluatees`.

## API Contract

| Field | Type | Details |
|-------|------|---------|
| Method | `GET` | |
| Path | `/api/v1/employees/{empId}/evaluatees` | `empId` = evaluator UUID |
| Query `offset` | int | Default 0, clamped ≥ 0 |
| Query `limit` | int | 1–200, default 50 |
| Query `q` | string | ILIKE search on name/email/employeeNumber; optional |
| Response `200` | `EmployeeListResponse` | `{ data: EmployeeListItem[], meta: { hasMore, limit, offset, total } }` |
| Response `400` | `APIErrorResponse` | Invalid UUID, limit out of range |
| Response `404` | `APIErrorResponse` | Evaluator not found |

`EmployeeListItem` includes `profileName`, `profileDescription`, `jobTitle` (reuses existing DTO in `api/internal/dto/org/org_dto.go`).

## Requirements

### Requirement: Paginated evaluatee list with search

The endpoint SHALL return direct reports of `empId` with `OFFSET`/`LIMIT` and optional search. SHALL exclude the evaluator's own employee ID from results. SHALL include profile info via JOIN with `evaluation_profiles`.

#### Scenario: Default page load

- GIVEN a manager with 120 direct reports
- WHEN `GET /employees/{managerId}/evaluatees` (no query params)
- THEN response has `limit:50`, `offset:0`, `total:120`, `hasMore:true`
- AND `data` contains 50 items with `profileName` populated
- AND manager's own employee ID is excluded

#### Scenario: Search filters results

- GIVEN a manager's evaluatees include "María Gómez" and "Carlos Pérez"
- WHEN `GET ...?q=María`
- THEN response `data` contains only María
- AND `meta.total` reflects filtered count

#### Scenario: Last page sets hasMore false

- GIVEN 95 evaluatees, `offset=50`, `limit=50`
- WHEN endpoint is called
- THEN `meta.hasMore` is false, `meta.total` is 95

#### Scenario: Invalid limit returns 400

- GIVEN `limit=300`
- WHEN endpoint is called
- THEN response is 400 with message "limit must be between 1 and 200"

### Requirement: Repository layer offset pagination

`EmployeeRepo` SHALL provide `ListByManagerPaginated(ctx, managerID, query, offset, limit)` with profile JOIN and `CountByManager(ctx, managerID, query)` for total count. Both SHALL exclude the manager's own ID.

**Files**: `api/internal/repository/org/employee_repo.go`

### Requirement: Service layer hasMore computation

`EvaluateeService` SHALL provide `GetMyEvaluateesPaginated(ctx, evaluatorID, query, offset, limit)` that computes `hasMore = offset + len(rows) < total` from repo results.

**Files**: `api/internal/service/org/evaluatee_service.go`

### Requirement: Handler query param parsing

`OrgHandler.GetMyEvaluatees` SHALL parse `offset`, `limit`, `q` from query string. When present, SHALL delegate to paginated service; when absent, SHALL default to `offset=0, limit=50`. SHALL clamp `limit` to 1–200 and `offset` to ≥ 0.

**Files**: `api/internal/handler/org/org_handler.go`
