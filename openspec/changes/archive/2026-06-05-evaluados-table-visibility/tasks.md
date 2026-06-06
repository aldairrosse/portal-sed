# Tasks: Evaluados Table Visibility Enhancements

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~50 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

## Phase 1: Data Computation ($derived)

- [x] 1.1 Add `goals` and `assignments` $derived from store getters at top of `<script>` block
- [x] 1.2 Add `progressMap` $derived Map keyed by employeeId, computing Σ(progress)/Σ(targetValue) with null fallback
- [x] 1.3 Add `completionSummary` $derived computing completed vs total from filteredEmployees

## Phase 2: Summary Chip

- [x] 2.1 Insert badge above search input: "{n} de {total} completaron" with dynamic color class
- [x] 2.2 Wire badge-success (≥80%) or badge-warning (<80%) based on completion ratio

## Phase 3: Progress Column

- [x] 3.1 Add `<th>Progreso global</th>` between "Perfil" and "Estado" in `<thead>`
- [x] 3.2 Add `<td>` with ProgressIndicator or "—" dash between Perfil and Estado in `<tbody>`

## Phase 4: Row Highlighting

- [x] 4.1 Add conditional `bg-warning/10` class to `<tr>` when `getStatus !== 'completed'`
- [x] 4.2 Preserve existing `hover:bg-base-200` on highlighted rows

## Phase 5: Verification

- [x] 5.1 Verify TypeScript compiles with no errors (`tsc --noEmit` or `pnpm run lint`)
- [x] 5.2 Verify Svelte 5 reactivity: progressMap recomputes when goals/employees change
- [ ] 5.3 Manual: employee with goals shows ProgressIndicator bar + badge
- [ ] 5.4 Manual: employee without goals shows "—"
- [ ] 5.5 Manual: chip shows correct "X de Y" for mixed states
- [ ] 5.6 Manual: pending rows have bg-warning/10 tint, completed rows do not

## Acceptance Criteria

- "Progreso global" column renders with ProgressIndicator for each employee with goals
- Employees with no goals or zero target show "—"
- Summary chip shows correct "X de Y completaron" with color-coded badge
- Pending/in-progress rows have bg-warning/10 background
- No new TypeScript errors or Svelte warnings
