# Design: Backend API para métricas de jerarquía RRHH

## Architecture Decision

Extender el handler/service/repository existente de `org` con un nuevo método dedicado a métricas de área. No se crea un bounded context nuevo porque el dominio es el mismo (org hierarchy) y los datos provienen de las mismas tablas.

## Endpoint

```
GET /api/v1/org-nodes/{nodeId}/area-metrics?cycleId={uuid}
```

- `nodeId` (path, required): UUID del nodo org
- `cycleId` (query, optional): UUID del ciclo. Si no se provee, se busca el ciclo activo de la organización del nodo

## Response Shape

```json
{
  "nodeId": "550e8400-e29b-41d4-a716-446655440000",
  "employeeCount": 4,
  "employeesWithGoals": 3,
  "avgProgress": 68.3,
  "completedGoals": 5,
  "pendingGoals": 7,
  "avgRating": 3.5,
  "ratingsCount": 3,
  "employees": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440001",
      "firstName": "Carmen",
      "lastName": "Jiménez Castro",
      "profileId": "550e8400-e29b-41d4-a716-446655440010"
    }
  ]
}
```

## Data Flow

```
Handler.GetAreaMetrics
  → parse nodeId (UUID validation)
  → parse optional cycleId
  → svc.GetAreaMetrics(nodeId, cycleId)
    → repo.GetDirectEmployees(nodeId)           -- employees WHERE org_node_id = nodeId
    → repo.GetGoalsByEmployees(employeeIds, cycleId) -- goals via categories
    → repo.GetRHEvaluationsByEmployees(employeeIds, cycleId) -- evaluation_competencies
    → compute aggregates in Go (not SQL)
    → return AreaMetricsResponse
```

La agregación se hace en Go (service layer), no en SQL. Esto mantiene la lógica testeable y separada de los queries.

## New Files

| File | Purpose |
|------|---------|
| `internal/handler/org/org_handler.go` | Add `GetAreaMetrics` method (extend existing) |
| `internal/handler/org/routes.go` | Register new route in existing read group |
| `internal/service/org/metrics_service.go` | New file: `MetricsService` interface + impl |
| `internal/repository/org/metrics_repo.go` | New file: `MetricsRepo` with 3 queries |
| `internal/dto/org/metrics_dto.go` | New file: request/response types |

## SQL Queries

### 1. GetDirectEmployees

```sql
SELECT e.id, e.first_name, e.last_name, e.profile_id
FROM employees e
WHERE e.org_node_id = $1
  AND e.is_active = true
ORDER BY e.last_name, e.first_name
```

Simple, no recursion. Matches employees directly assigned to the node.

### 2. GetGoalsByEmployees

```sql
SELECT g.id, g.name, g.target_value, g.current_value, g.state,
       gc.employee_id
FROM goals g
JOIN goal_categories gc ON gc.id = g.category_id
WHERE gc.employee_id = ANY($1::uuid[])
  AND g.target_value > 0
```

Notas:
- `goal_categories` links goals to employees (via `employee_id`)
- No necesitamos `goal_assignments` porque ya tenemos el employee_id directamente del goal_category
- Excluimos `target_value = 0` para evitar división por cero
- El filtro por cycle_id se aplica a través de la relación employee → cycle (via `goal_assignments` si existe, o directamente al employee si el ciclo es el activo)

### 3. GetRHEvaluationsByEmployees

```sql
SELECT ec.rh_rating, ec.evaluation_id,
       e.employee_id
FROM evaluation_competencies ec
JOIN evaluations e ON e.id = ec.evaluation_id
WHERE e.employee_id = ANY($1::uuid[])
  AND e.cycle_id = $2
  AND ec.rh_rating IS NOT NULL
```

Notas:
- `evaluation_competencies` tiene el `rh_rating` directamente
- Se filtra por `cycle_id` para obtener solo evaluaciones del ciclo solicitado
- Solo competencias con `rh_rating` no nulo

## Aggregation Logic (Service Layer)

```go
func (s *metricsService) GetAreaMetrics(ctx context.Context, nodeID, cycleID string) (*dto.AreaMetricsResponse, error) {
    // 1. Get direct employees
    employees, err := s.repo.GetDirectEmployees(ctx, nodeID)

    // 2. Get goals for those employees
    empIDs := extractIDs(employees)
    goals, err := s.repo.GetGoalsByEmployees(ctx, empIDs)

    // 3. Get RH evaluations
    ratings, err := s.repo.GetRHEvaluationsByEmployees(ctx, empIDs, cycleID)

    // 4. Compute aggregates
    avgProgress := computeAvgProgress(goals)  // mean of (current/target)*100, exclude target=0
    completed := countCompleted(goals)         // current >= target
    pending := countPending(goals)             // current < target
    avgRating := computeAvgRating(ratings)     // mean of rhRating, null if empty

    return &AreaMetricsResponse{
        NodeID:           nodeID,
        EmployeeCount:    len(employees),
        EmployeesWithGoals: countEmployeesWithGoals(goals),
        AvgProgress:      avgProgress,
        CompletedGoals:   completed,
        PendingGoals:     pending,
        AvgRating:        avgRating,
        RatingsCount:     len(ratings),
        Employees:        mapEmployees(employees),
    }, nil
}
```

## Route Registration

In `routes.go`, add inside the existing read group:

```go
r.Group(func(r chi.Router) {
    r.Use(middleware.RateLimit(readRateLimit))
    r.Use(readReplicaMiddleware)
    r.Get("/org-nodes/{nodeId}/area-metrics", handler.GetAreaMetrics)
})
```

## Testing Strategy

- **Unit tests** for aggregation functions in `metrics_service_test.go`
- **Repository tests** with test fixtures (or mock DB)
- **Handler tests** following existing pattern (Chi context + httptest)

## Edge Cases

| Case | Behavior |
|------|----------|
| nodeId not found | 404 with domain error |
| No employees in node | Return employeeCount=0, all metrics null/zero |
| No goals for any employee | avgProgress=null, completed=0, pending=0 |
| No RH evaluations | avgRating=null, ratingsCount=0 |
| cycleId not found | Use current active cycle, or 404 if none |
| Node has no organization | 404 (node must belong to an org) |
