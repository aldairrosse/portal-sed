# Design: Fix Competency Network Data Loading

## Technical Approach

Add a read-only endpoint that returns competency ratings for an arbitrary employee by `employee_id + cycle_id`, wire the store to call it when an `employeeId` argument is passed, and invoke `load(employeeId)` from the route's `+page.svelte` on mount. The existing `CompetencyNetworkView` is unchanged — it already reads from `evaluationStore` + `competencyStore` independently.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Endpoint shape | Flat competency ratings (not nested pillars) | `CompetencyNetworkView` already joins `competencyStore` (pillars/competencies) with `evaluationStore` (ratings). A nested response would duplicate data and require store restructuring. |
| Auth model | Reuse `RequireAuth` middleware; no new RBAC | Proposal scopes out new permissions. The endpoint returns the same data the user can already see via `GET /evaluations/{id}`. |
| Repo query style | Raw SQL (not Ent) | Follows existing pattern in `evaluation_repo.go` for cross-table joins. The query touches 3 tables (`evaluations` → `evaluation_competencies` → `competencies`) with a `pillar_id` join — raw SQL is clearer here. |
| Store backward compat | Optional `employeeId` param on `load()` | No-arg call preserves current behavior (logged-in user). |

## Data Flow

```
+page.svelte (mount)
  │  load(employeeId)
  ▼
evaluationStore.load(employeeId?)
  │  if employeeId → GET /evaluations/employee/{employeeId}?cycle_id=X
  │  else          → GET /evaluations/{id}  (existing path)
  ▼
EvaluationHandler.GetEmployeeCompetencies
  │  validate cycle_id, parse employeeId
  ▼
EvaluationService.GetEmployeeCompetencyRatings
  │  1. Find evaluation by (employee_id, cycle_id)
  │  2. Fetch competency ratings with competency + pillar info
  ▼
EvaluationRepo.GetCompetencyRatingsByEmployee
  │  SQL: JOIN evaluations → evaluation_competencies → competencies
  ▼
Response: { employeeId, cycleId, ratings: [{ competencyId, rating, comments }] }
  │
  ▼
normalizeApiData() → StoreData.competencyRatings
  │
  ▼
CompetencyNetworkView reads getCompetencyRatings(employeeId) + competencyStore
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/internal/handler/evaluation/evaluation_handler.go` | Modify | Add `GetEmployeeCompetencies` handler method |
| `api/internal/handler/evaluation/routes.go` | Modify | Register `GET /evaluations/employee/{employeeId}` |
| `api/internal/service/evaluation/evaluation_service.go` | Modify | Add `GetEmployeeCompetencyRatings` method |
| `api/internal/repository/evaluation/evaluation_repo.go` | Modify | Add `GetCompetencyRatingsByEmployee` raw SQL query |
| `api/internal/dto/evaluation/dto.go` | Modify | Add `EmployeeCompetencyRatingsResponse` DTO |
| `web/src/lib/stores/evaluationStore.svelte.ts` | Modify | `load(employeeId?)` — branch on param |
| `web/src/routes/evaluacion/9x9/competencias/[employeeId]/+page.svelte` | Modify | Call `evaluationStore.load(employeeId)` on mount + skeleton |
| `web/src/lib/components/evaluation/CompetencyNetworkSkeleton.svelte` | Create | DaisyUI skeleton matching view layout |
| `openspec/openapi.yaml` | Modify | Add endpoint schema |

## Interfaces / Contracts

### Endpoint: `GET /api/v1/evaluations/employee/{employeeId}?cycle_id={cycleId}`

**Response 200:**
```json
{
  "employeeId": "uuid",
  "cycleId": "uuid",
  "ratings": [
    {
      "competencyId": "uuid",
      "rating": 3,
      "comments": "text"
    }
  ]
}
```

**Response 400:** `{ "code": "MISSING_PARAM", "message": "cycle_id is required" }`
**Response 404:** `{ "code": "EVALUATION_NOT_FOUND", "message": "..." }` (no evaluation for that employee+cycle)

### Repo query (SQL)

```sql
SELECT ec.competency_id, ec.rating, ec.comments
FROM evaluations e
JOIN evaluation_competencies ec ON ec.evaluation_id = e.id
WHERE e.employee_id = $1 AND e.cycle_id = $2
ORDER BY ec.competency_id
```

Uses existing index `idx_evaluations_emp_cycle` on `(employee_id, cycle_id)` and `idx_eval_comp_eval_comp` on `(evaluation_id, competency_id)`.

### Store signature change

```ts
// Before:
export async function load(): Promise<void>

// After:
export async function load(employeeId?: string): Promise<void>
// if employeeId → GET /evaluations/employee/{employeeId}?cycle_id=...
// else → existing path (GET /evaluations/{id} with session user)
```

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | `GetCompetencyRatingsByEmployee` repo query | Table test with sqlmock: verify SQL args, empty result, multi-row |
| Unit (Go) | `GetEmployeeCompetencyRatings` service | Mock repo; test evaluation-not-found → 404 |
| Unit (TS) | `evaluationStore.load(employeeId)` | Vitest: mock `client.GET`, verify correct endpoint called with cycle_id |
| Integration | Handler `GetEmployeeCompetencies` | httptest: missing cycle_id → 400, valid → 200 with ratings |
| E2E | Route renders data | Playwright: navigate to `/evaluacion/9x9/competencias/{id}`, assert skeleton → data |

## Migration / Rollout

No migration required. Endpoint is additive; no schema changes. Store change is backward-compatible (optional param).

## Open Questions

- [ ] Should the endpoint return `rhRating` / `rhComment` alongside self-evaluation ratings? Current `evaluation_competencies` stores a single `rating` + `comments` — RH ratings may overwrite or live in a separate column. Need to verify which rating the view expects.
