# Design: Nine-Box Computed Matrix

## Context

Current flow (verified):

```
POST /api/v1/nine-box/recompute/{cycleId}/{phaseId}
  → RecomputeMatrix (service) 
    → GetGoalAssigneesByCycle → GetManagerMapping (evaluator = employee.manager_id)
    → GetMatrixByPhase / CreateMatrixWithPhase
    → GetGoalProgressByEmployee → ComputePerformanceTier
    → GetCompetencyRatingsByEmployee (self/HR by timestamp heuristic) → ComputePotentialTier (50/50)
    → ComputeQuadrantFromTiers → UpsertEntryByTiers (ON CONFLICT DO UPDATE)
GET /api/v1/nine-box/matrices[?cycle_id&phase_id&evaluator_id] → reads persisted snapshot only
```

Ent schema `nine_box_matrix` already has unique `(cycle_id, evaluator_id, phase_id)`; the goose migration and live DB still carry `idx_nine_box_matrixes_cycle_eval` without `phase_id` → duplicate-key on the second phase.

## Decisions

### D1. Source of truth: derive on read with 1h TTL cache

The matrix is a **view over live data**, not a primary store, served through a time-based cache.

- `GET /nine-box/matrices` and `GET /nine-box/matrices/{id}/entries` serve through `ComputeMatrixView`.
- The persisted `nine_box_matrices` / `nine_box_entries` tables are the **cache**: if a snapshot exists for `(cycle_id, evaluator_id, phase_id)` and is **fresher than 1 hour**, serve it; otherwise re-derive from live data and upsert (same upsert as today) so legacy consumers and the existing DB shape keep working. **TTL = 1h** (user decision).
- Phase resolution: if `phase_id` query param is absent, use `cycle.current_phase` (derived, not guessed). A different cycle is a different matrix by definition (cycle_id key).

Rationale: 1h TTL gives bounded staleness with amortized O(1) reads between refreshes; re-derivation is batched (one JOIN for goals per evaluatee set, one for ratings) to avoid per-employee round trips. Cache staleness marker: reuse the matrix `updated_at` (Audit/Time mixin).

### D2. Viewer scoping (resolves TODO auth:C7)

- `viewerID` comes from the authenticated token; role from RBAC.
- Scope rule:
  - `rh` and `director-general` → all employees (no org filter).
  - `jefe` and above → employees whose org node is a **descendant** of the viewer's org node (`getDescendants`, already available from org-hierarchy).
  - Root of the org chart (manager_id = NULL) has no evaluator and is skipped, as today.
- API behavior:
  - `GET /nine-box/matrices?evaluator_id=X` → if X is not within the viewer's scope and viewer is not RH/DG → `403`.
  - `GET /nine-box/matrices` without `evaluator_id` → only matrices whose evaluator is within scope.
  - `GET /nine-box/matrices/{matrixId}/entries` → verify matrix's evaluator is within scope → else `403`.
- Reuse the existing `scope=team` pattern in `evaluation_repo.go` (manager_id based) rather than inventing a new query.

### D3. Explicit rating source + weighted potential

- Add `source` enum to `evaluation_competency`: `self` | `rh` (SQL enum + Ent schema field + migration).
- Backfill: for existing rows, derive from the existing timestamp heuristic (compare `evaluations.self_evaluation_completed_at` vs `rh_evaluation_completed_at`); new writes always set it explicitly.
- Weights: configurable constants (defaults `rh = 0.8`, `self = 0.2` for competencies) in the quadrant package or service config; no new config table unless a second consumer appears (YAGNI). **Note for the team**: leave a `ponytail:`-style comment marking these weights as business-tunable — the final client may change them; editing the constants is the intended lever.
- `ComputePotentialTier` changes signature to accept `(selfRating, rhRating, weightSelf, weightRH)` — weighted average, then the existing 1–3 mapping.
- Performance tier keeps using goal progress (`(current/target)` or `(baseline-current)/baseline` for descending) — unchanged semantics.

### D4. DB migration

One goose migration (new file, e.g. `0000xx_ninebox_phase_unique.sql`):
1. `ALTER TABLE nine_box_matrixes DROP CONSTRAINT/DROP INDEX idx_nine_box_matrixes_cycle_eval;` (exact name from live DB)
2. `CREATE UNIQUE INDEX ... ON nine_box_matrixes (cycle_id, evaluator_id, phase_id);` (name matching Ent convention)
3. `ALTER TABLE evaluation_competencies ADD COLUMN source ... ;` with backfill UPDATE.

Ent auto-migrate must stay consistent with the migration (schema already declares the 3-column unique, so no schema change needed for the index; schema change only for `source`).

### D5. API contract

`NineBoxEntryDTO` gains (all optional/read-only):
- `source` (self|rh) per rating — actually exposed as `selfRating`, `hrRating`
- `goalProgressPercent` (float)
- `performanceTier`, `potentialTier`, `quadrant` (already present)
- `weights` (self, hr) used for potential

OpenAPI: add fields to schema; document `403` for out-of-scope reads; keep all existing fields for backward compatibility.

## Data flow (target)

```
GET /nine-box/matrices?cycle_id=C[&phase_id=P][&evaluator_id=E]
  → auth middleware → viewerID + role
  → resolve scope (D2)
  → if E given and E ∉ scope and role ∉ {rh, dg} → 403
  → phase P = P or cycle.current_phase
  → evaluatees = scope ∩ employees with goal_assignments(cycle C)
  → for each: goalProgress (batched) → perfTier; ratings self/rh (batched) → weighted potTier
  → quadrant = ComputeQuadrantFromTiers
  → upsert cache entries (D1)
  → return DTOs (D5)
```

## Decisions (user-confirmed)

1. Weights: **rh = 0.8 / self = 0.2** (configurable constants; note for client adjustment).
2. Cache: re-derive on **1h TTL** — serve snapshot if fresh, else recompute + upsert.
3. `POST /nine-box/recompute` stays as a manual force-refresh alias of the compute path.
