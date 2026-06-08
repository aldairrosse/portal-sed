# Design: C4 — goals-api

## 1. Overview

C4 delivers the core employee-facing REST API for goals management in the SED Evaluación de Desempeño platform.

This change implements the full CRUD surface for:
- `GoalCategory` — custom per-employee categories with weight.
- `Goal` — weighted objectives inside categories, with progress tracking.
- `KPI` — reusable shared indicators.
- `GoalKpiLink` — N:M binding between goals and KPIs.
- `GoalAssignment` — mapping an employee's goal tree to an evaluation cycle.

It enforces the **Double 100%** weighting rule (decision #1), **per-employee custom categories** (decision #5), **reusable KPIs** (decision #6), and **phase-based editing restrictions** driven by the evaluation lifecycle (C2).

This is the highest-traffic API in the system; every employee touches it at least monthly during a cycle. The design prioritizes correctness under concurrency, atomic batch writes, and defensive validation.

---

## 2. Package Structure

```
api/internal/
├── handler/goal/
│   ├── category_handler.go
│   ├── category_handler_test.go
│   ├── goal_handler.go
│   ├── goal_handler_test.go
│   ├── kpi_handler.go
│   ├── kpi_handler_test.go
│   ├── assignment_handler.go
│   ├── assignment_handler_test.go
│   ├── weight_handler.go
│   ├── weight_handler_test.go
│   └── routes.go              # Chi router wiring
├── service/goal/
│   ├── goal_service.go          # Business rules, phase checks, weight validation
│   ├── goal_service_test.go
│   ├── kpi_service.go
│   └── kpi_service_test.go
├── repository/goal/
│   ├── goal_repo.go             # Ent queries, SUM aggregates, N:M links
│   ├── goal_repo_test.go
│   ├── kpi_repo.go
│   └── kpi_repo_test.go
└── pkg/
    ├── validation/
    │   └── weighting.go         # Double 100% math + epsilon logic
    └── batch/
        └── batch.go             # Batch operation helper (atomic tx wrapper)
```

**Rationale:**
- `handler/goal/` groups all HTTP-facing code for the goals bounded context.
- `service/goal/` isolates business rules (weight sums, phase gating) from HTTP and persistence details.
- `repository/goal/` holds all Ent-specific queries; no Ent client leaks into service interfaces.
- `pkg/validation/` and `pkg/batch/` are tiny, reusable packages with no external dependencies beyond stdlib.

---

## 3. Handler Layer

### 3.1 Endpoint Inventory

| # | Method | Path | Request Body | Response Body | Allowed Phases |
|---|--------|------|--------------|---------------|----------------|
| 1 | `GET` | `/api/v1/employees/{empId}/categories` | — | `CategoryListResponse` | all |
| 2 | `POST` | `/api/v1/employees/{empId}/categories` | `CreateCategoryRequest` | `CategoryResponse` | `asignacion` |
| 3 | `PUT` | `/api/v1/employees/{empId}/categories/{catId}` | `UpdateCategoryRequest` | `CategoryResponse` | `asignacion`, `avance`* |
| 4 | `DELETE` | `/api/v1/employees/{empId}/categories/{catId}` | — | `204 No Content` | `asignacion` |
| 5 | `POST` | `/api/v1/employees/{empId}/categories/{catId}/goals` | `CreateGoalRequest` | `GoalResponse` | `asignacion` |
| 6 | `PUT` | `/api/v1/goals/{goalId}` | `UpdateGoalRequest` | `GoalResponse` | `asignacion`, `avance`** |
| 7 | `DELETE` | `/api/v1/goals/{goalId}` | — | `204 No Content` | `asignacion` |
| 8 | `PATCH` | `/api/v1/goals/{goalId}/progress` | `UpdateProgressRequest` | `GoalResponse` | `avance` |
| 9 | `POST` | `/api/v1/goals/batch` | `BatchGoalRequest` | `BatchGoalResponse` | `asignacion` |
| 10 | `POST` | `/api/v1/employees/{empId}/validate-weights` | — | `WeightValidationResponse` | all |
| 11 | `GET` | `/api/v1/kpis` | — | `KpiListResponse` | all |
| 12 | `POST` | `/api/v1/kpis` | `CreateKpiRequest` | `KpiResponse` | all*** |
| 13 | `PUT` | `/api/v1/kpis/{kpiId}` | `UpdateKpiRequest` | `KpiResponse` | all*** |
| 14 | `DELETE` | `/api/v1/kpis/{kpiId}` | — | `204 No Content` | all*** |
| 15 | `POST` | `/api/v1/goals/{goalId}/kpis` | `LinkKpiRequest` | `GoalResponse` | `asignacion` |
| 16 | `DELETE` | `/api/v1/goals/{goalId}/kpis/{kpiId}` | — | `204 No Content` | `asignacion` |
| 17 | `GET` | `/api/v1/employees/{empId}/assignments` | — | `AssignmentResponse` | all |
| 18 | `POST` | `/api/v1/employees/{empId}/assignments` | `CreateAssignmentRequest` | `AssignmentResponse` | `asignacion` |

\* In `avance`, only `weight` may be updated if the category contains no goal changes; otherwise `PHASE_RESTRICTED`.  
\** In `avance`, only `currentValue` and non-weight/non-target fields are allowed; weight/target changes are rejected.  
\*** KPI catalog management is not phase-restricted, but requires `admin` or `rh` role (enforced by RBAC middleware, not this change).

### 3.2 Request / Response Schemas

#### `CreateCategoryRequest`
```json
{
  "name": "Crecimiento",
  "description": "Objetivos de desarrollo personal",
  "weight": 30.0
}
```
- `name`: required, 1–100 chars, unique per employee.
- `weight`: required, float, `0 < weight <= 100`.

#### `UpdateCategoryRequest`
```json
{
  "name": "Crecimiento",
  "description": "Objetivos de desarrollo",
  "weight": 35.0
}
```
- Same constraints as create; `name` uniqueness checked excluding current category.

#### `CreateGoalRequest`
```json
{
  "name": "Completar curso de liderazgo",
  "description": "Curso interno de 40h",
  "unit": "porcentaje",
  "weight": 50.0,
  "targetValue": 100.0,
  "kpiIds": ["uuid-kpi-1", "uuid-kpi-2"]
}
```
- `name`: required, 1–255 chars.
- `unit`: enum `porcentaje | moneda | numero`.
- `weight`: required, float, `0 < weight <= 100`.
- `targetValue`: required, float, `> 0`.
- `kpiIds`: optional, max 5. Validated for existence in `KPI` table.

#### `UpdateGoalRequest`
```json
{
  "name": "Completar curso avanzado",
  "description": "Curso de 60h",
  "unit": "porcentaje",
  "weight": 50.0,
  "targetValue": 100.0,
  "version": 3,
  "kpiIds": ["uuid-kpi-1"]
}
```
- `version`: required for optimistic locking. Mismatch → `409 CONCURRENT_MODIFICATION`.
- Phase restrictions apply (see 3.1).

#### `UpdateProgressRequest`
```json
{
  "currentValue": 75.0
}
```
- `currentValue`: float, `>= 0`.
- Only accepted in `avance` phase.

#### `BatchGoalRequest`
```json
{
  "items": [
    {
      "operation": "create",
      "categoryId": "uuid-cat-1",
      "goal": { "name": "...", "weight": 30.0, ... }
    },
    {
      "operation": "update",
      "goalId": "uuid-goal-2",
      "goal": { "name": "...", "weight": 40.0, "version": 2, ... }
    }
  ]
}
```
- Max 50 items per batch.
- Atomic: all succeed or all fail.
- Weight validation runs on the final projected state.

#### `WeightValidationResponse`
```json
{
  "valid": false,
  "categorySum": 95.0,
  "expectedSum": 100.0,
  "deficit": 5.0,
  "goalSums": [
    {
      "categoryId": "uuid-cat-1",
      "categoryName": "Crecimiento",
      "sum": 110.0,
      "expectedSum": 100.0,
      "deficit": -10.0
    }
  ]
}
```

### 3.3 Validation Rules

| Rule | HTTP | Error Code |
|------|------|------------|
| Duplicate category name per employee | `409` | `DUPLICATE_CATEGORY_NAME` |
| Category weight not in `(0, 100]` | `400` | `INVALID_WEIGHT_RANGE` |
| Goal weight not in `(0, 100]` | `400` | `INVALID_WEIGHT_RANGE` |
| Goal targetValue `<= 0` | `400` | `INVALID_TARGET_VALUE` |
| KPI link count > 5 | `400` | `KPI_LINK_LIMIT_EXCEEDED` |
| Optimistic lock version mismatch | `409` | `CONCURRENT_MODIFICATION` |
| Operation not allowed in current phase | `403` | `PHASE_RESTRICTED` |
| Weight sums invalid (after write) | `422` | `WEIGHT_SUM_INVALID` |
| Goal not found | `404` | `GOAL_NOT_FOUND` |
| Category not found | `404` | `CATEGORY_NOT_FOUND` |
| KPI not found | `404` | `KPI_NOT_FOUND` |
| KPI linked to goals on delete | `409` | `KPI_LINKED_CANNOT_DELETE` |
| Batch item count > 50 | `400` | `BATCH_SIZE_EXCEEDED` |
| Idempotency key reused with different payload | `409` | `IDEMPOTENCY_KEY_REUSE` |

### 3.4 Phase-Based Access Control

Phase enforcement is implemented at the **service layer**, not middleware, because the phase check depends on the employee's active cycle (queried from C2). The handler calls `service.EnforcePhase(ctx, empID, allowedPhases)` before any mutating operation.

```go
// Pseudo-code for handler phase gating
func (h *GoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
    goalID := chi.URLParam(r, "goalId")
    empID := r.Context().Value("employeeId").(string)

    // Service checks the employee's current cycle phase via C2 integration
    // and returns PHASE_RESTRICTED if the operation is disallowed.
    goal, err := h.svc.UpdateGoal(r.Context(), empID, goalID, req)
    // ...
}
```

---

## 4. Service Layer

### 4.1 Interface

```go
package goal

type Service interface {
    // Categories
    ListCategories(ctx context.Context, empID string, cursor string, limit int) (CategoryList, error)
    CreateCategory(ctx context.Context, empID string, req CreateCategoryRequest) (Category, error)
    UpdateCategory(ctx context.Context, empID, catID string, req UpdateCategoryRequest) (Category, error)
    DeleteCategory(ctx context.Context, empID, catID string) error

    // Goals
    CreateGoal(ctx context.Context, empID, catID string, req CreateGoalRequest) (Goal, error)
    UpdateGoal(ctx context.Context, empID, goalID string, req UpdateGoalRequest) (Goal, error)
    DeleteGoal(ctx context.Context, empID, goalID string) error
    UpdateGoalProgress(ctx context.Context, empID, goalID string, req UpdateProgressRequest) (Goal, error)
    BatchCreateUpdateGoals(ctx context.Context, empID string, req BatchGoalRequest) ([]Goal, error)

    // KPIs
    ListKPIs(ctx context.Context, cursor string, limit int) (KpiList, error)
    CreateKPI(ctx context.Context, req CreateKpiRequest) (KPI, error)
    UpdateKPI(ctx context.Context, kpiID string, req UpdateKpiRequest) (KPI, error)
    DeleteKPI(ctx context.Context, kpiID string) error
    LinkKPI(ctx context.Context, empID, goalID string, req LinkKpiRequest) error
    UnlinkKPI(ctx context.Context, empID, goalID, kpiID string) error

    // Weight validation
    ValidateDoubleWeighting(ctx context.Context, empID string) (WeightValidationResult, error)

    // Assignments
    GetAssignment(ctx context.Context, empID string) (GoalAssignment, error)
    CreateAssignment(ctx context.Context, empID string, req CreateAssignmentRequest) (GoalAssignment, error)
}
```

### 4.2 Weight Validation — Critical Path

```go
// WeightValidationResult is returned by ValidateDoubleWeighting.
type WeightValidationResult struct {
    Valid        bool
    CategorySum  float64
    ExpectedSum  float64
    Deficit      float64
    GoalSums     []CategoryGoalSum
}

type CategoryGoalSum struct {
    CategoryID   string
    CategoryName string
    Sum          float64
    ExpectedSum  float64
    Deficit      float64
}
```

**Algorithm:**
1. Query all `GoalCategory` rows for `employee_id = ?`.
2. Sum `weight` → `categorySum`.
3. For each category, query all `Goal` rows and sum `weight` → `goalSum`.
4. `valid = |categorySum - 100.0| <= ε && every goalSum within ε of 100.0`.
5. `ε = 0.01` (one hundredth of a percent).

**Go implementation (service layer):**
```go
const epsilon = 0.01

func (s *service) ValidateDoubleWeighting(ctx context.Context, empID string) (WeightValidationResult, error) {
    cats, err := s.repo.ListCategoriesByEmployee(ctx, empID)
    if err != nil {
        return WeightValidationResult{}, err
    }

    var catSum float64
    var goalSums []CategoryGoalSum
    valid := true

    for _, c := range cats {
        catSum += c.Weight
        goals, err := s.repo.ListGoalsByCategory(ctx, c.ID)
        if err != nil {
            return WeightValidationResult{}, err
        }
        var gSum float64
        for _, g := range goals {
            gSum += g.Weight
        }
        if math.Abs(gSum-100.0) > epsilon {
            valid = false
        }
        goalSums = append(goalSums, CategoryGoalSum{
            CategoryID:   c.ID,
            CategoryName: c.Name,
            Sum:          gSum,
            ExpectedSum:  100.0,
            Deficit:      100.0 - gSum,
        })
    }

    if math.Abs(catSum-100.0) > epsilon {
        valid = false
    }

    return WeightValidationResult{
        Valid:       valid,
        CategorySum: catSum,
        ExpectedSum: 100.0,
        Deficit:     100.0 - catSum,
        GoalSums:    goalSums,
    }, nil
}
```

**Why `0.01`?** Currency-style rounding errors from user input (e.g., 33.33 × 3). The frontend should still guide users to exact 100%, but the backend tolerates minute float variance.

### 4.3 CreateGoal — Preventing Category Overflow

```go
func (s *service) CreateGoal(ctx context.Context, empID, catID string, req CreateGoalRequest) (Goal, error) {
    if err := s.enforcePhase(ctx, empID, phaseAsignacion); err != nil {
        return Goal{}, err
    }

    // Transaction: SELECT FOR UPDATE on category + SUM(goal.weight)
    goal, err := s.repo.WithTx(ctx, func(r Repository) (Goal, error) {
        cat, err := r.LockCategory(ctx, catID) // SELECT FOR UPDATE
        if err != nil {
            return Goal{}, err
        }
        if cat.EmployeeID != empID {
            return Goal{}, ErrCategoryNotFound
        }

        sum, err := r.SumGoalWeightsByCategory(ctx, catID)
        if err != nil {
            return Goal{}, err
        }
        if sum+req.Weight > 100.0+epsilon {
            return Goal{}, ErrWeightSumInvalid
        }

        return r.CreateGoal(ctx, catID, req)
    })
    return goal, err
}
```

### 4.4 UpdateGoalProgress

```go
func (s *service) UpdateGoalProgress(ctx context.Context, empID, goalID string, req UpdateProgressRequest) (Goal, error) {
    if err := s.enforcePhase(ctx, empID, phaseAvance); err != nil {
        return Goal{}, err
    }
    return s.repo.UpdateGoalCurrentValue(ctx, goalID, req.CurrentValue)
}
```

### 4.5 BatchCreateUpdateGoals — Atomic

```go
func (s *service) BatchCreateUpdateGoals(ctx context.Context, empID string, req BatchGoalRequest) ([]Goal, error) {
    if err := s.enforcePhase(ctx, empID, phaseAsignacion); err != nil {
        return nil, err
    }
    if len(req.Items) > 50 {
        return nil, ErrBatchSizeExceeded
    }

    return s.repo.WithTx(ctx, func(r Repository) ([]Goal, error) {
        var out []Goal
        for _, item := range req.Items {
            switch item.Operation {
            case "create":
                g, err := r.CreateGoal(ctx, item.CategoryID, item.Goal)
                if err != nil {
                    return nil, err // rolls back entire tx
                }
                out = append(out, g)
            case "update":
                g, err := r.UpdateGoal(ctx, item.GoalID, item.Goal)
                if err != nil {
                    return nil, err
                }
                out = append(out, g)
            }
        }
        // Post-batch validation: compute final weight sums atomically inside tx
        if err := s.validateWeightInvariantTx(ctx, r, empID); err != nil {
            return nil, err
        }
        return out, nil
    })
}
```

---

## 5. Repository Layer

### 5.1 Repository Interface

```go
package goal

type Repository interface {
    // Categories
    ListCategoriesByEmployee(ctx context.Context, empID string) ([]Category, error)
    CreateCategory(ctx context.Context, empID string, req CreateCategoryRequest) (Category, error)
    UpdateCategory(ctx context.Context, catID string, req UpdateCategoryRequest) (Category, error)
    DeleteCategory(ctx context.Context, catID string) error
    LockCategory(ctx context.Context, catID string) (Category, error) // SELECT FOR UPDATE

    // Goals
    ListGoalsByCategory(ctx context.Context, catID string) ([]Goal, error)
    CreateGoal(ctx context.Context, catID string, req CreateGoalRequest) (Goal, error)
    UpdateGoal(ctx context.Context, goalID string, req UpdateGoalRequest) (Goal, error)
    DeleteGoal(ctx context.Context, goalID string) error
    UpdateGoalCurrentValue(ctx context.Context, goalID string, current float64) (Goal, error)
    SumGoalWeightsByCategory(ctx context.Context, catID string) (float64, error)

    // KPIs
    ListKPIs(ctx context.Context, cursor string, limit int) (KpiList, error)
    CreateKPI(ctx context.Context, req CreateKpiRequest) (KPI, error)
    UpdateKPI(ctx context.Context, kpiID string, req UpdateKpiRequest) (KPI, error)
    DeleteKPI(ctx context.Context, kpiID string) error
    CountGoalLinksByKPI(ctx context.Context, kpiID string) (int, error)
    LinkKPI(ctx context.Context, goalID, kpiID string) error
    UnlinkKPI(ctx context.Context, goalID, kpiID string) error
    ReplaceGoalKpiLinks(ctx context.Context, goalID string, kpiIDs []string) error

    // Weight validation
    SumCategoryWeightsByEmployee(ctx context.Context, empID string) (float64, error)
    SumGoalWeightsByCategoryID(ctx context.Context, catID string) (float64, error)

    // Assignments
    GetAssignment(ctx context.Context, empID string) (GoalAssignment, error)
    CreateAssignment(ctx context.Context, empID string, req CreateAssignmentRequest) (GoalAssignment, error)

    // Tx
    WithTx(ctx context.Context, fn func(Repository) error) error
}
```

### 5.2 Ent Patterns for N:M KPI Links

```go
// LinkKPI creates a GoalKpiLink edge.
func (r *entRepo) LinkKPI(ctx context.Context, goalID, kpiID string) error {
    return r.client.GoalKpiLink.Create().
        SetGoalID(goalID).
        SetKpiID(kpiID).
        OnConflict().
        DoNothing(). // idempotent
        Exec(ctx)
}

// ReplaceGoalKpiLinks deletes existing links and recreates them atomically.
func (r *entRepo) ReplaceGoalKpiLinks(ctx context.Context, goalID string, kpiIDs []string) error {
    _, err := r.client.GoalKpiLink.Delete().
        Where(goalkpilink.GoalID(goalID)).
        Exec(ctx)
    if err != nil {
        return err
    }
    for _, id := range kpiIDs {
        if err := r.LinkKPI(ctx, goalID, id); err != nil {
            return err
        }
    }
    return nil
}
```

### 5.3 Weight Calculation Queries (SUM Aggregates)

```go
func (r *entRepo) SumGoalWeightsByCategory(ctx context.Context, catID string) (float64, error) {
    var result struct{ Sum *float64 }
    err := r.client.Goal.Query().
        Where(goal.CategoryID(catID)).
        Aggregate(ent.Sum(goal.FieldWeight)).
        Scan(ctx, &result)
    if err != nil {
        return 0, err
    }
    if result.Sum == nil {
        return 0, nil
    }
    return *result.Sum, nil
}
```

### 5.4 Phase-Aware Queries

The repository itself does **not** enforce phase; it exposes raw CRUD. Phase gating lives in the service layer because it requires a cross-domain call to C2 (`evaluation-lifecycle-api`) to fetch the current cycle and phase.

---

## 6. Phase Enforcement

### 6.1 Source of Truth

The evaluation cycle phase is owned by C2. The goals service queries it via an internal interface:

```go
type PhaseChecker interface {
    GetCurrentPhase(ctx context.Context, empID string) (CyclePhase, error)
}
```

### 6.2 Enforcement Matrix

| Operation | `asignacion` | `avance` | `cierre` |
|-----------|--------------|----------|----------|
| Create category | ✅ | ❌ `PHASE_RESTRICTED` | ❌ `PHASE_RESTRICTED` |
| Update category name/desc | ✅ | ❌ | ❌ |
| Update category weight | ✅ | ⚠️* | ❌ |
| Delete category | ✅ | ❌ | ❌ |
| Create goal | ✅ | ❌ | ❌ |
| Update goal (weight/target) | ✅ | ❌ | ❌ |
| Update goal (name/desc/unit) | ✅ | ⚠️** | ❌ |
| Update progress (`currentValue`) | ❌ | ✅ | ❌ |
| Delete goal | ✅ | ❌ | ❌ |
| Batch create/update | ✅ | ❌ | ❌ |
| Link/unlink KPI | ✅ | ❌ | ❌ |
| Read (all GET) | ✅ | ✅ | ✅ |
| Validate weights | ✅ | ✅ | ✅ |
| Assignment create | ✅ | ❌ | ❌ |

\* Allowed only if no goals are modified concurrently; service checks via tx.  
\** Allowed in `avance` only if the field is not weight/target; handler strips disallowed fields before calling service.

### 6.3 Service-Level Phase Check

```go
func (s *service) enforcePhase(ctx context.Context, empID string, required ...CyclePhase) error {
    phase, err := s.phaseChecker.GetCurrentPhase(ctx, empID)
    if err != nil {
        return err
    }
    for _, p := range required {
        if p == phase {
            return nil
        }
    }
    return ErrPhaseRestricted
}
```

---

## 7. OpenAPI 3.1 Spec

```yaml
openapi: 3.1.0
info:
  title: SED Goals API
  version: '1.0.0'
  description: Core employee-facing API for goals, categories, KPIs, and weight validation.

servers:
  - url: /api/v1

paths:
  /employees/{empId}/categories:
    get:
      operationId: listCategories
      parameters:
        - $ref: '#/components/parameters/EmpId'
        - $ref: '#/components/parameters/Cursor'
        - $ref: '#/components/parameters/Limit'
      responses:
        '200':
          description: List of categories with nested goals
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CategoryListResponse'
        '429': { $ref: '#/components/responses/RateLimit' }
    post:
      operationId: createCategory
      parameters:
        - $ref: '#/components/parameters/EmpId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateCategoryRequest'
      responses:
        '201':
          description: Created category
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CategoryResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }
        '409': { $ref: '#/components/responses/DuplicateCategoryName' }

  /employees/{empId}/categories/{catId}:
    put:
      operationId: updateCategory
      parameters:
        - $ref: '#/components/parameters/EmpId'
        - $ref: '#/components/parameters/CatId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateCategoryRequest'
      responses:
        '200':
          description: Updated category
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/CategoryResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }
    delete:
      operationId: deleteCategory
      parameters:
        - $ref: '#/components/parameters/EmpId'
        - $ref: '#/components/parameters/CatId'
      responses:
        '204': { description: Deleted }
        '403': { $ref: '#/components/responses/PhaseRestricted' }

  /employees/{empId}/categories/{catId}/goals:
    post:
      operationId: createGoal
      parameters:
        - $ref: '#/components/parameters/EmpId'
        - $ref: '#/components/parameters/CatId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateGoalRequest'
      responses:
        '201':
          description: Created goal
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GoalResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }
        '422': { $ref: '#/components/responses/WeightInvalid' }

  /goals/{goalId}:
    put:
      operationId: updateGoal
      parameters:
        - $ref: '#/components/parameters/GoalId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateGoalRequest'
      responses:
        '200':
          description: Updated goal
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GoalResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }
        '409': { $ref: '#/components/responses/ConcurrentModification' }
    delete:
      operationId: deleteGoal
      parameters:
        - $ref: '#/components/parameters/GoalId'
      responses:
        '204': { description: Deleted }
        '403': { $ref: '#/components/responses/PhaseRestricted' }

  /goals/{goalId}/progress:
    patch:
      operationId: updateGoalProgress
      parameters:
        - $ref: '#/components/parameters/GoalId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateProgressRequest'
      responses:
        '200':
          description: Updated progress
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GoalResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }

  /goals/batch:
    post:
      operationId: batchGoals
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/BatchGoalRequest'
      responses:
        '200':
          description: Batch processed
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BatchGoalResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }
        '422': { $ref: '#/components/responses/WeightInvalid' }

  /employees/{empId}/validate-weights:
    post:
      operationId: validateWeights
      parameters:
        - $ref: '#/components/parameters/EmpId'
      responses:
        '200':
          description: Validation result
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/WeightValidationResponse'

  /kpis:
    get:
      operationId: listKPIs
      parameters:
        - $ref: '#/components/parameters/Cursor'
        - $ref: '#/components/parameters/Limit'
      responses:
        '200':
          description: KPI list
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/KpiListResponse'
    post:
      operationId: createKPI
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateKpiRequest'
      responses:
        '201':
          description: Created KPI
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/KpiResponse'

  /kpis/{kpiId}:
    put:
      operationId: updateKPI
      parameters:
        - $ref: '#/components/parameters/KpiId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/UpdateKpiRequest'
      responses:
        '200':
          description: Updated KPI
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/KpiResponse'
    delete:
      operationId: deleteKPI
      parameters:
        - $ref: '#/components/parameters/KpiId'
      responses:
        '204': { description: Deleted }
        '409': { $ref: '#/components/responses/KpiLinked' }

  /goals/{goalId}/kpis:
    post:
      operationId: linkKPI
      parameters:
        - $ref: '#/components/parameters/GoalId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/LinkKpiRequest'
      responses:
        '200':
          description: Linked
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/GoalResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }

  /goals/{goalId}/kpis/{kpiId}:
    delete:
      operationId: unlinkKPI
      parameters:
        - $ref: '#/components/parameters/GoalId'
        - $ref: '#/components/parameters/KpiId'
      responses:
        '204': { description: Unlinked }
        '403': { $ref: '#/components/responses/PhaseRestricted' }

  /employees/{empId}/assignments:
    get:
      operationId: getAssignment
      parameters:
        - $ref: '#/components/parameters/EmpId'
      responses:
        '200':
          description: Assignment data
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AssignmentResponse'
    post:
      operationId: createAssignment
      parameters:
        - $ref: '#/components/parameters/EmpId'
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateAssignmentRequest'
      responses:
        '201':
          description: Created assignment
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/AssignmentResponse'
        '403': { $ref: '#/components/responses/PhaseRestricted' }

components:
  parameters:
    EmpId:
      name: empId
      in: path
      required: true
      schema: { type: string, format: uuid }
    CatId:
      name: catId
      in: path
      required: true
      schema: { type: string, format: uuid }
    GoalId:
      name: goalId
      in: path
      required: true
      schema: { type: string, format: uuid }
    KpiId:
      name: kpiId
      in: path
      required: true
      schema: { type: string, format: uuid }
    Cursor:
      name: cursor
      in: query
      schema: { type: string }
    Limit:
      name: limit
      in: query
      schema: { type: integer, default: 20, maximum: 100 }

  schemas:
    CreateCategoryRequest:
      type: object
      required: [name, weight]
      properties:
        name: { type: string, minLength: 1, maxLength: 100 }
        description: { type: string }
        weight: { type: number, minimum: 0.01, maximum: 100 }
    UpdateCategoryRequest:
      type: object
      required: [name, weight]
      properties:
        name: { type: string, minLength: 1, maxLength: 100 }
        description: { type: string }
        weight: { type: number, minimum: 0.01, maximum: 100 }
    CategoryResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        employeeId: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
        weight: { type: number }
        goals: { type: array, items: { $ref: '#/components/schemas/GoalResponse' } }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
    CategoryListResponse:
      type: object
      properties:
        items: { type: array, items: { $ref: '#/components/schemas/CategoryResponse' } }
        nextCursor: { type: string, nullable: true }
    CreateGoalRequest:
      type: object
      required: [name, unit, weight, targetValue]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string }
        unit: { type: string, enum: [porcentaje, moneda, numero] }
        weight: { type: number, minimum: 0.01, maximum: 100 }
        targetValue: { type: number, minimum: 0.01 }
        kpiIds: { type: array, items: { type: string, format: uuid }, maxItems: 5 }
    UpdateGoalRequest:
      type: object
      required: [name, unit, weight, targetValue, version]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        description: { type: string }
        unit: { type: string, enum: [porcentaje, moneda, numero] }
        weight: { type: number, minimum: 0.01, maximum: 100 }
        targetValue: { type: number, minimum: 0.01 }
        version: { type: integer }
        kpiIds: { type: array, items: { type: string, format: uuid }, maxItems: 5 }
    GoalResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        categoryId: { type: string, format: uuid }
        name: { type: string }
        description: { type: string }
        unit: { type: string }
        weight: { type: number }
        targetValue: { type: number }
        currentValue: { type: number }
        state: { type: string }
        version: { type: integer }
        kpis: { type: array, items: { $ref: '#/components/schemas/KpiResponse' } }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
    UpdateProgressRequest:
      type: object
      required: [currentValue]
      properties:
        currentValue: { type: number, minimum: 0 }
    BatchGoalRequest:
      type: object
      required: [items]
      properties:
        items:
          type: array
          maxItems: 50
          items:
            type: object
            required: [operation]
            properties:
              operation: { type: string, enum: [create, update] }
              categoryId: { type: string, format: uuid }
              goalId: { type: string, format: uuid }
              goal: { $ref: '#/components/schemas/CreateGoalRequest' }
    BatchGoalResponse:
      type: object
      properties:
        items: { type: array, items: { $ref: '#/components/schemas/GoalResponse' } }
    WeightValidationResponse:
      type: object
      properties:
        valid: { type: boolean }
        categorySum: { type: number }
        expectedSum: { type: number }
        deficit: { type: number }
        goalSums:
          type: array
          items:
            type: object
            properties:
              categoryId: { type: string, format: uuid }
              categoryName: { type: string }
              sum: { type: number }
              expectedSum: { type: number }
              deficit: { type: number }
    CreateKpiRequest:
      type: object
      required: [name, unit]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        unit: { type: string, enum: [porcentaje, moneda, numero] }
        description: { type: string }
    UpdateKpiRequest:
      type: object
      required: [name, unit]
      properties:
        name: { type: string, minLength: 1, maxLength: 255 }
        unit: { type: string, enum: [porcentaje, moneda, numero] }
        description: { type: string }
    KpiResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        name: { type: string }
        unit: { type: string }
        description: { type: string }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
    KpiListResponse:
      type: object
      properties:
        items: { type: array, items: { $ref: '#/components/schemas/KpiResponse' } }
        nextCursor: { type: string, nullable: true }
    LinkKpiRequest:
      type: object
      required: [kpiId]
      properties:
        kpiId: { type: string, format: uuid }
    AssignmentResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        employeeId: { type: string, format: uuid }
        cycleId: { type: string, format: uuid }
        categories: { type: array, items: { $ref: '#/components/schemas/CategoryResponse' } }
        createdAt: { type: string, format: date-time }
    CreateAssignmentRequest:
      type: object
      properties:
        cycleId: { type: string, format: uuid }

  responses:
    PhaseRestricted:
      description: Operation not allowed in current cycle phase
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    WeightInvalid:
      description: Weight validation failed
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    DuplicateCategoryName:
      description: Category name already exists for employee
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    ConcurrentModification:
      description: Optimistic locking version mismatch
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    KpiLinked:
      description: KPI cannot be deleted because it is linked to goals
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    RateLimit:
      description: Rate limit exceeded
      content:
        application/json:
          schema: { $ref: '#/components/schemas/Error' }
    Error:
      type: object
      properties:
        error:
          type: object
          properties:
            code: { type: string }
            message: { type: string }
            details: { type: array }
            trace_id: { type: string }
```

---

## 8. Concurrency Details

### 8.1 SELECT FOR UPDATE on Goal Creation

When creating a goal, the repository locks the parent `GoalCategory` row to serialize weight-sum calculations for that category. This prevents two concurrent requests from each seeing 80% used and inserting a 30% goal, resulting in 110% total.

```go
func (r *entRepo) LockCategory(ctx context.Context, catID string) (Category, error) {
    c, err := r.client.GoalCategory.Query().
        Where(goalcategory.ID(catID)).
        ForUpdate(). // Ent ForUpdate() translates to SELECT FOR UPDATE
        Only(ctx)
    // ...
}
```

Isolation: `Read Committed` is sufficient because the lock guarantees serial access to the category's weight sum.

### 8.2 Optimistic Locking on Goal Updates

The `Goal` schema includes a `version` integer. Every `UPDATE` increments it and includes `WHERE version = ?`. If no rows are affected, the service returns `CONCURRENT_MODIFICATION`.

```go
func (r *entRepo) UpdateGoal(ctx context.Context, goalID string, req UpdateGoalRequest) (Goal, error) {
    n, err := r.client.Goal.UpdateOneID(goalID).
        SetName(req.Name).
        SetDescription(req.Description).
        SetUnit(req.Unit).
        SetWeight(req.Weight).
        SetTargetValue(req.TargetValue).
        Where(goal.Version(req.Version)). // optimistic lock
        Save(ctx)
    if ent.IsNotFound(err) || n == 0 {
        return Goal{}, ErrConcurrentModification
    }
    // ...
}
```

### 8.3 Batch Transaction Pattern

Batch operations run inside a single `ent.Tx`. If any item fails validation or hits a constraint, the entire transaction rolls back. The service passes a `Repository` wrapper bound to the transaction so every sub-call participates in the same tx.

```go
func (r *entRepo) WithTx(ctx context.Context, fn func(Repository) error) error {
    tx, err := r.client.Tx(ctx)
    if err != nil {
        return err
    }
    defer func() {
        if r := recover(); r != nil {
            _ = tx.Rollback()
            panic(r)
        }
    }()

    if err := fn(&txRepo{tx: tx}); err != nil {
        _ = tx.Rollback()
        return err
    }
    return tx.Commit()
}
```

### 8.4 Advisory Locks on GoalAssignment

Creating a `GoalAssignment` uses a PostgreSQL advisory lock keyed by `hashtext(employee_id || cycle_id)` to prevent duplicate assignments if the user double-clicks or retries aggressively.

```go
func (r *entRepo) CreateAssignment(ctx context.Context, empID string, req CreateAssignmentRequest) (GoalAssignment, error) {
    lockID := hashEmployeeCycle(empID, req.CycleID)
    if err := r.execAdvisoryLock(ctx, lockID); err != nil {
        return GoalAssignment{}, err
    }
    // ... check existing, then insert
}
```

### 8.5 Idempotency Keys

Write endpoints accept `Idempotency-Key: <uuid>` in headers. The service stores `(key, payload_hash, response_snapshot)` in Redis with TTL 24h. If the same key arrives with a different payload hash, return `IDEMPOTENCY_KEY_REUSE`.

---

## 9. Testing Strategy

### 9.1 Unit Tests (`go test` + `gomock` or `mockery`)

| Target | Coverage Focus |
|--------|----------------|
| `goal_service.go` | 100% of phase gating paths, weight math edge cases, error mapping. |
| `goal_handler.go` | HTTP status mapping, request decoding, response encoding, middleware interaction. |
| `validation/weighting.go` | Epsilon comparisons, empty input, boundary values. |

### 9.2 Integration Tests (`testcontainers` PostgreSQL)

| Scenario | Description |
|----------|-------------|
| **Weight validation edge cases** | Categories sum to 99.99% (should fail), 100.01% (should fail), 100.0% (pass). Goals inside category at 99.99%, 100.01%. Empty category (no goals) should report sum 0. |
| **Phase restriction tests** | Attempt writes in `avance` and `cierre`; assert `PHASE_RESTRICTED` or `GOAL_NOT_DELETABLE_IN_PHASE`. Verify `PATCH /progress` rejected in `asignacion`. |
| **Batch atomicity** | Batch of 3 creates + 2 updates where the last update violates weight limit → entire batch fails, no rows modified in DB. |
| **Optimistic locking** | Two concurrent `PUT /goals/{id}` with same version; one succeeds, the second returns `409`. |
| **SELECT FOR UPDATE weight overflow** | 50 goroutines concurrently create goals in the same category with weight 50. Only 2 succeed; the rest receive `WEIGHT_SUM_INVALID`. |
| **Duplicate category name** | Create two categories with same name for same employee → second returns `409 DUPLICATE_CATEGORY_NAME`. |
| **KPI link limits** | Attempt to link 6 KPIs → `400 KPI_LINK_LIMIT_EXCEEDED`. Delete linked KPI → `409 KPI_LINKED_CANNOT_DELETE`. |
| **Assignment deduplication** | Two concurrent `POST /assignments` for same employee+cycle → one succeeds, the other finds existing row (idempotent). |

### 9.3 Load / Concurrency Tests

- **Race detector:** `go test -race` on all repository tests.
- **Concurrent weight creation:** Spawn 50 goroutines creating goals (weight=20) in the same category. Assert total weight never exceeds 100%+ε. Use `sync.WaitGroup` + `pgx` pool under `testcontainers`.
- **Batch performance:** 50-item batch must complete in <500ms on local PostgreSQL (measured in CI with `-bench`).

### 9.4 OpenAPI Validation

- Run `openapi-generator-cli validate` against the spec in CI.
- Generate TypeScript types via `openapi-typescript` and ensure zero type errors.

---

## 10. Dependencies

| Change | Role |
|--------|------|
| **C1: data-model-core** | Provides Ent schemas and migrations for `GoalCategory`, `Goal`, `KPI`, `GoalKpiLink`, `GoalAssignment`. |
| **B3: goals-and-weighting** | Functional spec: double 100%, custom categories, reusable KPIs, editing rules. |
| **C2: evaluation-lifecycle-api** | Source of truth for cycle phase (`asignacion`, `avance`, `cierre`). |
| **C7: auth-api** | Middleware injects `employeeId` and roles; this change assumes they exist in context. |

---

## 11. Non-Goals (Reiteration)

- **Evaluation scoring:** C6 handles final goal scoring and weighted averages.
- **Competencies:** C3 manages competency frameworks.
- **Auth/RBAC:** C7 provides identity; goals-api only checks phase and ownership.
- **Notifications:** C8 sends emails when goals change.
- **Excel import/export:** Out of scope; requires dedicated design for streaming parsing.
