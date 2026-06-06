# Design: Evaluados Table Visibility Enhancements

## Technical Approach

Single-component modification to `EmployeeEvaluationTable.svelte`. All data comes from existing store getters (`getGoals`, `getAssignments`, `getEvaluationStatus`). No new stores, no new components — only reuse of `ProgressIndicator.svelte`. Reactivity handled via `$derived` memoization at the component level.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Progress memoization | `$derived` Map keyed by `employeeId` | Single reactive computation; recalculates only when `goals` or `employees` props change. Avoids per-row function calls in template. |
| Progress formula | `Σ(progress ?? 0) / Σ(targetValue)` per employee | Matches spec: aggregated across all assigned goals. `null` when `goalIds` empty or `Σ targetValue = 0`. |
| Summary chip placement | Above search input, inside the `flex flex-col gap-6` wrapper | Visible without scrolling; uses DaisyUI `badge` for consistency. |
| Row highlight strategy | Conditional `bg-warning/10` class on `<tr>` | Spec-mandated. Combined with existing `hover:bg-base-200` via Svelte class expression. |
| Store interaction | Read-only getters only | Spec requirement: no mutations. `getGoals()` and `getAssignments()` return `$state` arrays — Svelte 5 tracks reads inside `$derived`. |

## Data Flow

```
goalsStore ($state)          evaluationStore ($state)
  │ getGoals()                 │ getEvaluationStatus()
  │ getAssignments()           │
  ▼                            ▼
┌─────────────────────────────────────────┐
│  EmployeeEvaluationTable.svelte         │
│                                         │
│  $derived progressMap: Map<string, ...> │
│  $derived completionSummary: {n, total} │
│  $derived filteredEmployees (existing)  │
│                                         │
│  ┌──────────────────────────────────┐   │
│  │ Chip: "{n} de {total}..."        │   │
│  │ Table:                           │   │
│  │   <th>Progreso global</th>       │   │
│  │   <td><ProgressIndicator/></td>  │   │
│  │   <tr class={highlight}>         │   │
│  └──────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modify | Add progress column, summary chip, row highlighting |

No other files change. `ProgressIndicator.svelte`, `goalsStore`, `evaluationStore` remain untouched.

## Implementation Details

### 1. Progress Map (`$derived`)

```ts
import { getGoals, getAssignments } from '$lib/stores/goalsStore.svelte';

const goals = $derived(getGoals());
const assignments = $derived(getAssignments());

const progressMap = $derived(
  new Map(
    employees.map((emp) => {
      const empGoals = goals.filter((g) => emp.goalIds.includes(g.id));
      const sumTarget = empGoals.reduce((s, g) => s + g.targetValue, 0);
      if (empGoals.length === 0 || sumTarget === 0) return [emp.employeeId, null] as const;
      const sumProgress = empGoals.reduce((s, g) => s + (g.progress ?? 0), 0);
      return [emp.employeeId, (sumProgress / sumTarget) * 100] as const;
    })
  )
);
```

`null` = show "—" (no goals or zero target). Otherwise pass `value={pct}` to `ProgressIndicator` (max=100 default).

### 2. Summary Chip (`$derived`)

```ts
const completionSummary = $derived(() => {
  const completed = filteredEmployees.filter(
    (e) => getStatus(e.employeeId) === 'completed'
  ).length;
  return { completed, total: filteredEmployees.length };
});
```

Render as: `<span class="badge {pct >= 80 ? 'badge-success' : 'badge-warning'}">{completed} de {total} completaron</span>`

### 3. Column Insertion

Insert `<th>Progreso global</th>` between "Perfil" and "Estado" in `<thead>`. In `<tbody>`:

```svelte
<td>
  {#if progressMap.get(emp.employeeId) !== null}
    <ProgressIndicator value={progressMap.get(emp.employeeId)} />
  {:else}
    <span class="text-base-content/30">—</span>
  {/if}
</td>
```

### 4. Row Highlighting

```svelte
<tr class="hover:bg-base-200 {getStatus(emp.employeeId) !== 'completed' ? 'bg-warning/10' : ''}">
```

### 5. Reactivity Notes

- `goals` and `assignments` are `$derived` from store getters — Svelte 5 tracks the `$state` arrays they read.
- `progressMap` depends on `goals`, `assignments`, and `employees` (prop). Recalculates when any changes.
- `completionSummary` depends on `filteredEmployees` (which depends on `searchQuery` + `employees`) and `getStatus()` (which reads `competencyRatings` + `goalClosures` from `$state`).
- No `$effect` needed — all derived values are pure computations.
- Map lookup in template is O(1) per row — no performance concern for fixture-scale data.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Visual | Progress column renders correctly | Manual: verify bar + badge for employees with/without goals |
| Visual | Chip shows correct count | Manual: verify "X de Y" matches completed employees |
| Visual | Row highlighting | Manual: pending/in-progress rows have `bg-warning/10` tint |
| Edge case | Employee with no goals shows "—" | Manual: verify dash, no broken bar |
| Edge case | `Σ targetValue = 0` shows "—" | Manual: verify fallback |

## Open Questions

- None — all requirements covered by existing store API and component props.
