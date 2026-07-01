# Design: Connect Evaluados RH Table to Real API

## Technical Approach

Create a new `rhEvaluadosStore` module using the simpler module-level `$state` pattern (matching `orgHierarchyStore`). It wraps `GET /api/v1/employees` with cursor pagination and debounced server-side search. `EmployeeEvaluationTable` gains a `mode` prop (`'rh' | 'manager'`) to discriminate rendering — RH mode reads `profileName` directly from the row, manager mode keeps the existing `goalsStore` lookup. The RH page swaps its data source from `goalsStore.getAssignments()` to the new store.

## Architecture Decisions

### Decision: Store shape — module-level $state

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Class-based `$state` (goalsStore pattern) | More structure, but overkill for a flat list + pagination | **No** |
| Module-level `$state` (orgHierarchyStore pattern) | Simpler, fewer lines, same reactivity | **Yes** |

**Rationale**: The store holds a flat list + cursor + search string. No nested normalization needed. Module-level `$state` is the lighter pattern already used in the codebase.

### Decision: Table component — mode prop instead of wrapper

| Option | Tradeoff | Decision |
|--------|----------|----------|
| New RH-specific table component | Duplicates 80% of table markup | **No** |
| Union type prop `rows: EmployeeAssignment[] \| EmployeeListItem[]` | Type narrowing needed per-row, messy | **No** |
| `mode` prop (`'rh' \| 'manager'`) with separate row props | Clean branching, one component | **Yes** |

**Rationale**: The table columns are identical; only the data source and cell rendering differ. A `mode` prop keeps the contract explicit and avoids type gymnastics.

### Decision: Missing OpenAPI fields — local type extension

| Option | Tradeoff | Decision |
|--------|----------|----------|
| Regenerate OpenAPI schema now | Backend change, out of scope for this frontend change | **No** |
| Local `interface EmployeeListItemExtended` with type assertion | Temporary, scoped to the store | **Yes** |

**Rationale**: The Go DTO returns `profileName`, `jobTitle`, `profileDescription` but the OpenAPI `Employee` schema omits them. A local interface in the store file bridges the gap until the schema is regenerated. Marked with a `// ponytail: remove when OpenAPI schema includes profileName` comment.

### Decision: Debounce — inline setTimeout

| Option | Tradeoff | Decision |
|--------|----------|----------|
| External debounce lib | New dependency for 5 lines | **No** |
| `setTimeout` / `clearTimeout` in the store | Zero deps, trivial | **Yes** |

## Data Flow

```
+page.svelte (RH)
    │
    ├─ onMount → rhEvaluadosStore.load()
    │                │
    │                └─ GET /api/v1/employees?limit=50
    │                        │
    │                        └─ items[], meta{nextCursor, hasMore}
    │
    ├─ search("María") → debounce 300ms → load(q="María")
    │
    ├─ next() → load(cursor=nextCursor)
    │
    └─ passes items[] + mode="rh" → EmployeeEvaluationTable
                                        │
                                        └─ renders rows from EmployeeListItem
                                           profileName direct from row
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/stores/rhEvaluadosStore.svelte.ts` | Create | Store with `load`, `next`, `prev`, `search`, reactive `items`/`loading`/`error`/`hasMore`/`hasPrev` |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modify | Add `mode` prop; when `'rh'`, accept `EmployeeListItem[]` rows and render `profileName` directly |
| `web/src/routes/rh/evaluaciones/+page.svelte` | Modify | Replace `goalsStore.getAssignments()` with `rhEvaluadosStore`; pass `mode="rh"` to table |

## Interfaces / Contracts

```typescript
// In rhEvaluadosStore.svelte.ts
// ponytail: remove when OpenAPI schema includes profileName/jobTitle
interface EmployeeListItemExtended {
  id: string;
  firstName: string;
  lastName: string;
  profileName: string;
  jobTitle?: string;
  isActive: boolean;
}

// Store exports
export function load(q?: string, cursor?: string): Promise<void>;
export function next(): Promise<void>;
export function prev(): Promise<void>;
export function search(query: string): void; // debounced 300ms

// Reactive state
export const items: EmployeeListItemExtended[];  // $state
export const loading: boolean;                    // $state
export const error: string | null;                // $state
export const hasMore: boolean;                    // $state
export const hasPrev: boolean;                    // $state
```

API response consumed (from OpenAPI `EmployeeListResponse`):
```typescript
{ data: Employee[], meta: { nextCursor?: string, hasMore: boolean, limit: number } }
```

## Testing Strategy

| Layer | What | How |
|-------|------|-----|
| Unit | Store pagination logic | Verify `next()`/`prev()` toggle `hasMore`/`hasPrev` correctly with mock responses |
| Unit | Debounce | Verify rapid `search()` calls produce single API call after 300ms |
| Manual | Table rendering | Confirm RH page shows `profileName` from API; manager page unchanged |
| Type check | `pnpm run check` | Catches any type mismatch from the extended interface |

## Migration / Rollout

No migration required. The new store is additive; the RH page simply switches its import. Manager view (`/evaluaciones`) is untouched.

## Open Questions

- [ ] OpenAPI schema for `Employee` is missing `profileName`, `jobTitle`, `profileDescription` — should we regenerate the schema as part of this change or keep the local type extension?
