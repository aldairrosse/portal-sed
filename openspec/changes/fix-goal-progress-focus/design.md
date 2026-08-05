# Design: Fix Goal Progress Focus Loss + goal_progress_logs

## Technical Approach

Two independent changes that converge on the same store method:

1. **Frontend**: Convert `updateGoalProgress` from a reload-based pattern to optimistic local update (same pattern as `addGoalComment` at lines 472-492). Add 500ms debounce so rapid keystrokes generate a single PATCH. Guard against no-change calls by comparing against the goal's current `progress` in the store.

2. **Backend**: Add `goal_progress_logs` audit table via goose migration. Modify `UpdateGoalCurrentValue` to use a transaction: SELECT current_value, UPDATE conditionally (`WHERE current_value IS DISTINCT FROM $newValue`), INSERT log row only if the value changed.

No new API endpoints. No Ent schema (the table is write-only from application code). No UI for log history (YAGNI).

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Debounce location | Store (`updateGoalProgress`), not `GoalRow.svelte` | `GoalRow.svelte` is not touched per scope. The store already owns the PATCH call; adding a debounce wrapper there is minimal. |
| Optimistic update pattern | Same as `addGoalComment` (lines 472-492) | Proven pattern in the same file: mutate `storeState.data` locally, call API, rollback on error. No `reload()`. |
| Guard: same value | Compare against `goal.progress` in store | Already have `storeState.data.goals[n].progress`; avoid network round-trip to check. |
| Transaction in repo | `db.BeginTx` + SELECT + conditional UPDATE + INSERT | Atomicity: the log row and the UPDATE must happen together or not at all. |
| Conditional UPDATE | `WHERE current_value IS DISTINCT FROM $newValue` | Handles NULL correctly (unlike `!=`). When value unchanged, `RowsAffected() == 0` → skip INSERT, return current row. |
| No Ent schema | Raw SQL for migration + repo | `goal_progress_logs` is write-only from app code; no queries, no Ent codegen benefit. Follows existing pattern (`UpdateGoalCurrentValue` already uses raw SQL). |
| created_by nullable | `UUID NULL REFERENCES employees(id)` | System-triggered updates (e.g., from KPI sync) may have no user context. The FK is for integrity when a user is known. |
| No API exposure | Logs are internal-only | YAGNI: no UI or integration consumes this data yet. Adding an endpoint now would be speculative. |

## Data Flow

```
GoalRow.svelte (oninput / onchange)
  │  updateGoalProgress(goalId, newValue)
  ▼
goalsStore.updateGoalProgress (debounce 500ms)
  │  guard: if newValue === goal.progress → return (no-op)
  │  snapshot old value for rollback
  │  optimistic: mutate storeState.data.goals[n].progress = newValue
  │  PATCH /goals/{goalId}/progress { current_value: newValue }
  │  on error: restore old value, rethrow
  ▼
PATCH /goals/{goalId}/progress
  │
  ▼
ProgressService.UpdateGoalProgress
  │  phase check, ownership check
  ▼
GoalRepo.UpdateGoalCurrentValue (modified)
  │  BEGIN tx
  │  SELECT current_value FROM goals WHERE id = $1
  │  if current_value IS NOT DISTINCT FROM $newValue → COMMIT, return row (no change)
  │  UPDATE goals SET current_value = $1, updated_at = now(), version = version + 1
  │    WHERE id = $2 AND current_value IS DISTINCT FROM $1
  │  INSERT INTO goal_progress_logs (goal_id, value, created_by)
  │    VALUES ($1, $2, $3)
  │  COMMIT
  │  return updated GoalRow
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/stores/goalsStore.svelte.ts` | Modify | `updateGoalProgress`: optimistic local + debounce + same-value guard. No `reload()`. |
| `api/cmd/server/migrations/000027_goal_progress_logs.up.sql` | Create | Table `goal_progress_logs` with FK cascade, index |
| `api/cmd/server/migrations/000027_goal_progress_logs.down.sql` | Create | `DROP TABLE IF EXISTS goal_progress_logs` |
| `api/internal/repository/goal/goal_repo.go` | Modify | `UpdateGoalCurrentValue`: transaction + conditional UPDATE + INSERT log |
| `api/internal/service/goal/progress_service.go` | Modify | Pass `createdBy` (from session/context) through to repo for log insertion |

## Migration SQL

### Up

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS goal_progress_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    goal_id     UUID        NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    value       FLOAT       NOT NULL,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by  UUID        NULL REFERENCES employees(id)
);

CREATE INDEX IF NOT EXISTS idx_goal_progress_logs_goal_time
    ON goal_progress_logs (goal_id, recorded_at DESC);
