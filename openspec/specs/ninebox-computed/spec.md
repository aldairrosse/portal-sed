# ninebox-computed Specification

## Purpose
Compute the nine-box matrix on demand from live data with viewer scoping and bounded freshness.

## Requirements

### Requirement: Matrix SHALL be computed on demand from live data with viewer scoping

`GET /nine-box/matrices?cycle_id={cycleId}` SHALL compute entries on demand from live goal progress and competency ratings, defaulting `phase_id` to `cycle.current_phase`, and SHALL scope evaluatees to the viewer's org descendants (`rh`/`director-general` see all). Freshness SHALL be bounded by a 1h TTL over the persisted cache (see REQ-NBM-001/003 below).

#### Scenario: Default phase from current cycle

- WHEN a jefe calls `GET /nine-box/matrices?cycle_id=cycle-2026` without `phase_id`
- THEN entries reflect live data for the cycle's `current_phase` scoped to the jefe's descendants

### Requirement: REQ-NBM-001: On-demand matrix computation

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

### Requirement: REQ-NBM-002: Weighted potential from explicit rating source

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

### Requirement: REQ-NBM-003: Viewer-scoped reads

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

### Requirement: REQ-NBM-004: DB index aligned with schema

The nine_box_matrixes unique index SHALL be `(cycle_id, evaluator_id, phase_id)`, matching the Ent schema, so one matrix per evaluator per cycle per phase is possible.

#### Scenario: Two phases, two matrices

- GIVEN a database migrated with the new index
- WHEN two matrices are inserted for the same (cycle_id, evaluator_id) with phase avance and cierre
- THEN both inserts succeed
- AND no duplicate-key error

### Requirement: REQ-NBM-006: 1-hour TTL cache

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

### Requirement: Fuente por fase para 9-box

El sistema SHALL construir el snapshot de avance solo con evaluaciones de avance (`avance_progress`) y el snapshot de cierre solo con evaluaciones de cierre (`cierre_progress`). Cada snapshot SHALL ser inmutable una vez que su fase deja de ser activa.

#### Scenario: Snapshot avance usa evaluaciones avance

- **WHEN** se calcula el 9-box de avance
- **THEN** cada meta aporta `avance_progress` y nunca `cierre_progress`

#### Scenario: Snapshot cierre usa evaluaciones cierre

- **WHEN** se calcula el 9-box de cierre
- **THEN** cada meta aporta `cierre_progress` y `avance_progress` queda intacto como histórico

### Requirement: Escalado de metas por direction en 9-box

El sistema SHALL escalar cada meta según su `direction` (ascendente/descendente) a % completado con clamp 0–100 y luego a escala 1–3 (`<34=1, <67=2, else 3`).

#### Scenario: Meta ascendente escala a tier

- **WHEN** una meta ascendente tiene 50% completado
- **THEN** el tier de desempeño es `2`

#### Scenario: Meta descendente invierte el cálculo

- **WHEN** una meta descendente con base `100` y actual `20` se evalúa
- **THEN** el % completado refleja la reducción (clamp 0–100)
- **AND** el tier resultante sigue la escala `<34=1, <67=2, else 3`

### Requirement: RecomputeMatrix con promedio ponderado ignorando ausentes

`RecomputeMatrix` SHALL usar el promedio ponderado `0.8 RH / 0.2 self` (igual que `ComputeMatrixView`) e SHALL ignorar los valores ausentes en vez de tratarlos como 0.

#### Scenario: Ponderado coincide con ComputeMatrixView

- **WHEN** un empleado tiene calificación RH `4` y self `2`
- **THEN** `RecomputeMatrix` calcula `4*0.8 + 2*0.2 = 3.6`
- **AND** el resultado coincide con `ComputeMatrixView` para los mismos insumos

#### Scenario: Valores ausentes se ignoran

- **WHEN** un empleado solo tiene calificación RH `4` (self ausente)
- **THEN** el promedio es `4` (solo RH)
- **AND** no se promedia con 0 por la ausencia de self
