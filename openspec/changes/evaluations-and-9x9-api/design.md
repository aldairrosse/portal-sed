# C6: evaluations-and-9x9-api — Technical Design

## 1. Overview

C6 implements the **year-end closing REST API** for SED Evaluación de Desempeño. It is the highest-traffic change in the system: during the annual closing window, **all employees submit evaluations simultaneously**, generating thousands to millions of requests in a short period.

The change delivers three parallel, independent evaluation paths (Decision #4 from B1/B5):

1. **Self-Evaluation** — Employee submits competency ratings (1–5) and goal closing comments.
2. **RH Formal Evaluation** — HR submits formal competency ratings and closes the evaluation.
3. **Manager 9×9 Matrix** — Manager rates evaluatees on performance (1–9) and potential (1–9); the quadrant (1–9) is computed on write and denormalized.

All endpoints are versioned under `/api/v1/`. Authentication and RBAC are out of scope for C6 and marked with `TODO(auth:C7)`.

---

## 2. Package Structure

```
api/internal/
├── handler/evaluation/
│   ├── evaluation_handler.go      # 8 evaluation endpoints
│   ├── evaluation_handler_test.go # Handler tests
│   ├── ninebox_handler.go         # 9 nine-box endpoints
│   ├── ninebox_handler_test.go    # Handler tests
│   └── routes.go                  # Chi router registration
├── service/evaluation/
│   ├── evaluation_service.go      # Self-eval, RH-eval, finalize logic
│   ├── ninebox_service.go         # 9×9 entry & batch logic
│   └── evaluation_service_test.go # Service layer tests
├── repository/evaluation/
│   ├── evaluation_repo.go         # Evaluation & competency CRUD + tx
│   ├── ninebox_repo.go            # Matrix & entry CRUD + tx
│   └── evaluation_repo_test.go    # Repository tests
└── pkg/
    ├── quadrant/
    │   └── compute.go             # Pure quadrant computation
    └── state/
        └── machine.go             # Evaluation state transition guards
```

**Rationale:**
- Handlers are thin: JSON decode, validate, call service, encode response.
- Services hold business rules: phase validation, state transitions, quadrant computation.
- Repositories encapsulate Ent queries and transaction boundaries.
- `pkg/quadrant` and `pkg/state` are pure, dependency-free packages for deterministic logic.

---

## 3. Schema Deltas (from C1 data-model-core)

The following fields and enum values are **required by C6** but not present in the C1 schema. They must be added before implementation.

| Entity | Delta | Reason |
|--------|-------|--------|
| `Evaluation` | Add `version int` (default 0) | Optimistic locking (`If-Match`) |
| `Evaluation` | Add `state` enum value `en_progreso` | Proposal state machine |
| `NineBoxEntry` | Add `version int` (default 0) | Optimistic locking on entry updates |
| `EvaluationCompetency` | Add `comments text` (optional) | Self-eval and RH-eval comments |
| PostgreSQL | Create materialized view `evaluation_summary` | Fast dashboard counts by state |
| PostgreSQL | Create GIN index on `evaluation_summary` if filtered by org | Tenant-scoped aggregation |

---

## 4. Handler Layer

### 4.1 Endpoint Registry

| # | Method | Path | Handler Function | Auth (TODO) |
|---|--------|------|-------------------|-------------|
| 1 | `GET` | `/api/v1/evaluations` | `ListEvaluations` | `rh`, `admin` |
| 2 | `GET` | `/api/v1/evaluations/{id}` | `GetEvaluation` | owner, manager, `rh` |
| 3 | `POST` | `/api/v1/evaluations/{id}/self-evaluation` | `SubmitSelfEvaluation` | owner |
| 4 | `PUT` | `/api/v1/evaluations/{id}/self-evaluation` | `UpdateSelfEvaluation` | owner |
| 5 | `POST` | `/api/v1/evaluations/{id}/rh-evaluation` | `SubmitRHEvaluation` | `rh`, `admin` |
| 6 | `PUT` | `/api/v1/evaluations/{id}/rh-evaluation` | `UpdateRHEvaluation` | `rh`, `admin` |
| 7 | `POST` | `/api/v1/evaluations/{id}/finalize` | `FinalizeEvaluation` | `rh`, `admin` |
| 8 | `GET` | `/api/v1/evaluations/summary` | `GetEvaluationSummary` | `rh`, `admin` |
| 9 | `GET` | `/api/v1/nine-box/matrices` | `ListMatrices` | evaluator, `rh` |
| 10 | `GET` | `/api/v1/nine-box/matrices/{matrixId}` | `GetMatrix` | evaluator owner, `rh` |
| 11 | `POST` | `/api/v1/nine-box/matrices` | `CreateMatrix` | evaluator, `rh` |
| 12 | `GET` | `/api/v1/nine-box/matrices/{matrixId}/entries` | `ListMatrixEntries` | evaluator owner, `rh` |
| 13 | `POST` | `/api/v1/nine-box/matrices/{matrixId}/entries` | `UpsertMatrixEntry` | evaluator owner |
| 14 | `PUT` | `/api/v1/nine-box/entries/{entryId}` | `UpdateEntry` | evaluator owner |
| 15 | `POST` | `/api/v1/nine-box/batch` | `BatchSubmitEntries` | evaluator owner |
| 16 | `GET` | `/api/v1/nine-box/scales` | `GetScales` | any authenticated |
| 17 | `GET` | `/api/v1/nine-box/quadrants` | `GetQuadrants` | any authenticated |

### 4.2 Request / Response DTOs (Go)

```go
// --- Shared ---
type ErrorResponse struct {
    Error struct {
        Code    string   `json:"code"`
        Message string   `json:"message"`
        Details []string `json:"details,omitempty"`
        TraceID string   `json:"trace_id"`
    } `json:"error"`
}

// --- Evaluations ---
type CompetencyRatingInput struct {
    CompetencyID uuid.UUID `json:"competencyId" validate:"required"`
    Rating       int       `json:"rating" validate:"min=1,max=5"`
    Comments     string    `json:"comments,omitempty"`
}

type GoalCommentInput struct {
    GoalID  uuid.UUID `json:"goalId" validate:"required"`
    Comment string    `json:"comment,omitempty"`
}

type SelfEvaluationRequest struct {
    Competencies []CompetencyRatingInput `json:"competencies" validate:"required,min=1,dive"`
    GoalComments []GoalCommentInput      `json:"goalComments,omitempty"`
}

type RHEvaluationRequest struct {
    Competencies  []CompetencyRatingInput `json:"competencies" validate:"required,min=1,dive"`
    FinalComments string                  `json:"finalComments,omitempty"`
}

type FinalizeEvaluationRequest struct {
    Reason string `json:"reason,omitempty"`
}

type EvaluationListResponse struct {
    Data       []EvaluationListItem `json:"data"`
    NextCursor string               `json:"nextCursor,omitempty"`
}

type EvaluationListItem struct {
    ID         uuid.UUID `json:"id"`
    EmployeeID uuid.UUID `json:"employeeId"`
    CycleID    uuid.UUID `json:"cycleId"`
    State      string    `json:"state"`
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}

type EvaluationDetailResponse struct {
    ID                       uuid.UUID              `json:"id"`
    EmployeeID               uuid.UUID              `json:"employeeId"`
    CycleID                  uuid.UUID              `json:"cycleId"`
    State                    string                 `json:"state"`
    SelfEvalCompletedAt      *time.Time             `json:"selfEvaluationCompletedAt,omitempty"`
    RHEvalCompletedAt        *time.Time             `json:"rhEvaluationCompletedAt,omitempty"`
    CompetencyRatings        []CompetencyRatingDTO  `json:"competencies"`
    GoalRatings              []GoalRatingDTO        `json:"goals"`
    Version                  int                    `json:"version"`
    CreatedAt                time.Time              `json:"createdAt"`
    UpdatedAt                time.Time              `json:"updatedAt"`
}

type CompetencyRatingDTO struct {
    CompetencyID uuid.UUID `json:"competencyId"`
    Rating       int       `json:"rating"`
    Comments     string    `json:"comments,omitempty"`
}

type GoalRatingDTO struct {
    GoalID        uuid.UUID `json:"goalId"`
    FinalRating   *int      `json:"finalRating,omitempty"`
    FinalComments string    `json:"finalComments,omitempty"`
}

type EvaluationSummaryResponse struct {
    CycleID uuid.UUID         `json:"cycleId"`
    Counts  map[string]int64  `json:"counts"` // key = state
}

// --- Nine-Box ---
type NineBoxMatrixResponse struct {
    ID          uuid.UUID           `json:"id"`
    CycleID     uuid.UUID           `json:"cycleId"`
    EvaluatorID uuid.UUID           `json:"evaluatorId"`
    Entries     []NineBoxEntryDTO   `json:"entries"`
    CreatedAt   time.Time           `json:"createdAt"`
    UpdatedAt   time.Time           `json:"updatedAt"`
}

type NineBoxEntryDTO struct {
    ID               uuid.UUID `json:"id"`
    EvaluateeID      uuid.UUID `json:"evaluateeId"`
    PerformanceScore int       `json:"performanceScore"`
    PotentialScore   int       `json:"potentialScore"`
    Quadrant         int       `json:"quadrant"`
    QuadrantLabel    string    `json:"quadrantLabel"`
    QuadrantColor    string    `json:"quadrantColor"`
    Comments         string    `json:"comments,omitempty"`
    Version          int       `json:"version"`
}

type NineBoxEntryInput struct {
    EvaluateeID      uuid.UUID `json:"evaluateeId" validate:"required"`
    PerformanceScore int       `json:"performanceScore" validate:"min=1,max=9"`
    PotentialScore   int       `json:"potentialScore" validate:"min=1,max=9"`
    Comments         string    `json:"comments,omitempty"`
}

type NineBoxBatchRequest struct {
    Entries []NineBoxEntryInput `json:"entries" validate:"required,min=1,max=20,dive"`
}

type NineBoxScaleDTO struct {
    Axis        string `json:"axis"`
    Level       int    `json:"level"`
    Label       string `json:"label"`
    Description string `json:"description"`
}

type NineBoxQuadrantDTO struct {
    Quadrant             int    `json:"quadrant"`
    Label                string `json:"label"`
    Description          string `json:"description"`
    Color                string `json:"color"`
    ActionRecommendation string `json:"actionRecommendation"`
}
```

### 4.3 Error Mapping

| Error Code | HTTP | Trigger |
|------------|------|---------|
| `EVALUATION_NOT_FOUND` | `404` | Evaluation ID does not exist |
| `MATRIX_NOT_FOUND` | `404` | Matrix ID does not exist |
| `ENTRY_NOT_FOUND` | `404` | Entry ID does not exist |
| `EVALUATION_ALREADY_FINALIZED` | `409` | Evaluation state == `completada` |
| `SELF_EVAL_DEADLINE_PASSED` | `409` | Cycle self-eval deadline expired |
| `INVALID_PHASE` | `409` | Cycle phase != `cierre` |
| `QUADRANT_OUT_OF_RANGE` | `400` | Score outside 1–9 |
| `UNAUTHORIZED_EVALUATOR` | `403` | User is not matrix owner |
| `CONCURRENT_UPDATE` | `409` | Optimistic lock failure (`version` mismatch) |
| `IDEMPOTENCY_KEY_CONFLICT` | `409` | Same key, different payload |
| `RATE_LIMIT_EXCEEDED` | `429` | Org limit breached |
| `REQUEST_TIMEOUT` | `408` | Context deadline exceeded |
| `SERVICE_UNAVAILABLE` | `503` | Circuit breaker open |

---

## 5. Service Layer

### 5.1 Interface

```go
package evaluation

// EvaluationService orchestrates self-evaluation, RH evaluation, and finalization.
type EvaluationService interface {
    SubmitSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req SelfEvaluationRequest, idempotencyKey string) (*EvaluationDetailResponse, error)
    UpdateSelfEvaluation(ctx context.Context, evaluationID uuid.UUID, req SelfEvaluationRequest, ifMatch int) (*EvaluationDetailResponse, error)
    SubmitRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req RHEvaluationRequest, idempotencyKey string) (*EvaluationDetailResponse, error)
    UpdateRHEvaluation(ctx context.Context, evaluationID uuid.UUID, req RHEvaluationRequest, ifMatch int) (*EvaluationDetailResponse, error)
    FinalizeEvaluation(ctx context.Context, evaluationID uuid.UUID, req FinalizeEvaluationRequest) (*EvaluationDetailResponse, error)
    GetEvaluation(ctx context.Context, id uuid.UUID) (*EvaluationDetailResponse, error)
    ListEvaluations(ctx context.Context, cycleID uuid.UUID, state string, cursor string, limit int) (*EvaluationListResponse, error)
    GetSummary(ctx context.Context, cycleID uuid.UUID) (*EvaluationSummaryResponse, error)
}

// NineBoxService handles matrix entries and quadrant computation.
type NineBoxService interface {
    CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*NineBoxMatrixResponse, error)
    GetMatrix(ctx context.Context, matrixID uuid.UUID) (*NineBoxMatrixResponse, error)
    ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]NineBoxMatrixResponse, error)
    UpsertEntry(ctx context.Context, matrixID uuid.UUID, req NineBoxEntryInput) (*NineBoxEntryDTO, error)
    UpdateEntry(ctx context.Context, entryID uuid.UUID, req NineBoxEntryInput, ifMatch int) (*NineBoxEntryDTO, error)
    BatchSubmitEntries(ctx context.Context, matrixID uuid.UUID, req NineBoxBatchRequest) ([]NineBoxEntryDTO, error)
    GetScales(ctx context.Context) ([]NineBoxScaleDTO, error)
    GetQuadrants(ctx context.Context) ([]NineBoxQuadrantDTO, error)
}
```

### 5.2 Critical Service Methods

#### `SubmitSelfEvaluation`

1. **Idempotency check** — Query Redis for `idempotency:{orgID}:{key}`. If hit and payload matches, return cached result. If hit and payload differs, return `IDEMPOTENCY_KEY_CONFLICT`.
2. **Phase validation** — Call `CycleService.GetPhase(cycleID)` (C2 dependency). If not `cierre`, return `INVALID_PHASE`.
3. **Deadline check** — If `cycle.self_evaluation_deadline < now`, return `SELF_EVAL_DEADLINE_PASSED`.
4. **Transaction** (isolation `REPEATABLE READ`):
   - `SELECT FOR UPDATE` on `Evaluation`.
   - If `state == completada`, return `EVALUATION_ALREADY_FINALIZED`.
   - Bulk upsert `EvaluationCompetency` rows (all ratings in one query).
   - Update `EvaluationGoal` closing comments.
   - Update `Evaluation.state` → `en_progreso` (if first submission) or keep current.
   - Set `self_evaluation_completed_at = now`.
   - Increment `version`.
5. **Post-commit** — Store response in Redis (TTL 24h) under idempotency key.
6. Return `EvaluationDetailResponse`.

#### `SubmitRHEvaluation`

Same flow as `SubmitSelfEvaluation`, but:
- Actor must have `rh` or `admin` role (`TODO(auth:C7)`).
- No deadline check.
- Sets `rh_evaluation_completed_at = now`.
- Updates `Evaluation.state` → `en_progreso` or keeps it.

#### `FinalizeEvaluation`

1. Acquire **PostgreSQL advisory lock** on `evaluation_id` (int64 hash of UUID).
2. Read evaluation. If `state == completada`, return `EVALUATION_ALREADY_FINALIZED`.
3. Verify cycle phase == `cierre`.
4. Transaction (`REPEATABLE READ`):
   - `SELECT FOR UPDATE` on `Evaluation`.
   - Update `state = completada`, set `updated_at`, increment `version`.
5. Refresh materialized view `evaluation_summary` (concurrently, if possible).
6. Release advisory lock.
7. Return finalized evaluation.

#### `UpsertEntry` (9×9)

1. Verify user owns the matrix (`TODO(auth:C7)`).
2. Validate scores 1–9.
3. Transaction:
   - `SELECT FOR UPDATE` on `NineBoxEntry` (or row to be inserted).
   - Compute `quadrant = ComputeQuadrant(performanceScore, potentialScore)`.
   - Fetch `NineBoxQuadrant` label/color for denormalized response fields.
   - Upsert entry with `quadrant`, `version++`.
4. Return `NineBoxEntryDTO` with denormalized `QuadrantLabel` and `QuadrantColor`.

#### `BatchSubmitEntries`

1. Validate all entries (max 20, unique `evaluateeId`s).
2. Transaction (isolation `REPEATABLE READ`):
   - Lock all existing entries for the matrix with `SELECT FOR UPDATE`.
   - For each entry: compute quadrant, upsert.
   - All-or-nothing: any failure rolls back the entire batch.
3. Target duration: < 300 ms for 20 entries.

---

## 6. Quadrant Computation

```go
// pkg/quadrant/compute.go
package quadrant

// ComputeQuadrant maps performance (1–9) and potential (1–9) scores to a quadrant (1–9).
// It is a pure, deterministic function with no side effects.
//
// Tiering:
//   1–3 → tier 1 (low)
//   4–6 → tier 2 (medium)
//   7–9 → tier 3 (high)
//
// Quadrant numbering (potential tier as rows, performance tier as columns):
//   potential=3, perf=1 → 7   potential=3, perf=2 → 8   potential=3, perf=3 → 9
//   potential=2, perf=1 → 4   potential=2, perf=2 → 5   potential=2, perf=3 → 6
//   potential=1, perf=1 → 1   potential=1, perf=2 → 2   potential=1, perf=3 → 3
//
// This numbering MUST match the seed data in the NineBoxQuadrant catalog table.
func ComputeQuadrant(performance, potential int) int {
    if performance < 1 || performance > 9 {
        panic("performance out of range")
    }
    if potential < 1 || potential > 9 {
        panic("potential out of range")
    }
    pt := tier(potential)
    perf := tier(performance)
    return (pt-1)*3 + perf
}

func tier(score int) int {
    switch {
    case score <= 3:
        return 1
    case score <= 6:
        return 2
    default:
        return 3
    }
}
```

**Why pure?** No database calls, no external state. Testable with an 81-case table. Cacheable at the call site if needed.

---

## 7. Repository Layer

### 7.1 Evaluation Repository Interface

```go
type EvaluationRepo interface {
    GetByID(ctx context.Context, id uuid.UUID) (*ent.Evaluation, error)
    ListByCycle(ctx context.Context, cycleID uuid.UUID, state string, cursor string, limit int) ([]*ent.Evaluation, string, error)
    GetDetail(ctx context.Context, id uuid.UUID) (*ent.Evaluation, error) // with CompetencyRatings + GoalRatings

    // Atomic submission: updates competencies, goals, and state in one tx.
    SubmitCompetenciesAndGoals(ctx context.Context, tx *ent.Tx, evalID uuid.UUID, comps []CompetencyUpsert, goals []GoalCommentUpsert) error

    Finalize(ctx context.Context, evalID uuid.UUID) error // uses advisory lock externally

    // Summary reads from materialized view.
    GetSummaryByCycle(ctx context.Context, cycleID uuid.UUID) (map[string]int64, error)
}
```

### 7.2 NineBox Repository Interface

```go
type NineBoxRepo interface {
    CreateMatrix(ctx context.Context, cycleID, evaluatorID uuid.UUID) (*ent.NineBoxMatrix, error)
    GetMatrixByID(ctx context.Context, id uuid.UUID) (*ent.NineBoxMatrix, error)
    ListMatrices(ctx context.Context, cycleID, evaluatorID uuid.UUID) ([]*ent.NineBoxMatrix, error)
    GetMatrixEntries(ctx context.Context, matrixID uuid.UUID) ([]*ent.NineBoxEntry, error)

    UpsertEntry(ctx context.Context, tx *ent.Tx, matrixID uuid.UUID, evaluateeID uuid.UUID, perf, pot int, quadrant int, comments string) (*ent.NineBoxEntry, error)
    UpdateEntry(ctx context.Context, tx *ent.Tx, entryID uuid.UUID, perf, pot int, quadrant int, comments string, version int) (*ent.NineBoxEntry, error)
    BatchUpsertEntries(ctx context.Context, tx *ent.Tx, matrixID uuid.UUID, entries []EntryUpsert) ([]*ent.NineBoxEntry, error)

    GetScales(ctx context.Context) ([]*ent.NineBoxScale, error)
    GetQuadrants(ctx context.Context) ([]*ent.NineBoxQuadrant, error)
}
```

### 7.3 Transaction Patterns

**Evaluation Submission Transaction:**
```go
func (r *evaluationRepo) SubmitCompetenciesAndGoals(ctx context.Context, tx *ent.Tx, evalID uuid.UUID, comps []CompetencyUpsert, goals []GoalCommentUpsert) error {
    // 1. Lock evaluation row
    ev, err := tx.Evaluation.Query().Where(evaluation.ID(evalID)).ForUpdate().Only(ctx)
    if err != nil { return err }

    // 2. Bulk upsert competencies (Ent CreateBulk + OnConflict)
    builders := make([]*ent.EvaluationCompetencyCreate, len(comps))
    for i, c := range comps {
        builders[i] = tx.EvaluationCompetency.Create().
            SetEvaluationID(evalID).
            SetCompetencyID(c.CompetencyID).
            SetRating(c.Rating).
            SetNillableComments(&c.Comments)
    }
    if err := tx.EvaluationCompetency.CreateBulk(builders...).
        OnConflictColumns("evaluation_id", "competency_id").
        UpdateNewValues().Exec(ctx); err != nil {
        return err
    }

    // 3. Update goal comments
    for _, g := range goals {
        if err := tx.EvaluationGoal.Update().
            Where(evalgoal.EvaluationID(evalID), evalgoal.GoalID(g.GoalID)).
            SetFinalComments(g.Comment).
            Exec(ctx); err != nil {
            return err
        }
    }

    // 4. State transition + version bump
    return tx.Evaluation.UpdateOne(ev).
        SetState(evaluation.StateEnProgreso).
        SetVersion(ev.Version + 1).
        Exec(ctx)
}
```

**NineBox Batch Upsert:**
```go
func (r *nineBoxRepo) BatchUpsertEntries(ctx context.Context, tx *ent.Tx, matrixID uuid.UUID, items []EntryUpsert) ([]*ent.NineBoxEntry, error) {
    // Lock existing entries to prevent concurrent quadrant overwrite
    _, err := tx.NineBoxEntry.Query().
        Where(nineboxentry.MatrixID(matrixID)).
        ForUpdate().
        All(ctx)
    if err != nil { return nil, err }

    builders := make([]*ent.NineBoxEntryCreate, len(items))
    for i, it := range items {
        builders[i] = tx.NineBoxEntry.Create().
            SetMatrixID(matrixID).
            SetEvaluateeID(it.EvaluateeID).
            SetPerformanceScore(it.PerformanceScore).
            SetPotentialScore(it.PotentialScore).
            SetQuadrant(it.Quadrant).
            SetNillableComments(&it.Comments)
    }

    err = tx.NineBoxEntry.CreateBulk(builders...).
        OnConflictColumns("matrix_id", "evaluatee_id").
        UpdateNewValues().Exec(ctx)
    if err != nil { return nil, err }

    // Re-fetch to return persisted rows with IDs
    return tx.NineBoxEntry.Query().
        Where(nineboxentry.MatrixID(matrixID)).
        All(ctx)
}
```

### 7.4 Materialized View

```sql
CREATE MATERIALIZED VIEW evaluation_summary AS
SELECT
    cycle_id,
    state,
    COUNT(*) AS count
FROM evaluations
GROUP BY cycle_id, state;

CREATE UNIQUE INDEX idx_evaluation_summary_cycle_state ON evaluation_summary(cycle_id, state);

-- Refresh trigger (or called explicitly after finalization)
REFRESH MATERIALIZED VIEW CONCURRENTLY evaluation_summary;
```

The `GetSummaryByCycle` repository method queries this view instead of aggregating live.

---

## 8. Year-End Burst Design

### 8.1 Database Connection Management

Configure the Ent client (pgx pool) with:

```go
poolConfig, _ := pgxpool.ParseConfig(databaseURL)
poolConfig.MaxConns = 40
poolConfig.MinConns = 15
poolConfig.MaxConnLifetime = time.Hour
poolConfig.MaxConnIdleTime = 30 * time.Minute
```

- **MinConns = 15** pre-warms connections before the peak window.
- **MaxConns = 40** caps total DB connections; remaining capacity is reserved for background jobs and replica reads.
- Expose Prometheus metrics: `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms`.

### 8.2 Queue-Based Write Smoothing

When concurrent evaluation submissions exceed **500 in-flight writes**, new `POST /evaluations/{id}/self-evaluation` and `POST /evaluations/{id}/rh-evaluation` requests are diverted to a **Redis-backed queue** instead of hitting PostgreSQL directly.

**Flow:**
1. Middleware counts in-flight writes (atomic counter in Redis: `writes:inflight:{orgID}`).
2. If count > 500:
   - Serialize request payload + idempotency key.
   - Push to Redis list `queue:evaluations:{orgID}`.
   - Return `202 Accepted` with `Retry-After: 5` header.
3. Worker pool (goroutines, 10 workers per org) drains the queue:
   - Batches up to **10 submissions** per transaction.
   - Priority sort: submissions with deadline < 24h are processed first (stored in a sorted set `queue:evaluations:priority:{orgID}`).
4. On worker completion, decrement in-flight counter.

**Deadline-aware priority:**
- Each queued item carries `deadline` timestamp.
- Workers pop from the priority sorted set (score = deadline asc) before the standard list.

### 8.3 Batch Aggregation for Competency Ratings

Instead of N individual `UPDATE` statements:
- Build a single `INSERT ... ON CONFLICT (evaluation_id, competency_id) DO UPDATE` query for all competency ratings in a submission.
- Ent `CreateBulk` + `OnConflict` generates this efficiently.
- Target: one round-trip per submission for competency writes.

### 8.4 Read Replicas

All `GET` endpoints (list, detail, scales, quadrants, summary) route to **read replicas** via a replica-aware Ent driver. Writes always use the primary.

### 8.5 Circuit Breaker

A Chi middleware monitors pool saturation:
- If `pgx_pool_conns_busy / MaxConns > 0.90` for > 10s:
  - Open circuit for write endpoints.
  - Return `503 SERVICE_UNAVAILABLE`.
  - Client retries with exponential backoff (2s, 4s, 8s, max 60s).
- Auto-recovery after 30s of healthy pool metrics.

---

## 9. OpenAPI 3.1 Specification

```yaml
openapi: 3.1.0
info:
  title: SED Evaluación y 9×9 API
  version: "1.0.0"
  description: |
    Year-end closing API for SED Evaluación de Desempeño.
    Handles self-evaluations, RH evaluations, and 9×9 manager matrices.
servers:
  - url: http://localhost:8080/api/v1
    description: Local development
paths:
  /evaluations:
    get:
      operationId: listEvaluations
      summary: List evaluations by cycle
      parameters:
        - name: cycle_id
          in: query
          required: true
          schema: { type: string, format: uuid }
        - name: state
          in: query
          schema: { type: string }
        - name: cursor
          in: query
          schema: { type: string }
        - name: limit
          in: query
          schema: { type: integer, default: 20, maximum: 100 }
      responses:
        "200":
          description: Paginated evaluation list
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationListResponse" }
        "429": { $ref: "#/components/responses/RateLimit" }

  /evaluations/{id}:
    get:
      operationId: getEvaluation
      summary: Get evaluation detail
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: Evaluation with competencies and goals
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "404": { $ref: "#/components/responses/NotFound" }

  /evaluations/{id}/self-evaluation:
    post:
      operationId: submitSelfEvaluation
      summary: Submit self-evaluation
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: Idempotency-Key
          in: header
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/SelfEvaluationRequest" }
      responses:
        "200":
          description: Updated evaluation
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "409": { $ref: "#/components/responses/Conflict" }
        "429": { $ref: "#/components/responses/RateLimit" }
    put:
      operationId: updateSelfEvaluation
      summary: Update self-evaluation
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: If-Match
          in: header
          required: true
          schema: { type: integer }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/SelfEvaluationRequest" }
      responses:
        "200":
          description: Updated evaluation
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "409": { $ref: "#/components/responses/Conflict" }

  /evaluations/{id}/rh-evaluation:
    post:
      operationId: submitRHEvaluation
      summary: Submit RH evaluation
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: Idempotency-Key
          in: header
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/RHEvaluationRequest" }
      responses:
        "200":
          description: Updated evaluation
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "409": { $ref: "#/components/responses/Conflict" }
        "429": { $ref: "#/components/responses/RateLimit" }
    put:
      operationId: updateRHEvaluation
      summary: Update RH evaluation
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: If-Match
          in: header
          required: true
          schema: { type: integer }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/RHEvaluationRequest" }
      responses:
        "200":
          description: Updated evaluation
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "409": { $ref: "#/components/responses/Conflict" }

  /evaluations/{id}/finalize:
    post:
      operationId: finalizeEvaluation
      summary: Finalize evaluation
      parameters:
        - name: id
          in: path
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        content:
          application/json:
            schema: { $ref: "#/components/schemas/FinalizeEvaluationRequest" }
      responses:
        "200":
          description: Finalized evaluation
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationDetailResponse" }
        "409": { $ref: "#/components/responses/Conflict" }
        "503": { $ref: "#/components/responses/ServiceUnavailable" }

  /evaluations/summary:
    get:
      operationId: getEvaluationSummary
      summary: Dashboard summary counts
      parameters:
        - name: cycle_id
          in: query
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: Counts by state
          content:
            application/json:
              schema: { $ref: "#/components/schemas/EvaluationSummaryResponse" }

  /nine-box/matrices:
    get:
      operationId: listMatrices
      summary: List 9×9 matrices
      parameters:
        - name: cycle_id
          in: query
          schema: { type: string, format: uuid }
        - name: evaluator_id
          in: query
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: Matrix list
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/NineBoxMatrixResponse" }
    post:
      operationId: createMatrix
      summary: Create a 9×9 matrix
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                cycleId: { type: string, format: uuid }
                evaluatorId: { type: string, format: uuid }
              required: [cycleId, evaluatorId]
      responses:
        "201":
          description: Created matrix
          content:
            application/json:
              schema: { $ref: "#/components/schemas/NineBoxMatrixResponse" }

  /nine-box/matrices/{matrixId}:
    get:
      operationId: getMatrix
      summary: Get matrix with entries
      parameters:
        - name: matrixId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: Matrix detail
          content:
            application/json:
              schema: { $ref: "#/components/schemas/NineBoxMatrixResponse" }
        "404": { $ref: "#/components/responses/NotFound" }

  /nine-box/matrices/{matrixId}/entries:
    get:
      operationId: listMatrixEntries
      summary: List entries in a matrix
      parameters:
        - name: matrixId
          in: path
          required: true
          schema: { type: string, format: uuid }
      responses:
        "200":
          description: Entry list
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/NineBoxEntryDTO" }
    post:
      operationId: upsertMatrixEntry
      summary: Upsert a matrix entry
      parameters:
        - name: matrixId
          in: path
          required: true
          schema: { type: string, format: uuid }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/NineBoxEntryInput" }
      responses:
        "200":
          description: Created or updated entry
          content:
            application/json:
              schema: { $ref: "#/components/schemas/NineBoxEntryDTO" }
        "400": { $ref: "#/components/responses/BadRequest" }

  /nine-box/entries/{entryId}:
    put:
      operationId: updateEntry
      summary: Update an existing entry
      parameters:
        - name: entryId
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: If-Match
          in: header
          required: true
          schema: { type: integer }
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/NineBoxEntryInput" }
      responses:
        "200":
          description: Updated entry
          content:
            application/json:
              schema: { $ref: "#/components/schemas/NineBoxEntryDTO" }
        "409": { $ref: "#/components/responses/Conflict" }

  /nine-box/batch:
    post:
      operationId: batchSubmitEntries
      summary: Batch submit 9×9 entries
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: "#/components/schemas/NineBoxBatchRequest" }
      responses:
        "200":
          description: Processed entries
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/NineBoxEntryDTO" }
        "400": { $ref: "#/components/responses/BadRequest" }
        "503": { $ref: "#/components/responses/ServiceUnavailable" }

  /nine-box/scales:
    get:
      operationId: getScales
      summary: Get 9×9 scale definitions
      responses:
        "200":
          description: Scale list
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/NineBoxScaleDTO" }
        "304": { description: Not Modified }

  /nine-box/quadrants:
    get:
      operationId: getQuadrants
      summary: Get 9×9 quadrant definitions
      responses:
        "200":
          description: Quadrant list
          content:
            application/json:
              schema:
                type: array
                items: { $ref: "#/components/schemas/NineBoxQuadrantDTO" }
        "304": { description: Not Modified }

components:
  schemas:
    EvaluationListResponse:
      type: object
      properties:
        data:
          type: array
          items: { $ref: "#/components/schemas/EvaluationListItem" }
        nextCursor: { type: string }

    EvaluationListItem:
      type: object
      properties:
        id: { type: string, format: uuid }
        employeeId: { type: string, format: uuid }
        cycleId: { type: string, format: uuid }
        state: { type: string }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }

    EvaluationDetailResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        employeeId: { type: string, format: uuid }
        cycleId: { type: string, format: uuid }
        state: { type: string }
        selfEvaluationCompletedAt: { type: string, format: date-time, nullable: true }
        rhEvaluationCompletedAt: { type: string, format: date-time, nullable: true }
        competencies:
          type: array
          items: { $ref: "#/components/schemas/CompetencyRatingDTO" }
        goals:
          type: array
          items: { $ref: "#/components/schemas/GoalRatingDTO" }
        version: { type: integer }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }

    CompetencyRatingDTO:
      type: object
      properties:
        competencyId: { type: string, format: uuid }
        rating: { type: integer, minimum: 1, maximum: 5 }
        comments: { type: string }

    GoalRatingDTO:
      type: object
      properties:
        goalId: { type: string, format: uuid }
        finalRating: { type: integer, minimum: 1, maximum: 5, nullable: true }
        finalComments: { type: string }

    SelfEvaluationRequest:
      type: object
      properties:
        competencies:
          type: array
          items: { $ref: "#/components/schemas/CompetencyRatingInput" }
        goalComments:
          type: array
          items: { $ref: "#/components/schemas/GoalCommentInput" }
      required: [competencies]

    CompetencyRatingInput:
      type: object
      properties:
        competencyId: { type: string, format: uuid }
        rating: { type: integer, minimum: 1, maximum: 5 }
        comments: { type: string }
      required: [competencyId, rating]

    GoalCommentInput:
      type: object
      properties:
        goalId: { type: string, format: uuid }
        comment: { type: string }
      required: [goalId]

    RHEvaluationRequest:
      type: object
      properties:
        competencies:
          type: array
          items: { $ref: "#/components/schemas/CompetencyRatingInput" }
        finalComments: { type: string }
      required: [competencies]

    FinalizeEvaluationRequest:
      type: object
      properties:
        reason: { type: string }

    EvaluationSummaryResponse:
      type: object
      properties:
        cycleId: { type: string, format: uuid }
        counts:
          type: object
          additionalProperties: { type: integer }

    NineBoxMatrixResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        cycleId: { type: string, format: uuid }
        evaluatorId: { type: string, format: uuid }
        entries:
          type: array
          items: { $ref: "#/components/schemas/NineBoxEntryDTO" }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }

    NineBoxEntryDTO:
      type: object
      properties:
        id: { type: string, format: uuid }
        evaluateeId: { type: string, format: uuid }
        performanceScore: { type: integer, minimum: 1, maximum: 9 }
        potentialScore: { type: integer, minimum: 1, maximum: 9 }
        quadrant: { type: integer, minimum: 1, maximum: 9 }
        quadrantLabel: { type: string }
        quadrantColor: { type: string }
        comments: { type: string }
        version: { type: integer }

    NineBoxEntryInput:
      type: object
      properties:
        evaluateeId: { type: string, format: uuid }
        performanceScore: { type: integer, minimum: 1, maximum: 9 }
        potentialScore: { type: integer, minimum: 1, maximum: 9 }
        comments: { type: string }
      required: [evaluateeId, performanceScore, potentialScore]

    NineBoxBatchRequest:
      type: object
      properties:
        entries:
          type: array
          items: { $ref: "#/components/schemas/NineBoxEntryInput" }
          maxItems: 20
      required: [entries]

    NineBoxScaleDTO:
      type: object
      properties:
        axis: { type: string, enum: [performance, potential] }
        level: { type: integer, minimum: 1, maximum: 9 }
        label: { type: string }
        description: { type: string }

    NineBoxQuadrantDTO:
      type: object
      properties:
        quadrant: { type: integer, minimum: 1, maximum: 9 }
        label: { type: string }
        description: { type: string }
        color: { type: string }
        actionRecommendation: { type: string }

    ErrorResponse:
      type: object
      properties:
        error:
          type: object
          properties:
            code: { type: string }
            message: { type: string }
            details:
              type: array
              items: { type: string }
            trace_id: { type: string }

  responses:
    NotFound:
      description: Resource not found
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorResponse" }
    BadRequest:
      description: Validation error
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorResponse" }
    Conflict:
      description: Concurrent update or business rule conflict
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorResponse" }
    RateLimit:
      description: Too many requests
      headers:
        Retry-After:
          schema: { type: integer }
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorResponse" }
    ServiceUnavailable:
      description: Circuit breaker open or pool saturated
      headers:
        Retry-After:
          schema: { type: integer }
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorResponse" }
```

---

## 10. Concurrency Details

### 10.1 Optimistic Locking

`Evaluation` and `NineBoxEntry` carry a `version` integer. Every mutating request must include:
- `If-Match: {version}` header for `PUT` and finalize operations.
- The repository increments `version` on successful update.
- If `version` mismatches, the service returns `409 CONCURRENT_UPDATE`.

**Why:** Prevents lost updates when the employee has the evaluation open in two browser tabs, or when RH and self-evaluation overlap.

### 10.2 SELECT FOR UPDATE

Used in:
- `SubmitSelfEvaluation` / `SubmitRHEvaluation` — locks the `Evaluation` row for the duration of the transaction.
- `UpsertEntry` / `BatchSubmitEntries` — locks all existing `NineBoxEntry` rows for the target matrix.

**Why:** Prevents stale quadrant computation when the manager updates the same evaluatee from multiple sessions.

### 10.3 Advisory Locks

`FinalizeEvaluation` acquires a PostgreSQL advisory lock before reading the evaluation:

```go
_, err := tx.ExecContext(ctx, "SELECT pg_advisory_lock($1)", hashUUID(evalID))
// ... finalize logic ...
_, _ = tx.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", hashUUID(evalID))
```

**Why:** Finalization is a one-way, irreversible action. The advisory lock guarantees exactly one successful finalization even if RH double-clicks the button or retries the request.

### 10.4 Idempotency Keys

- Clients send `Idempotency-Key: <uuid>` on `POST` self-evaluation and `POST` rh-evaluation.
- Server stores `{key: hash(payload), response: json}` in Redis with TTL 24h.
- On replay with matching payload hash: return cached response (200).
- On replay with different payload: return `409 IDEMPOTENCY_KEY_CONFLICT`.
- On first execution: run transaction, store result, return.

**Key format:** `idempotency:{orgID}:{idempotencyKey}`

### 10.5 Soft State Machine

State transitions use conditional updates without extra locking:

```sql
UPDATE evaluations
SET state = 'en_progreso', updated_at = now(), version = version + 1
WHERE id = $1 AND state = 'pendiente_evaluacion_final';
```

If `ROWS_AFFECTED == 0`, the evaluation is already in a different state; the service inspects the current state and returns the appropriate error (`EVALUATION_ALREADY_FINALIZED`, etc.).

---

## 11. Testing Strategy

### 11.1 Quadrant Computation

- **Table-driven test** covering all 81 combinations of `performance (1..9) × potential (1..9)`.
- Assert deterministic output for each pair.
- Assert panic on out-of-range inputs (0, 10, negative).
- Coverage target: 100% of `pkg/quadrant`.

### 11.2 State Machine

- Unit tests for every valid transition:
  - `pendiente_evaluacion_final` → `en_progreso` (on first submission)
  - `en_progreso` → `completada` (on finalize)
- Unit tests for invalid transitions:
  - `completada` → any (blocked)
  - Any state → `completada` without RH role (blocked, `TODO(auth:C7)`)
- Guard validation: phase must be `cierre`.

### 11.3 Concurrent Evaluation Submissions

- **Race test:** Spawn 100 goroutines. All call `SubmitSelfEvaluation` on the same evaluation ID with the same idempotency key.
- Expected: exactly 1 succeeds; 99 receive the cached idempotency result (or `CONCURRENT_UPDATE` if key differs).
- **Race test variant:** 50 goroutines with different idempotency keys. Expected: 1 wins the `SELECT FOR UPDATE`; the rest get `409 CONCURRENT_UPDATE`.
- Run with `go test -race`.

### 11.4 Batch Entry Tests

- Submit 20 valid entries atomically. Assert all exist and quadrants are correct.
- Submit 20 entries with one invalid score. Assert transaction rolls back; zero entries modified.
- Submit duplicate `evaluateeId`s in the same batch. Assert `400` validation error before touching the database.
- P95 latency assertion: < 300ms for 20 entries (local Postgres, warmed pool).

### 11.5 Load Tests

- `POST /api/v1/evaluations/{id}/self-evaluation` with **1000 concurrent submissions**.
- Target: all complete in < 30s, p95 latency < 500ms.
- Use `k6` or `go test` with `golang.org/x/sync/errgroup` and a real Postgres instance in Docker.

### 11.6 Idempotency Tests

- Submit evaluation with idempotency key `A`. Wait for 200.
- Re-submit identical payload with key `A`. Assert 200 and identical response body (including `version` and timestamps).
- Re-submit different payload with key `A`. Assert `409 IDEMPOTENCY_KEY_CONFLICT`.
- Wait 24h + 1s (or TTL mock). Re-submit with key `A`. Assert new execution.

### 11.7 Handler Tests

- Happy path for each of the 17 endpoints.
- Validation errors (missing required fields, out-of-range ratings).
- `404` when resource does not exist.
- `409` when evaluation already finalized.
- `409` when cycle phase is not `cierre`.
- `409` optimistic lock mismatch (`If-Match`).

---

## 12. Dependencies and Integration Points

| Change | Interface | Usage |
|--------|-----------|-------|
| **C1** | Ent schemas (`Evaluation`, `EvaluationCompetency`, `EvaluationGoal`, `NineBoxMatrix`, `NineBoxEntry`, etc.) | Direct repository consumption |
| **C2** | `CycleService.GetPhase(cycleID)` | Phase validation (`cierre` guard) |
| **C4** | `GoalService.GetGoalsByEvaluation(evaluationID)` | Read-only verification that goals exist before commenting |
| **C5** | `OrgService.ResolveEvaluator(employeeID)` | Validate manager/evaluator ownership (`TODO(auth:C7)`) |
| **C7** | Auth middleware + RBAC | All endpoints decorated with `TODO(auth:C7)` markers |

---

## 13. Non-Goals (Reminder)

As defined in `proposal.md`:
- No goal CRUD (C4).
- No competency catalog management (C3).
- No notifications/email (C7).
- No full auth/RBAC (C7 — only `TODO` markers).
- No aggregated cross-evaluator reports.
- No historical cycle comparison.
- No PDF/Excel export.
- No 9×9 for RH (manager-only).
- No rollback of finalized evaluations.

---

*Design generated for change C6: evaluations-and-9x9-api.*
*Source of truth: proposal.md + architecture.md + data-and-orm.md + data-model-core/design.md.*
