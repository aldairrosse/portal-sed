# Proposal: Nine-Box Computed Matrix

## Intent

The 9×9 matrix is currently a **persisted snapshot** computed by an explicit `POST /nine-box/recompute/{cycleId}/{phaseId}` call, read back from `nine_box_matrices`/`nine_box_entries`. This creates several product gaps:

1. **Stale data**: GET returns whatever snapshot was last recomputed; it does not reflect the current phase of the cycle or new ratings/goal progress.
2. **No viewer scoping**: the evaluator is a query parameter; any authenticated user can read any evaluator's matrix (`TODO(auth:C7)` in the handler). There is no "jefe sees their team, RH/DG see everyone".
3. **Weighted self/HR ratings**: `ComputePotentialTier` averages self and HR ratings 50/50, and the self/HR split is inferred from timestamps (fragile). There is no `source` column on `evaluation_competency`. Target weights: rh = 0.8 / self = 0.2, business-tunable.
4. **No freshness bound**: the matrix never reflects late ratings or progress; the persisted snapshot is served indefinitely. Target: 1h TTL cache.
5. **Schema/migration drift**: the Ent schema defines the unique index as `(cycle_id, evaluator_id, phase_id)`, but the goose migration and live DB still have the old `idx_nine_box_matrixes_cycle_eval` **without** `phase_id`. The second matrix per phase (avance + cierre) fails with a duplicate key.

**User need**: The matrix must be **computed on demand** from live data — goal progress (performance) and competency ratings (potential) — showing the state based on the **current phase of the current cycle**, served from cache for up to 1 hour. A new cycle produces a new matrix. Results are filtered by the viewer's role: jefes see their team, director-general and RH see everyone. Competency potential weights jefe/RH ratings over self-evaluation (0.8/0.2, business-tunable), and ratings carry an explicit source.

## Scope

### In Scope
- Compute the matrix on read (on-demand derivation from goals + evaluation_competencies) using the **current phase** of the cycle; keep the persisted matrix only as an optional recomputable cache
- Add `source` (`self` | `rh`) to `evaluation_competency` and weight self vs jefe ratings in potential computation (configurable weights; default jefe > self for competencies)
- Scope matrix/entries reads by viewer: `jefe` and above → descendants of viewer's org node; `director-general` and `rh` → all employees; requesting an out-of-scope `evaluator_id` → 403
- Synchronize the DB index with the Ent schema: unique `(cycle_id, evaluator_id, phase_id)` via goose migration
- Expose raw inputs (goal progress %, self/HR ratings, weights, source) in `NineBoxEntryDTO` for transparency/audit
- Update `api/openapi/evaluations-and-9x9.yaml` to match

### Out of Scope
- Manual tier overrides by managers (tiers remain auto-computed)
- Matrix export (PDF/Excel)
- Cross-cycle comparison
- Notification system for evaluation completion
- UI changes in `web/` (API contract only; UI consumes new DTO fields in a follow-up)
- Replacing the self/HR timestamp heuristic outside `evaluation_competency.source`

## Capabilities

### New Capabilities
- `ninebox-computed-matrix`: On-demand computation of the 9×9 matrix from live goal progress and competency ratings, phase-driven, viewer-scoped, with explicit rating source and configurable weights

### Modified Capabilities
- `manager-9x9`: `RecomputeMatrix`/matrix reads become phase-aware and viewer-scoped; DTO exposes raw inputs
- `ninebox-employee-visibility`: scope resolution is absorbed into the viewer-scoping rule (jefe → descendants; RH/DG → all). After this change lands, `nine-box-employee-visibility` should be reviewed for archival.

## Approach

1. **DB migration**: add a goose migration that drops `idx_nine_box_matrixes_cycle_eval` and creates the unique index `(cycle_id, evaluator_id, phase_id)` matching the Ent schema.
2. **Rating source**: add `source` enum column (`self`/`rh`) to `evaluation_competency` (+ Ent schema field + migration). Populate on write from the evaluation flow (which side completed); fall back to the existing timestamp heuristic for existing rows.
3. **Computed read path**: `GET /nine-box/matrices?cycle_id=&phase_id=` serves through a new service `ComputeMatrixView(ctx, cycleID, viewerID, phaseID)`:
   - derive phase from `cycle.current_phase` when `phase_id` is omitted;
   - **1h TTL cache**: serve the persisted snapshot if `updated_at` is fresher than 1 hour, else re-derive and upsert;
   - resolve scope from viewer role/token (descendants vs all);
   - compute performance tier from goal progress, potential tier from weighted (rh 0.8 / self 0.2) competency ratings, quadrant via existing `pkg/quadrant`.
4. **Viewer scoping**: add RBAC/auth middleware to GET matrix/entries routes; `evaluator_id` outside scope → 403.
5. **OpenAPI**: extend `NineBoxEntryDTO` with raw inputs; document `source`; document 403 scope behavior.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/schema/evaluationcompetency.go` | Modified | Add `source` field |
| `api/internal/schema/nineboxmatrix.go` | Unchanged | Index already correct in Ent |
| `api/cmd/server/migrations/` | Modified | New migration: index fix + `source` column + backfill heuristic |
| `api/internal/repository/evaluation/ninebox_repo.go` | Modified | Scope filtering, source-aware rating queries, raw input queries |
| `api/internal/service/evaluation/ninebox_service.go` | Modified | `ComputeMatrixView`, weighted potential, phase derivation |
| `api/internal/pkg/quadrant/compute.go` | Modified | Weighted potential from source-aware ratings |
| `api/internal/handler/evaluation/evaluation_handler.go` | Modified | Viewer scoping (resolve TODO auth:C7), phase default |
| `api/internal/auth/rbac.go` | Modified (likely) | Permission/scope helper for nine-box reads |
| `api/openapi/evaluations-and-9x9.yaml` | Modified | DTO raw inputs, source, 403 scope behavior |
| `api/integration/*_test.go` | Modified | Tests for scoped reads, weighted potential, phase default |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| On-demand compute is N+1 per employee (goals + ratings) | Medium | Batch queries in repo; cache snapshot per (cycle, evaluator, phase) invalidated on phase transition; document performance budget |
| `source` backfill from timestamps is imperfect | Medium | Heuristic only for legacy rows; new rows always explicit |
| Changing read semantics breaks UI expecting persisted evaluatorId semantics | Medium | Keep response shape compatible; add fields, don't remove |
| RBAC scope regression (someone loses access) | High | Scope matrix from existing `scope=team` pattern in evaluation_repo; explicit tests for jefe/DG/RH |
