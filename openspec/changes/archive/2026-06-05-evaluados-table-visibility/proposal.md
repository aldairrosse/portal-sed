# Proposal: Evaluados Table Visibility Enhancements

## Intent

Improve manager visibility into their team's evaluation status. Currently the "Mis evaluados" table shows only names, areas, and status badges. Managers need:
1. **Progress context** — aggregated goal progress per employee to quickly identify who is on/off track
2. **Incomplete emphasis** — clear visual distinction for employees who haven't completed their evaluations

Frontend-only changes using existing fixture data. No backend or API changes.

## Scope

### In Scope
- Add "Progreso global" column to `EmployeeEvaluationTable.svelte`
- Aggregate calculation: `sum(current_value) / sum(target_value)` per employee across all goals
- Reuse `ProgressIndicator.svelte` for the visual bar (color-coded: <40% red, <80% yellow, ≥80% green)
- Add summary indicator: "X de Y evaluados completaron"
- Visual emphasis for `pending` and `in-progress` rows (highlight, icon, or row tint)

### Out of Scope
- Backend API integration (fixture-based only)
- New store functions beyond getters
- Sorting/filtering by progress column
- Export or bulk actions

## Capabilities

### New Capabilities
- `evaluados-progress-column`: Shows aggregated goal completion per employee in the evaluados table.

### Modified Capabilities
- None — existing spec behavior unchanged, only UI presentation enhanced.

## Approach

### Enhancement 1: Global Goals Progress Column

1. In `EmployeeEvaluationTable.svelte`, add a helper function `getEmployeeGoalProgress(employeeId)` that:
   - Gets the employee's `EmployeeAssignment` from `assignments` store
   - Filters `goals` store to get goals matching those `goalIds`
   - Calculates: `Σ(goal.progress) / Σ(goal.targetValue) * 100` for percentage
   - Falls back to 0 if no goals

2. Add new `<th>Progreso global</th>` column between "Área" and "Estado"
3. Render `<ProgressIndicator value={progress} />` in the cell
4. Handle edge case: employee with no goals → show "—" dash

### Enhancement 2: Incomplete Evaluations Visibility

1. Add summary chip above the table: calculate completed count, render "X de Y evaluados completaron" using `getEvaluationStatus()` per employee
2. Add row highlight for `pending`/`in-progress` states:
   - Option A (preferred): `class="hover:bg-base-200 {!getStatus(emp.employeeId) === 'completed' ? 'bg-warning/10' : ''}"`
   - Option B: Add a subtle left border accent for pending rows
3. Reuse existing `EvaluationStatusBadge` — no new component needed

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modified | New column + row highlights + summary chip |
| `web/src/lib/components/goals/ProgressIndicator.svelte` | Reused | No changes |
| `web/src/lib/stores/goalsStore.svelte.ts` | Read-only | `getGoals()`, `getAssignments()` already exist |
| `web/src/lib/stores/evaluationStore.svelte.ts` | Read-only | `getEvaluationStatus()` already exists |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Performance: many employees with many goals | Low | $derived memoization handles reactivity; fixture data is small |
| Duplicate calculation on every render | Low | Use $derived for filteredEmployees to cache |

## Rollback Plan

1. Revert changes in `EmployeeEvaluationTable.svelte` to previous `<thead>` and `<tbody>` markup
2. No data migration needed (fixture-only)
3. No store changes to revert

## Dependencies

- `goalsStore.getAssignments()` — already returns `EmployeeAssignment[]` with `goalIds`
- `goalsStore.getGoals()` — already returns all goals with `progress` and `targetValue`
- `evaluationStore.getEvaluationStatus()` — already computes status
- `ProgressIndicator.svelte` — already handles color coding

## Success Criteria

- [ ] "Progreso global" column renders with `ProgressIndicator` for each employee
- [ ] Employees with no goals show "—" instead of broken percentage
- [ ] Summary chip shows correct "X de Y completaron" count
- [ ] Pending/in-progress rows have subtle visual distinction
- [ ] No new TypeScript errors or Svelte warnings