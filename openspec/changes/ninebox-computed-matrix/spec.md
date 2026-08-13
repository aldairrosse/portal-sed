# Delta Spec: Nine-Box Computed Matrix

## New Capability: ninebox-computed-matrix

| Field | Detail |
|-------|--------|
| **Purpose** | Compute the 9×9 matrix on demand from live goal progress and competency ratings, phase-driven from the current cycle, scoped by viewer role, with explicit rating source and configurable self/jefe weights. |
| **Depends on** | `org-hierarchy` (getDescendants), `manager-9x9` (tier computation), `evaluation-lifecycle` (cycle phases, evaluation ratings) |

### REQ-NBM-001: On-demand matrix computation

`GET /nine-box/matrices?cycle_id={cycleId}` SHALL compute matrix entries on demand from live data (goal progress + competency ratings) instead of reading a stored snapshot. When `phase_id` is omitted, the SHALL use `cycle.current_phase` of the requested cycle.

**API contract change:**

```yaml
# GET /nine-box/matrices — phase_id now OPTIONAL
parameters:
  - cycle_id: { required: true }
  - phase_id: { required: false }   # defaults to cycle.current_phase
  - evaluator_id: { required: false } # must be within viewer scope (REQ-NBM-003)
```

#### Scenario: Default phase from current cycle

- GIVEN cycle-2026 with `current_phase = avance` and a jefe with goal_assignments
- WHEN `GET /nine-box/matrices?cycle_id=cycle-2026` (no phase_id)
- THEN response uses phase `avance`
- AND entries reflect current goal progress and competency ratings at request time

#### Scenario: New cycle produces a new matrix

- GIVEN cycle-2027 exists with its own goal_assignments
- WHEN the same evaluator requests `GET /nine-box/matrices?cycle_id=cycle-2027`
- THEN the returned entries are computed for cycle-2027
- AND they do not mix cycle-2026 data

### REQ-NBM-002: Weighted potential from explicit rating source

`evaluation_competency` SHALL carry a `source` field (`self` | `rh`). `ComputePotentialTier` SHALL compute a weighted average of self and jefe/RH ratings (default weights: rh = 0.8, self = 0.2) before mapping to the 1–3 tier. The weights SHALL be configurable constants with a code note marking them as business-tunable (the final client may change them). Rows created before the migration SHALL be backfilled via the timestamp heuristic; rows created after SHALL set `source` explicitly.

#### Scenario: Jefe rating weighs more than self rating

- GIVEN employee with self competency rating 5 and jefe rating 3 (weights rh=0.8, self=0.2)
- WHEN potential tier is computed
- THEN weighted = 3*0.8 + 5*0.2 = 3.4 → tier 2 (≤3.66)
- AND the tier differs from the unweighted 50/50 average (4.0 → tier 3)

#### Scenario: Legacy rows backfilled by timestamps

- GIVEN an existing `evaluation_competency` row with no `source`
- WHEN the migration runs
- THEN `source` is set from `evaluations.self_evaluation_completed_at` vs `rh_evaluation_completed_at`
- AND the row is readable with an explicit source

### REQ-NBM-003: Viewer-scoped reads

Reads SHALL be scoped to the authenticated viewer:
- `rh` and `director-general` SHALL see all employees.
- `jefe` and above SHALL see employees whose org node is a descendant of the viewer's org node.

List endpoint (`GET /nine-box/matrices`) SHALL **filter** results to the viewer's scope and return `200` with the subset (empty list when nothing is in scope). Direct access to a specific matrix outside the viewer's scope (`GET /nine-box/matrices/{matrixId}` and `GET /nine-box/matrices/{matrixId}/entries`) SHALL return `403`.

#### Scenario: Jefe sees only their team

- GIVEN jefe "Carlos" whose org node has 3 descendant collaborators
- WHEN `GET /nine-box/matrices?cycle_id=C`
- THEN 200 with matrices/entries only for those 3 collaborators (no out-of-team evaluators)

#### Scenario: RH sees all

- GIVEN viewer with role `rh`
- WHEN `GET /nine-box/matrices?cycle_id=C`
- THEN 200 with matrices for every evaluator in the cycle

#### Scenario: Direct access to out-of-scope matrix

- GIVEN viewer whose scope does not include evaluator X
- WHEN `GET /nine-box/matrices/{matrixIdOfX}` or `GET /nine-box/matrices/{matrixIdOfX}/entries`
- THEN 403

> Note: the 403 on direct matrix access is tracked as `TODO(auth:C7)` in the handler (org-scope resolution not yet wired there); the list filter is implemented.

### REQ-NBM-004: DB index aligned with schema

The nine_box_matrixes unique index SHALL be `(cycle_id, evaluator_id, phase_id)`, matching the Ent schema, so one matrix per evaluator per cycle per phase is possible.

#### Scenario: Two phases, two matrices

- GIVEN a database migrated with the new index
- WHEN two matrices are inserted for the same (cycle_id, evaluator_id) with phase avance and cierre
- THEN both inserts succeed
- AND no duplicate-key error

### REQ-NBM-006: 1-hour TTL cache

Matrix reads SHALL serve the persisted snapshot when it is **fresher than 1 hour** for `(cycle_id, evaluator_id, phase_id)`; otherwise `ComputeMatrixView` SHALL re-derive entries from live data and upsert the cache. Freshness SHALL be determined from the matrix `updated_at`. `POST /nine-box/recompute` SHALL bypass the TTL and force re-derivation.

#### Scenario: Fresh snapshot served from cache

- GIVEN a matrix upserted 10 minutes ago for (cycle-2026, evaluator, avance)
- WHEN `GET /nine-box/matrices?cycle_id=cycle-2026` is called
- THEN the cached entries are returned without re-derivation

#### Scenario: Stale snapshot re-derived

- GIVEN the same matrix upserted 2 hours ago
- WHEN the same GET is called
- THEN entries are re-derived from current goal progress and ratings
- AND the cache is updated (new `updated_at`)

## Modified Capability: manager-9x9

### REQ-NBM-005: DTO exposes raw inputs

`NineBoxEntryDTO` SHALL include read-only fields: `goalProgressPercent` (float), `selfRating` (float|null), `hrRating` (float|null), `weights` (`{self, hr}`). Existing fields (`evaluateeId`, `employeeName`, `profileId`, tiers, quadrant) SHALL remain unchanged for backward compatibility.

#### Scenario: Entry exposes computation inputs

- GIVEN an entry computed for employee with goal progress 75% and ratings self=5, hr=3
- WHEN the entry DTO is returned
- THEN `goalProgressPercent: 75`, `selfRating: 5`, `hrRating: 3`, `weights: {self: 0.3, hr: 0.7}`
- AND existing fields are still present

## Out of Scope (explicit)

- Manual tier overrides
- Matrix export
- Cross-cycle comparison
- UI changes in `web/`
- Replacing the timestamp heuristic outside backfill
