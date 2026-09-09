# mis-evaluados-store Specification

## Purpose

Reactive Svelte 5 store for server-side paginated evaluatee list. Mirror of `rhEvaluadosStore.svelte.ts` pattern, consuming `GET /employees/{empId}/evaluatees?offset=&limit=&q=`.

**File**: `web/src/lib/stores/misEvaluadosStore.svelte.ts`

## TypeScript Types

```ts
// Reuses EmployeeListItem from openapi-fetch generated types
// Fields: id, firstName, lastName, email, employeeNumber, orgNodeId,
//         managerId, profileId, profileName, profileDescription, jobTitle, isActive

// Response envelope
interface EvaluateeListResponse {
  data: EmployeeListItem[];
  meta: { hasMore: boolean; limit: number; offset: number; total: number };
}
```

## Requirements

### Requirement: Reactive module-level state

Store SHALL expose `$state` fields: `items`, `loading`, `error`, `currentPage`, `apiTotal`, `hasMore`, `hasPrev`, `currentQ`. SHALL export getter functions matching `rhEvaluadosStore` naming.

#### Scenario: Initial state

- GIVEN store is loaded but `init()` not yet called
- WHEN any getter is read
- THEN `items` is `[]`, `loading` is false, `error` is null, `currentPage` is 0

### Requirement: init(employeeId) sets evaluator context

`init(employeeId: string)` SHALL store the current user's employee ID for all subsequent API calls. SHALL be called once from the page `onMount`.

#### Scenario: Init sets evaluator context

- GIVEN a logged-in user with employee ID `emp-123`
- WHEN `init('emp-123')` is called from `onMount`
- THEN subsequent `load()` calls request `GET /employees/emp-123/evaluatees`

### Requirement: load() fetches paginated data

`load()` SHALL call `GET /employees/{employeeId}/evaluatees?offset={currentPage * PAGE_SIZE}&limit=50&q={currentQ}`. On success SHALL set `items`, `hasMore`, `apiTotal`. On failure SHALL set `error` and clear `items`. `loading` SHALL be true during fetch.

#### Scenario: Successful page load

- GIVEN `currentPage=1`, `currentQ=undefined`
- WHEN `load()` completes successfully
- THEN `items` contains page 2 employees, `apiTotal` reflects server total, `hasPrev` is true

### Requirement: next() / prev() paginate

`next()` SHALL increment `currentPage` and call `load()` when `hasMore` is true. `prev()` SHALL decrement and call `load()` when `currentPage > 0`. Both SHALL be no-ops when guard condition fails.

#### Scenario: Next and prev paginate with guards

- GIVEN `currentPage=0` and `hasMore=true`
- WHEN `next()` is called
- THEN `currentPage` becomes 1 and `load()` fetches page 2
- AND WHEN `prev()` is called with `currentPage=0`
- THEN it is a no-op and no API call is made

### Requirement: search(query) with 300ms debounce

`search(query)` SHALL debounce 300ms, reset `currentPage` to 0, set `currentQ`, and call `load()`. Rapid keystrokes SHALL cancel previous timer and re-arm.

#### Scenario: Debounce coalesces keystrokes

- GIVEN user types "M", "a", "r" with 50ms gaps
- WHEN 300ms passes after last keystroke
- THEN only ONE API call is made with `q=Mar`

### Requirement: PAGE_SIZE constant

`PAGE_SIZE` SHALL be 50. Page-offset SHALL compute as `currentPage * PAGE_SIZE`.

#### Scenario: Page offset computation

- GIVEN `currentPage=2` and `PAGE_SIZE=50`
- WHEN `load()` builds the request URL
- THEN `offset=100` and `limit=50` are sent
