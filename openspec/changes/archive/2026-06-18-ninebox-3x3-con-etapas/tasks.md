# Tasks: NineBox 3×3 con Etapas

## Review Workload Forecast

Decision needed before apply: No
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: Medium

| PR | Scope | Est. lines |
|----|-------|-----------|
| PR 1 | Schema + migration + seed | ~400 |
| PR 2 | Compute logic + service + repo | ~500 |
| PR 3 | API handlers + OpenAPI | ~400 |
| PR 4 | UI frontend (3×3, modal, phase) | ~600 |
| PR 5 | Integration + WCAG + verify | ~300 |

## PR 1 — Schema + Migration + Seed

- [x] 1.1 `nineboxmatrix.go`: +`phase_id` UUID FK; unique index → `(cycle, evaluator, phase)`
- [x] 1.2 `nineboxentry.go`: −scores 1–9; +`performance_tier`(1–3), +`potential_tier`(1–3)
- [x] 1.3 `nineboxquadrant.go`: +`title`, +`color_hex`; `color` → Optional
- [x] 1.4 `phasedefinition.go`: +edge to NineBoxMatrix
- [x] 1.5 ~Migration~ (auto-migrated via Ent schema; manual ALTER SQL deferred to deployment) — Ent regeneration reflects new schema
- [x] 1.6 Seed: 9 quadrants with title + color_hex; entries use tiers; matrices include phase_id
- [x] 1.7 PhaseDefinition seed already has `avance`/`cierre` for 2026 ✓ — verified in `seed/cycle.go` lines 50-52

## PR 2 — Compute Logic + Service + Repository

- [x] 2.1 `compute.go`: +`ComputePerformanceTier`, +`ComputePotentialTier`, +`ComputeQuadrantFromTiers`
- [x] 2.2 `compute_test.go`: table-driven — boundaries, no-data, partial, all 9 combos
- [x] 2.3 DTOs: EntryDTO uses tiers; MatrixResponse +PhaseID; +QuadrantUpdateInput
- [x] 2.4 `ninebox_service.go`: +`RecomputeMatrix`, +`GetMatrixByPhase`, +`UpdateQuadrantByNumber`
- [x] 2.5 `ninebox_repo.go`: +`CreateMatrixWithPhase`, +`GetMatrixByPhase`, +`UpsertEntryByTiers`
- [x] 2.6 Deprecate `UpsertEntry`/`BatchSubmitEntries` (methods preserved with deprecation comments)

## PR 3 — API + OpenAPI

- [x] 3.1 OpenAPI: tier DTOs; +PUT quadrants, +POST recompute, +GET quadrants
- [x] 3.2 `UpdateQuadrant` handler: validate → call service → return updated
- [x] 3.3 `RecomputeMatrix` handler: cycleId/phaseId → return matrix
- [x] 3.4 `GetQuadrants` handler: return 9 quadrants with title/colorHex (already existed, no changes needed)
- [x] 3.5 Modify `ListMatrices`: accept `phaseId` query param
- [x] 3.6 Modify `ListMatrixEntries`: return tiers not scores (DTO already updated, handler delegates to service)
- [x] 3.7 Update Chi routes: wire new, remove old upsert/slider routes
- [x] 3.8 Handler tests for each new/modified endpoint

## PR 4 — UI Frontend

- [x] 4.1 `nine-box.ts`: tiers (1–3); quadrant `colorHex`/`title`; matrix `phaseId`
- [x] 4.2 `NineBoxMatrix.svelte`: 3×3 grid with colorHex; employee dots by tiers
- [x] 4.3 `NineBoxEntryCard.svelte`: show tiers; read-only; no sliders
- [x] 4.4 `NineBoxCellConfig.svelte` (new): modal — RH edits title/desc/colorHex
- [x] 4.5 Delete `NineBoxSliders.svelte`
- [x] 4.6 `nineBoxStore.svelte.ts`: load by cycle+phase; remove manual score
- [x] 4.7 `+page.svelte`: phase selector; dynamic labels "Avance"/"Evaluación"
- [x] 4.8 Update fixtures for tier structure

## PR 5 — Integration + Polish + Verify

- [x] 5.1 Integration: POST recompute — seed goals + ratings, verify tiers
- [x] 5.2 Integration: PUT quadrant — persist title/colorHex, GET returns updated
- [x] 5.3 Integration: matrix by phase — verify two per evaluator
- [x] 5.4 WCAG 2.1 AA: contrast on hex cells (luminance-based textColor); keyboard nav (roving tabindex + aria-activedescendant); focus mgmt (aria-modal, Escape trap); aria-live announcements
- [x] 5.5 Verify all 8 spec acceptance criteria (7/8 full, 1 partial — "pendientes" not implemented)
- [x] 5.6 Clean: no dead score references remain in fixtures; deprecated DTO/methods preserved for migration compatibility; routes_test updated with new nine-box routes
