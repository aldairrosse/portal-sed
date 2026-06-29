## 1. OpenAPI and TypeScript Types

- [x] 1.1 Add `AreaMetrics` schema to `api/openapi/org-hierarchy.yaml`
- [x] 1.2 Add `GET /org-nodes/{nodeId}/area-metrics` endpoint to OpenAPI spec
- [x] 1.3 Generate TypeScript types: `openapi-typescript api/openapi/org-hierarchy.yaml -o web/src/lib/api/schemas/org-hierarchy.d.ts`
- [x] 1.4 Add `OrgHierarchyPaths` import to `web/src/lib/api/client.ts` and merge into `AppPaths`

## 2. Evaluator Scope Integration

- [x] 2.1 Add `fetchEvaluatorScope(evaluatorId, cycleId?)` function to `orgHierarchyStore`
- [x] 2.2 Modify `orgHierarchyStore.load()` to fetch scope before loading tree
- [x] 2.3 Add scope-based filtering: only show nodes in `scopeData.orgNodeIds`
- [x] 2.4 Handle scope fetch failure with error state

## 3. Role-Based View Routing

- [x] 3.1 Add profile detection from `devContext` store in hierarchy routes
- [x] 3.2 Create Jefe direct-reports table component for `/evaluacion/9x9/jerarquia`
- [x] 3.3 Add conditional rendering: jefe → table, director/DG/RH → tree
- [x] 3.4 Verify RH profile sees full tree in `/rh/jerarquia`

## 4. Area Metrics API Consumption

- [x] 4.1 Add `fetchAreaMetrics(nodeId, cycleId?)` function to `rhHierarchyStore`
- [x] 4.2 Replace local fixture computation with API call
- [x] 4.3 Handle API failure with EmptyState fallback
- [x] 4.4 Remove unused local aggregation functions from `rhHierarchyStore`

## 5. Verification

- [x] 5.1 Run `pnpm run check` — no new type errors
- [x] 5.2 Run `pnpm run lint` — no new lint errors
- [x] 5.3 Manual test: director sees filtered subtree (scope fallback: full tree when SCOPE_NOT_FOUND)
- [x] 5.4 Manual test: jefe sees direct reports table
- [x] 5.5 Manual test: RH sees full tree with real metrics