```

### Down

```sql
-- +goose Down
DROP TABLE IF EXISTS goal_progress_logs;
```

## Repo: UpdateGoalCurrentValue (modified)

```go
func (r *GoalRepo) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64, createdBy *uuid.UUID) (*GoalRow, error) {
    tx, err := r.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback()

    // SELECT current value
    var existingValue float64
    err = tx.QueryRowContext(ctx, `SELECT current_value FROM goals WHERE id = $1`, goalID).Scan(&existingValue)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, pkgerrors.ErrGoalNotFound
        }
        return nil, err
    }

    // Only update if value actually changed
    now := time.Now()
    res, err := tx.ExecContext(ctx,
        `UPDATE goals SET current_value = $1, updated_at = $2, version = version + 1
         WHERE id = $3 AND current_value IS DISTINCT FROM $1`,
        currentValue, now, goalID,
    )
    if err != nil {
        return nil, err
    }

    affected, _ := res.RowsAffected()
    if affected > 0 {
        // Value changed → log it
        _, err = tx.ExecContext(ctx,
            `INSERT INTO goal_progress_logs (goal_id, value, created_by) VALUES ($1, $2, $3)`,
            goalID, currentValue, createdBy,
        )
        if err != nil {
            return nil, err
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    g, err := r.client.Goal.Query().Where(goal.ID(goalID)).Only(ctx)
    if err != nil {
        return nil, err
    }
    return r.goalToRow(ctx, g)
}
```

## Store: updateGoalProgress (modified)

```ts
let debounceTimers = new Map<string, ReturnType<typeof setTimeout>>();

export async function updateGoalProgress(goalId: string, progress: number): Promise<void> {
    // Clear pending debounce for this goal
    const existing = debounceTimers.get(goalId);
    if (existing) clearTimeout(existing);

    return new Promise((resolve, reject) => {
        const timer = setTimeout(async () => {
            debounceTimers.delete(goalId);

            const goal = storeState.data?.goals.find(g => g.id === goalId);
            if (!goal) return resolve();

            // Guard: no change
            if (goal.progress === progress) return resolve();

            // Snapshot for rollback
            const previousProgress = goal.progress;

            // Optimistic update
            storeState.data = {
                ...storeState.data!,
                goals: storeState.data!.goals.map(g =>
                    g.id === goalId ? { ...g, progress, progressUpdatedAt: new Date().toISOString() } : g
                )
            };

            try {
                const { error: apiError } = await client.PATCH('/goals/{goalId}/progress', {
                    params: { path: { goalId } },
                    body: { current_value: progress }
                });
                if (apiError) throw new Error(
                    (apiError as { error?: { message?: string } })?.error?.message ?? 'Error al actualizar progreso'
                );
                resolve();
            } catch (e) {
                // Rollback
                storeState.data = {
                    ...storeState.data!,
                    goals: storeState.data!.goals.map(g =>
                        g.id === goalId ? { ...g, progress: previousProgress } : g
                    )
                };
                reject(e);
            }
        }, 500);

        debounceTimers.set(goalId, timer);
    });
}
```

## Interfaces / Contracts

### Repo signature change

```go
// Before:
func (r *GoalRepo) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64) (*GoalRow, error)

// After:
func (r *GoalRepo) UpdateGoalCurrentValue(ctx context.Context, goalID uuid.UUID, currentValue float64, createdBy *uuid.UUID) (*GoalRow, error)
```

### Service: pass createdBy from context

The service already has access to `empID` (the authenticated user). It passes this as `createdBy` to the repo. The `ProgressService.UpdateGoalProgress` signature does not change; it reads `createdBy` from the `empID` parameter it already receives.

### Store signature: unchanged

```ts
export async function updateGoalProgress(goalId: string, progress: number): Promise<void>
```

Same signature. Internal implementation changes from reload pattern to optimistic local + debounce.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | `UpdateGoalCurrentValue` with same value | Mock tx: SELECT returns same value as new → UPDATE affects 0 rows → no INSERT → commit → verify log table untouched |
| Unit (Go) | `UpdateGoalCurrentValue` with changed value | Mock tx: SELECT returns different value → UPDATE affects 1 row → INSERT into log → commit |
| Unit (Go) | `UpdateGoalCurrentValue` goal not found | Mock tx: SELECT returns ErrNoRows → rollback → ErrGoalNotFound |
| Unit (TS) | `updateGoalProgress` same value | Vitest: call with same progress as store → no PATCH called |
| Unit (TS) | `updateGoalProgress` changed value | Vitest: verify store updated optimistically before PATCH resolves, PATCH called with correct body, no reload |
| Unit (TS) | `updateGoalProgress` PATCH fails | Vitest: mock PATCH rejection → verify store rolled back to previous value |
| Unit (TS) | Debounce merges rapid calls | Vitest: fast-forward timers, call 3 times → verify only 1 PATCH with last value |
| Integration | Migration up/down | Run goose up, verify table + index; goose down, verify table gone |
| E2E | Focus not lost | Playwright: type in progress input → verify DOM not replaced, focus stays |

## Migration / Rollout

- Migration `000027` is additive (new table only). No changes to existing tables.
- `UpdateGoalCurrentValue` signature changes → all callers updated (repo mock in tests, service, handler tests).
- Store change is internal to `updateGoalProgress`; no callers need changes.

## Open Questions

None. All decisions resolved per spec.
