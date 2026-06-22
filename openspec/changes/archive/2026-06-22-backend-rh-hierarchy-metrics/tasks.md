# Tasks: backend-rh-hierarchy-metrics

## Phase 1 — DTO & Repository

- [x] 1.1 Create `internal/dto/org/metrics_dto.go` — `AreaMetricsResponse`, `AreaMetricsEmployee` types
- [x] 1.2 Create `internal/repository/org/metrics_repo.go` — `MetricsRepo` struct with `client *internal.Client` and `db *sql.DB`
- [x] 1.3 Implement `GetDirectEmployees(ctx, nodeID uuid.UUID) ([]*EmployeeRow, error)` — query employees WHERE org_node_id = $1 AND is_active = true
- [x] 1.4 Implement `GetGoalsByEmployees(ctx, employeeIDs []uuid.UUID) ([]*GoalRow, error)` — query goals via goal_categories WHERE employee_id = ANY($1) AND target_value > 0
- [x] 1.5 Implement `GetRHEvaluationsByEmployees(ctx, employeeIDs []uuid.UUID, cycleID uuid.UUID) ([]*RHRatingRow, error)` — query evaluation_competencies JOIN evaluations WHERE employee_id = ANY($1) AND cycle_id = $2 AND rh_rating IS NOT NULL
- [x] 1.6 Write unit tests for repo queries (`metrics_repo_test.go`)

## Phase 2 — Service Layer

- [x] 2.1 Create `internal/service/org/metrics_service.go` — `MetricsService` interface with `GetAreaMetrics(ctx, nodeID, cycleID string) (*dto.AreaMetricsResponse, error)`
- [x] 2.2 Implement `computeAvgProgress(goals)` — mean of (current_value / target_value) * 100, exclude target=0, return null if empty
- [x] 2.3 Implement `computeAvgRating(ratings)` — mean of rhRating, return null if empty
- [x] 2.4 Implement `countCompleted/goals)` — count where current_value >= target_value
- [x] 2.5 Implement `countPending(goals)` — count where current_value < target_value
- [x] 2.6 Implement `countEmployeesWithGoals(goals)` — distinct employees with at least one goal
- [x] 2.7 Implement `GetAreaMetrics` orchestration — call repo methods, compute aggregates, build response
- [x] 2.8 Handle edge cases: node not found (404), no employees (empty response), cycle not found (use active or 404)
- [x] 2.9 Write unit tests for aggregation functions (`metrics_service_test.go`)

## Phase 3 — Handler & Routes

- [x] 3.1 Add `GetAreaMetrics` method to `OrgHandler` in `org_handler.go` — parse nodeId, optional cycleId, call service, writeJSON
- [x] 3.2 Register route in `routes.go` — `GET /org-nodes/{nodeId}/area-metrics` in read group with auth + rate limit + read replica
- [x] 3.3 Write handler tests (`org_handler_test.go` for area-metrics) — valid request, missing nodeId, service error, no query params
- [x] 3.4 Run `go test ./internal/handler/org/...` — verify all handler tests pass
- [x] 3.5 Run `go test ./internal/service/org/...` — verify all service tests pass
- [x] 3.6 Run `go test ./internal/repository/org/...` — verify all repo tests pass

## Dependencies

- Phase 2 depends on Phase 1 (service uses repo)
- Phase 3 depends on Phase 2 (handler uses service)
- All phases can use existing `EmployeeRow` from `employee_repo.go`

## Notes

- Reuse `EmployeeRow` type from `employee_repo.go` for GetDirectEmployees
- New row types needed: `GoalRow` (for aggregated goals), `RHRatingRow` (for RH ratings)
- No schema changes (Ent models already have all needed fields)
- No migrations needed
