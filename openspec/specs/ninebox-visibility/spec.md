# ninebox-visibility Specification

## Purpose
Enrich nine-box entries with employee display data for direct rendering.

## Requirements

### Requirement: NineBoxEntryDTO SHALL include employee display data

The `NineBoxEntryDTO` response SHALL include `employeeName` (string) and `profileId` (UUID) populated via JOIN on `employees` during `toEntryDTO()` mapping, so the detail modal renders names without client-side lookup (see REQ-NB-001 below).

#### Scenario: Entry returns employee name

- WHEN `GET /nine-box/matrices/{matrixId}` returns an entry for evaluatee "María García"
- THEN the entry includes `employeeName: "María García"` and a valid `profileId`

### Requirement: REQ-NB-001: Enriched NineBoxEntryDTO

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

### Requirement: REQ-NB-002: Consistent Scope Resolution via getDescendants()

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

### Requirement: REQ-NB-003: Real Evaluator Resolution in RecomputeMatrix

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

### Requirement: REQ-NB-004: Quadrant Filter on Entries Endpoint

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
