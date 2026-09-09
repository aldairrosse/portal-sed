# Proposal: Nine-Box Employee Visibility

## Intent

The 9×9 matrix renders correctly but the employee detail modal shows **empty names** — the API returns `NineBoxEntryDTO` with only `evaluateeId` (UUID), no display name or profile info. Additionally, the scope resolution uses `getChildren()` (direct reports only), so managers with multi-level subordinates cannot see all employees in their tree. The `RecomputeMatrix` service also assigns each employee as their own evaluator (self-evaluator placeholder), which does not reflect the real org hierarchy.

**User need**: See all subnode employees (not just direct reports) with their names visible in each 9-Box cell, matching the behavior of the goal assignment selector.

## Scope

### In Scope
- Add `employeeName` and `profileId` fields to `NineBoxEntryDTO` response
- Replace `getChildren()` with `getDescendants()` in scope resolution for `jefe` profile in 9×9 context
- Fix `RecomputeMatrix` to resolve the real evaluator from org chart (manager_id) instead of self-evaluator placeholder
- Add optional `quadrant` filter parameter to entries API (`GET /nine-box/...?quadrant=N`)
- Frontend: populate employee names in modal from enriched DTO; remove client-side name resolution

### Out of Scope
- Manual tier overrides by managers (tiers remain auto-computed)
- Matrix export (PDF/Excel)
- Cross-cycle comparison
- RH evaluation workflow (separate spec)
- Notification system for evaluation completion

## Capabilities

### New Capabilities
- `ninebox-employee-visibility`: Enriches 9-Box entries with employee display data (name, profileId) and fixes scope resolution to include all descendants in the hierarchy tree

### Modified Capabilities
- `manager-9x9`: Scope resolution changes from direct-only to descendants; DTO gains employeeName and profileId fields; entries API gains quadrant filter

## Approach

1. **Backend DTO enrichment**: Add `employeeName string` and `profileId uuid.UUID` to `NineBoxEntryDTO`. Populate via JOIN in `toEntryDTO()` using employee lookup (already available in repo).
2. **Scope fix**: In `ninebox_service.go`, replace `getChildren()` call with `getDescendants()` for profiles above `jefe` level. The `org-hierarchy` spec already defines `getDescendants()` — reuse it.
3. **Evaluator resolution**: In `RecomputeMatrix`, replace the self-evaluator placeholder loop with a query to `org_chart_nodes` resolving `manager_id` as the evaluator. Group evaluatees by their real manager.
4. **Quadrant filter**: Add optional `quadrant` query param to the entries list endpoint. Apply as SQL WHERE clause in repo.
5. **Frontend**: Remove client-side name resolution from `nineBoxStore.svelte.ts`. Use `employeeName` directly from API response in `NineBoxEntryCard.svelte` modal.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/dto/evaluation/evaluation_dto.go` | Modified | Add `employeeName`, `profileId` to `NineBoxEntryDTO` |
| `api/internal/service/evaluation/ninebox_service.go` | Modified | Fix evaluator resolution in `RecomputeMatrix`; enrich DTO mapping |
| `api/internal/repository/evaluation/ninebox_repo.go` | Modified | Add employee JOIN query; add quadrant filter to list query |
| `api/openapi/evaluations-and-9x9.yaml` | Modified | Add `employeeName`, `profileId` to NineBoxEntry schema; add `quadrant` query param |
| `web/src/lib/types/nine-box.ts` | Modified | Add `employeeName`, `profileId` to entry type |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Modified | Remove client-side name resolution; use API names |
| `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` | Modified | Display `employeeName` from DTO in modal |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `getDescendants()` returns large sets for director-general | Low | Existing repo method uses materialized path (efficient); add LIMIT guard if >500 |
| Evaluator resolution query adds latency to RecomputeMatrix | Low | Single JOIN on indexed `manager_id`; batch in existing transaction |
| Frontend breaks if API returns null `employeeName` during migration | Medium | Frontend falls back to `evaluateeId` substring until DTO is fully populated |

## Rollback Plan

Revert the DTO changes (remove `employeeName`/`profileId` fields) and restore `getChildren()` in scope resolution. The quadrant filter is additive and safe to keep. Frontend falls back to previous behavior (empty names) if DTO fields are absent.

## Dependencies

- `org-hierarchy` spec: `getDescendants()` already defined and implemented in `org_repo.go`
- `manager-9x9` spec: existing tier computation logic unchanged

## Success Criteria

- [ ] `NineBoxEntryDTO` response includes `employeeName` and `profileId` for all entries
- [ ] Modal in `NineBoxEntryCard.svelte` displays employee names (not empty)
- [ ] Director profile sees all descendants (not just direct reports) in matrix
- [ ] `RecomputeMatrix` assigns entries to the real manager's matrix (not self-evaluator)
- [ ] `GET /nine-box/...?quadrant=5` returns only entries in quadrant 5
- [ ] `pnpm run check` and `go test ./...` pass
