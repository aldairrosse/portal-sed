# Tasks: Nine-Box Computed Matrix

## Task 1: DB migration — index + rating source

- [ ] Add goose migration `api/cmd/server/migrations/0000xx_ninebox_phase_unique.sql`:
  - Drop old unique index `idx_nine_box_matrixes_cycle_eval` on `nine_box_matrixes`
  - Create unique index `(cycle_id, evaluator_id, phase_id)` matching the Ent schema
  - Add `source` enum column (`self` | `rh`) to `evaluation_competencies`
  - Backfill `source` from `evaluations.self_evaluation_completed_at` vs `rh_evaluation_completed_at`
- [ ] Verify against live DB index name before writing DDL
- [ ] Run migration on a fresh test DB (5433) and confirm `TestMatrixByPhase_TwoPerEvaluator` passes the duplicate-key issue

## Task 2: Ent schema + generated code for `source`

- [ ] Add `source` field to `api/internal/schema/evaluationcompetency.go` (enum self/rh)
- [ ] `go generate ./...` (ent) in api/ and commit generated code
- [ ] Ensure auto-migrate stays consistent with the goose migration

## Task 3: Weighted potential in quadrant package

- [ ] Change `ComputePotentialTier` in `api/internal/pkg/quadrant/compute.go` to accept `(selfRating, hrRating float64, weightSelf, weightHR float64)` and use the weighted average before the 1–3 mapping
- [ ] Define default weights (rh=0.8, self=0.2) as package constants with a comment marking them business-tunable (`ponytail:`-style note: final client may change)
- [ ] Add unit tests for weighted tiers (3.4 → tier 2 example from REQ-NBM-002)

## Task 4: Source-aware rating queries

- [ ] In `api/internal/repository/evaluation/ninebox_repo.go`:
  - Replace/augment the timestamp-heuristic rating query with `source`-based grouping (self vs rh)
  - Batch goal progress + ratings per evaluatee set (avoid N+1)
- [ ] Keep legacy fallback for rows without source (shouldn't exist post-backfill)

## Task 5: ComputeMatrixView service + viewer scoping

- [ ] Add `ComputeMatrixView(ctx, cycleID, viewerID, phaseID)` in `api/internal/service/evaluation/ninebox_service.go`:
  - Default `phaseID` from `cycle.current_phase` when omitted
  - **TTL cache: serve persisted snapshot if `updated_at` < 1h old for (cycle, evaluator, phase); else re-derive**
  - Resolve viewer scope: `rh`/`director-general` → all; others → `getDescendants(viewerOrgNode)` (reuse org-hierarchy)
  - Compute perf tier from goal progress, potential tier from weighted ratings, quadrant from tiers
  - Upsert cache entries (existing `UpsertEntryByTiers`)
- [ ] `POST /nine-box/recompute` becomes an alias of the compute path (force refresh, bypasses TTL)
- [ ] Handler `evaluation_handler.go`: scope checks on GET matrices/entries (resolve `TODO(auth:C7)`); `evaluator_id`/matrix out of scope → 403; make `phase_id` optional

## Task 6: RBAC/scope helper

- [ ] In `api/internal/auth/rbac.go` (or a scope helper), expose whether a role sees all (`rh`, `director-general`) vs team-scoped
- [ ] Wire into the matrix GET routes middleware

## Task 7: OpenAPI contract

- [ ] `api/openapi/evaluations-and-9x9.yaml`:
  - `phase_id` optional on `GET /nine-box/matrices`
  - Add `goalProgressPercent`, `selfRating`, `hrRating`, `weights` to `NineBoxEntryDTO`
  - Document 403 for out-of-scope reads
- [ ] Regenerate/verify TS types if applicable (`web/src/lib/api/` — out of scope for code, note for UI follow-up)

## Task 8: Integration tests

- [ ] Update/extend `api/integration/ninebox_test.go` and `weighted_scoring_test.go`:
  - `TestMatrixByPhase_TwoPerEvaluator`: un-skip two-phase (depends on Task 1)
  - Scoped reads: jefe sees team, RH/DG see all, out-of-scope → 403 (REQ-NBM-003)
  - Default phase from current_cycle (REQ-NBM-001)
  - Weighted potential scenario (REQ-NBM-002)
  - DTO raw inputs present (REQ-NBM-005)
  - TTL cache: fresh snapshot served without re-derive; stale re-derived; recompute bypasses TTL (REQ-NBM-006)
- [ ] Ensure full suite passes: `go build ./...`, `go vet ./...`, `go test ./integration/...` against 5433

## Verification

- [ ] `openspec validate --all` (or manual review of artifacts)
- [ ] Integration suite green, including idempotent double-run
- [ ] No `web/` changes in this change
