# Delta Spec: Nine-Box Employee Visibility

## New Capability: ninebox-employee-visibility

| Field | Detail |
|-------|--------|
| **Purpose** | Enrich 9-Box entries with employee name, profileId, and position. Fix scope resolution to use all descendants. Fix RecomputeMatrix evaluator assignment. |
| **Depends on** | `org-hierarchy` (getDescendants), `manager-9x9` (tier computation) |

### REQ-NB-001: Enriched NineBoxEntryDTO

The `NineBoxEntryDTO` response SHALL include `employeeName` (string) and `profileId` (UUID) populated via JOIN on `employees` table during `toEntryDTO()` mapping.

**API contract change:**

```yaml
# NineBoxEntryDTO — ADDED fields
employeeName: { type: string }
profileId:    { type: string, format: uuid }
```

#### Scenario: Entry returns employee name

- GIVEN matrix with entry for evaluatee "María García"
- WHEN `GET /nine-box/matrices/{matrixId}` returns
- THEN each entry includes `employeeName: "María García"` and `profileId: "<uuid>"`
- AND frontend renders name without client-side lookup

#### Scenario: Orphan entry (evaluatee deleted)

- GIVEN entry where evaluatee no longer exists in `employees`
- WHEN `toEntryDTO()` maps the entry
- THEN `employeeName` falls back to `""` and `profileId` to `uuid.Nil`
- AND entry still returns (no 500)

### REQ-NB-002: Consistent Scope Resolution via getDescendants()

Scope resolution for the 9×9 matrix SHALL use `getDescendants()` instead of `getChildren()` for all profiles. For `jefe` (whose subordinates are leaf collaborators), `getDescendants()` is functionally equivalent to `getChildren()`. For `director` and `director-general`, this includes all indirect reports.

(Previously: Scope was resolved with `getChildren()` — direct reports only. Director and DG saw incomplete trees.)

#### Scenario: Director sees all indirect reports

- GIVEN director "A" with 2 jefes and 5 collaborators under them
- WHEN 9×9 scope resolves for director
- THEN `getDescendants(directorNodeId)` returns 7 evaluatees (2 jefes + 5 collaborators)
- AND all appear in the matrix

#### Scenario: Jefe sees same scope (no behavioral change)

- GIVEN jefe with 3 leaf collaborators
- WHEN 9×9 scope resolves for jefe
- THEN `getDescendants(jefeNodeId)` returns 3 collaborators
- AND result is identical to prior `getChildren()`

### REQ-NB-003: Real Evaluator Resolution in RecomputeMatrix

`RecomputeMatrix` SHALL resolve the real evaluator from `org_chart_nodes.manager_id` instead of using the self-evaluator placeholder. Employees SHALL be grouped under their actual manager, not themselves.

(Previously: Each employee was assigned as their own evaluator — `evaluatorGroups[empID] = empID`.)

#### Scenario: Evaluator resolved from org chart

- GIVEN employee "María" whose `manager_id` points to "Carlos" (a jefe)
- WHEN `RecomputeMatrix(cycleId, phaseId)` runs
- THEN an entry for María is created/upserted in Carlos's matrix
- AND the matrix `evaluatorId` is "Carlos" (not María)

#### Scenario: Employee with no manager (root/DG)

- GIVEN employee at root of org chart (manager_id = NULL)
- WHEN `RecomputeMatrix` processes them
- THEN the employee is skipped for matrix placement (no evaluator to assign to)
- AND the operation completes without error

### REQ-NB-004: Quadrant Filter on Entries Endpoint

`GET /nine-box/matrices/{matrixId}/entries` SHALL accept an optional `quadrant` query parameter (integer, 1–9). When provided, the response SHALL include only entries matching that quadrant.

#### Scenario: Filter entries by quadrant

- GIVEN matrix with entries in quadrants 1, 5, and 9
- WHEN `GET /nine-box/matrices/{id}/entries?quadrant=5`
- THEN response includes only quadrant-5 entries
- AND other entries are excluded

#### Scenario: Quadrant parameter omitted (backward compatible)

- GIVEN matrix with entries across all quadrants
- WHEN `GET /nine-box/matrices/{id}/entries` (no quadrant param)
- THEN all entries return (unchanged behavior)

#### Scenario: Invalid quadrant value

- WHEN `?quadrant=0` or `?quadrant=10`
- THEN API returns 400 with `code: "INVALID_QUADRANT"`

## Modified Capability: manager-9x9

### MODIFIED Requirements

#### Requirement: Vista de matriz por evaluador (MODIFIED)

El sistema SHALL renderizar la matriz como grilla 3×3 con 9 cuadrantes coloreados. Cada empleado se posiciona según sus tiers calculados. El scope de evaluatees usa `getDescendants()` para todos los perfiles superiores: jefe, director, director-general. El `NineBoxEntryDTO` ahora incluye `employeeName` (string) y `profileId` (UUID) poblados desde `employees`.

(Previously: scope usaba `getChildren()` — direct reports only. DTO no incluía `employeeName` ni `profileId`.)

#### Scenario: Entry card shows employee name from API

- GIVEN entry with `employeeName: "María García"` and `profileId: "<uuid>"`
- WHEN user clicks entry dot in matrix
- THEN modal displays "María García" as the employee name
- AND shows profile label resolved from `profileId`
- AND avatar renders first letter from `employeeName`

#### Scenario: Director ve todos bajo su jerarquía (sin cambios de spec, corregido en código)

- GIVEN director con 2 jefes y sus colaboradores (7 personas total)
- WHEN accede a la matriz
- THEN ve los 7 evaluatees posicionados en la matriz
- AND puede drill-down a competencias de cualquier evaluatee

### REMOVED Requirements

None. All existing manager-9x9 requirements remain valid.

## Frontend Contract

| File | Change |
|------|--------|
| `web/src/lib/types/nine-box.ts` | `NineBoxEntry.employeeName` and `.profileId` now populated from API (already typed) |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | `normalizeApiData()` SHALL map `dto.employeeName` and `dto.profileId` directly — remove empty-string placeholder |
| `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` | Already consumes `entry.employeeName` and `entry.profileId` — NO structural changes needed |

#### Scenario: Frontend renders names without client-side lookup

- GIVEN API returns entries with `employeeName: "María García"`
- WHEN `nineBoxStore.load()` normalizes data
- THEN `NineBoxEntry.employeeName` = "María García" (not "")
- AND no additional fetch or client-side name resolution occurs

## Acceptance Criteria

- [ ] `GET /nine-box/matrices/{id}` response includes `employeeName` and `profileId` per entry
- [ ] `GET /nine-box/matrices/{id}/entries?quadrant=N` filters correctly
- [ ] Director profile sees all descendants in matrix (not just direct reports)
- [ ] `RecomputeMatrix` assigns entries to real manager's matrix (not self-evaluator)
- [ ] Modal in `NineBoxEntryCard.svelte` displays employee name (not empty)
- [ ] `pnpm run check` and `go test ./...` pass without regressions
