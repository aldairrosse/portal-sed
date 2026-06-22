# Design: NineBox 3×3 con Etapas

## Technical Approach

Transformar la matriz 9×9 manual (scores 1–9 por eje, 81 celdas) en una matriz 3×3 automática con tiers 1–3 derivados de datos existentes. Los tiers se calculan desde avance de metas (desempeño) y ratings de competencias (potencial). La matriz se instancia por evaluador × etapa (`PhaseDefinition`). Los 9 cuadrantes son editables por RH (title, description, colorHex).

## Architecture Decisions

### Decision: GoalAssignment no necesita campo `progress`

**Choice**: Derivar progreso desde `Goal.current_value / Goal.target_value` (campos que ya existen en el schema `goal.go`).
**Alternatives considered**: Agregar `progress` float a `GoalAssignment`.
**Rationale**: `Goal` ya tiene `current_value`, `target_value`, `direction` y `baseline_value`. El progreso se calcula como porcentaje de avance. Agregar un campo redundante rompe single source of truth. La fórmula: `progress = (current_value / target_value) * 100` para direction=ascendente; invertir para descendente.

### Decision: NineBoxScale se mantiene sin cambios

**Choice**: No modificar ni eliminar `NineBoxScale`. Se conserva para compatibilidad histórica.
**Alternatives considered**: Eliminar la tabla o simplificarla a 3 niveles.
**Rationale**: El proposal lo explicita como out of scope. Datos históricos de seed permanecen. La nueva lógica de tiers 1–3 vive en `pkg/quadrant/compute.go` como funciones nuevas, no reutiliza NineBoxScale.

### Decision: Potential tier desde EvaluationCompetency.rating (1–5)

**Choice**: Calcular potential tier promediando `EvaluationCompetency.rating` de las evaluaciones self y RH del empleado en el ciclo+etapa.
**Alternatives considered**: Usar un campo separado de potencial; usar solo hr_rating.
**Rationale**: `EvaluationCompetency` ya tiene `rating` (1–5). La self-evaluación y la evaluación RH generan registros separados en la misma tabla, distinguibles por `evaluation_id` → `Evaluation.self_evaluation_completed_at` vs `rh_evaluation_completed_at`. Promediar ambos da una visión completa del potencial.

### Decision: Unique index (cycle_id, evaluator_id, phase_id)

**Choice**: Agregar `phase_id` al unique index existente de `NineBoxMatrix`.
**Alternatives considered**: Crear tabla separada por etapa.
**Rationale**: Mismo evaluador puede tener dos matrices (avance + cierre) en un ciclo. El index compuesto garantiza unicidad sin tabla extra.

### Decision: Quadrant formula se mantiene

**Choice**: `quadrant = (potentialTier - 1) * 3 + performanceTier` — misma fórmula que ya usa `ComputeQuadrant`.
**Alternatives considered**: Nueva fórmula de mapeo.
**Rationale**: La fórmula actual ya mapea (perf, potential) → 1–9 correctamente. Solo cambian los inputs (de 1–9 a 1–3). Reutilizar la lógica existente minimiza riesgo.

## Data Flow

```
Goal.current_value / Goal.target_value
        │
        ▼
  avgProgress ──→ ComputePerformanceTier ──→ performanceTier (1–3)
                                                    │
                                                    ▼
EvaluationCompetency.rating                  ComputeQuadrant ──→ quadrant (1–9)
  (self + RH, escala 1–5)                         ▲
        │                                         │
        ▼                                         │
  avgPotential ──→ ComputePotentialTier ──→ potentialTier (1–3)
```

```
RecomputeMatrix(cycleId, phaseId)
  │
  ├── 1. Get evaluators with evaluatees in cycle
  ├── 2. For each evaluator: GetOrCreate NineBoxMatrix(cycle, evaluator, phase)
  ├── 3. For each evaluatee:
  │      ├── Compute performanceTier from Goal progress
  │      ├── Compute potentialTier from EvaluationCompetency ratings
  │      ├── Compute quadrant from tiers
  │      └── Upsert NineBoxEntry(performanceTier, potentialTier, quadrant)
  └── 4. Return updated matrix
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/internal/schema/nineboxmatrix.go` | Modify | +`phase_id` UUID FK; unique index → `(cycle_id, evaluator_id, phase_id)`; +edge to PhaseDefinition |
| `api/internal/schema/nineboxentry.go` | Modify | −`performance_score`, −`potential_score`; +`performance_tier` int(1–3), +`potential_tier` int(1–3); `quadrant` se mantiene |
| `api/internal/schema/nineboxquadrant.go` | Modify | +`title` string; +`color_hex` string; `color` → Optional (backward compat); `description` ya existe |
| `api/internal/schema/phasedefinition.go` | Modify | +edge `nine_box_matrices` → NineBoxMatrix |
| `api/internal/pkg/quadrant/compute.go` | Modify | +`ComputePerformanceTier(avgProgress float64) int`; +`ComputePotentialTier(avgRating float64) int`; +`ComputeQuadrantFromTiers(perfTier, potTier int) int`; mantener `ComputeQuadrant` para compat |
| `api/internal/pkg/quadrant/compute_test.go` | Modify | +Tests para las 3 funciones nuevas con casos borde (sin datos, partial, límites) |
| `api/internal/dto/evaluation/evaluation_dto.go` | Modify | `NineBoxEntryDTO`: −PerformanceScore, −PotentialScore; +PerformanceTier, +PotentialTier. `NineBoxMatrixResponse`: +PhaseID. `NineBoxQuadrantDTO`: +Title, +ColorHex. +`NineBoxQuadrantUpdateInput` |
| `api/internal/service/evaluation/ninebox_service.go` | Modify | +`RecomputeMatrix(ctx, cycleId, phaseId)`; +`GetMatrixByPhase(ctx, cycleId, evaluatorId, phaseId)`; +`UpdateQuadrant(ctx, id, input)`; modificar `toEntryDTO` para tiers; deprecar `UpsertEntry`/`BatchSubmitEntries` (ya no son manuales) |
| `api/internal/repository/evaluation/ninebox_repo.go` | Modify | +`CreateMatrixWithPhase`; +`GetMatrixByPhase`; +`UpsertEntryByTiers`; +`BulkUpsertEntries`; ajustar SQL para nuevos campos |
| `api/internal/handler/evaluation/routes.go` | Modify | +`PUT /nine-box/quadrants/{id}`; +`POST /nine-box/recompute/{cycleId}/{phaseId}`; ajustar GET matrices con query param `phase_id` |
| `api/internal/seed/ninebox.go` | Modify | Seed 9 cuadrantes (no 7) con title, color_hex; entries con tiers en vez de scores; matrices con phase_id |
| `api/openapi/evaluations-and-9x9.yaml` | Modify | Actualizar schemas NineBoxEntryDTO, NineBoxMatrixResponse, NineBoxQuadrantDTO; +paths PUT quadrants, POST recompute |
| `web/src/lib/types/nine-box.ts` | Modify | `NineBoxScale` → `NineBoxTier` (1–3); `NineBoxEntry`: performance/potential → tiers; `NineBoxQuadrantDef`: colorClass → colorHex, +title; +phase en matrix |
| `web/src/lib/components/nine-box/NineBoxMatrix.svelte` | Modify | Grilla 3×3 en vez de 9×9; celdas con colorHex; dots de empleados; labels "Bajo/Medio/Alto" por tier |
| `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` | Modify | Mostrar tiers (1–3) en vez de scores (1–9); solo lectura |
| `web/src/lib/components/nine-box/NineBoxCellConfig.svelte` | Create | Modal RH: editar title, description, colorHex de cuadrante |
| `web/src/lib/components/nine-box/NineBoxSliders.svelte` | Delete | Sliders manuales ya no aplican |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Modify | Leer placements automáticos; eliminar lógica de score manual; +phase selector |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Modify | +Phase selector dropdown; labels dinámicos "Avance"/"Evaluación"; integrar NineBoxCellConfig |

## Interfaces / Contracts

### Schema changes (Ent)

```go
// nineboxmatrix.go —新增
field.UUID("phase_id", uuid.UUID{})
// Edge:
edge.From("phase", PhaseDefinition.Type).
    Ref("nine_box_matrices").Unique().Required().Field("phase_id")
// Index change:
index.Fields("cycle_id", "evaluator_id", "phase_id").Unique()
// (reemplaza el index actual de 2 campos)

// nineboxentry.go — cambios
// ELIMINAR: field.Int("performance_score").Range(1, 9)
// ELIMINAR: field.Int("potential_score").Range(1, 9)
// AGREGAR:
field.Int("performance_tier").Range(1, 3)
field.Int("potential_tier").Range(1, 3)
// quadrant(1-9) se MANTIENE

// nineboxquadrant.go — cambios
field.String("title").Optional()          // nuevo título editable
field.String("color_hex").Optional()      // ej: "#FF5733"
field.String("color").Optional()          // era NotEmpty → ahora Optional (compat)
```

### Computation functions

```go
// pkg/quadrant/compute.go — nuevas funciones

// ComputePerformanceTier maps average goal progress to tier 1–3.
//   0–33%   → 1 (low)
//   34–66%  → 2 (medium)
//   67–100% → 3 (high)
//   no data → 2 (default)
func ComputePerformanceTier(avgProgress float64) int

// ComputePotentialTier maps average competency rating (1–5) to tier 1–3.
//   1.00–2.33 → 1 (low)
//   2.34–3.66 → 2 (medium)
//   3.67–5.00 → 3 (high)
//   no data   → 2 (default)
func ComputePotentialTier(avgRating float64) int

// ComputeQuadrantFromTiers maps tier pair to quadrant 1–9.
// Same formula as existing: (potTier-1)*3 + perfTier
func ComputeQuadrantFromTiers(perfTier, potTier int) int
```

### DTO changes

```go
// NineBoxEntryDTO — response
type NineBoxEntryDTO struct {
    ID              uuid.UUID `json:"id"`
    EvaluateeID     uuid.UUID `json:"evaluateeId"`
    PerformanceTier int       `json:"performanceTier"`  // was performanceScore
    PotentialTier   int       `json:"potentialTier"`     // was potentialScore
    Quadrant        int       `json:"quadrant"`
    QuadrantLabel   string    `json:"quadrantLabel"`
    QuadrantColor   string    `json:"quadrantColor"`     // now colorHex
    Comments        string    `json:"comments,omitempty"`
    Version         int       `json:"version"`
}

// NineBoxMatrixResponse — +phaseId
type NineBoxMatrixResponse struct {
    // ...existing...
    PhaseID uuid.UUID `json:"phaseId"`
}

// NineBoxQuadrantDTO — +title, +colorHex
type NineBoxQuadrantDTO struct {
    Quadrant             int    `json:"quadrant"`
    Label                string `json:"label"`
    Title                string `json:"title"`
    Description          string `json:"description"`
    Color                string `json:"color"`      // legacy DaisyUI class
    ColorHex             string `json:"colorHex"`   // new hex color
    ActionRecommendation string `json:"actionRecommendation"`
}

// NineBoxQuadrantUpdateInput — new
type NineBoxQuadrantUpdateInput struct {
    Title       string `json:"title"`
    Description string `json:"description"`
    ColorHex    string `json:"colorHex" validate:"required,hexcolor"`
}
```

### OpenAPI endpoints (new/modified)

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| GET | `/api/v1/nine-box/matrices?cycleId=&phaseId=` | ListMatrices | Filter by phase (existing endpoint + param) |
| GET | `/api/v1/nine-box/matrices/{matrixId}/entries` | ListMatrixEntries | Returns tiers instead of scores |
| PUT | `/api/v1/nine-box/quadrants/{id}` | UpdateQuadrant | Edit title, description, colorHex |
| POST | `/api/v1/nine-box/recompute/{cycleId}/{phaseId}` | RecomputeMatrix | Auto-compute all placements |
| GET | `/api/v1/nine-box/quadrants` | GetQuadrants | Returns 9 quadrants with title, colorHex |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `ComputePerformanceTier` boundaries (0%, 33%, 34%, 66%, 67%, 100%, no-data) | Table-driven tests en `compute_test.go` |
| Unit | `ComputePotentialTier` boundaries (1.0, 2.33, 2.34, 3.66, 3.67, 5.0, no-data) | Table-driven tests en `compute_test.go` |
| Unit | `ComputeQuadrantFromTiers` all 9 combinations | Exhaustive table-driven |
| Unit | `RecomputeMatrix` service: mock repo, verify tier calculation + entry upsert | Mock NineBoxRepo + CatalogRepo |
| Integration | PUT quadrant persists title/colorHex; GET returns updated values | Test server + real DB (Ent testutil) |
| Integration | POST recompute creates entries with correct tiers from Goal + EvaluationCompetency data | Seed test data, call endpoint, verify |
| E2E | RH opens matrix, sees 3×3 grid with employees positioned correctly; edits quadrant color | Playwright: login → 9x9 page → select phase → verify grid → edit quadrant |

## Migration / Rollout

### Migration (Ent auto-migration + manual SQL)

1. **ALTER nine_box_matrices**: `ADD COLUMN phase_id UUID NOT NULL DEFAULT '<seed-phase-id>'`; add FK to phase_definitions; drop old unique index `(cycle_id, evaluator_id)`; create new unique index `(cycle_id, evaluator_id, phase_id)`.
2. **ALTER nine_box_entries**: `ADD COLUMN performance_tier INT NOT NULL DEFAULT 2`; `ADD COLUMN potential_tier INT NOT NULL DEFAULT 2`; backfill tiers from existing scores: `performance_tier = CASE WHEN performance_score <= 3 THEN 1 WHEN performance_score <= 6 THEN 2 ELSE 3 END`; same for potential. `DROP COLUMN performance_score`; `DROP COLUMN potential_score`.
3. **ALTER nine_box_quadrants**: `ADD COLUMN title TEXT`; `ADD COLUMN color_hex VARCHAR(7)`; backfill color_hex from existing labels (mapping DaisyUI class → default hex).
4. **Seed update**: Insert 2 quadrants missing (quadrant 8, 9) to complete 9; update all with title + color_hex defaults.

### Rollout

No feature flag needed. The change is backward-incompatible for the 9x9 API (scores → tiers). Communicate change to frontend team. Deploy backend migration first, then frontend in same release.

## Open Questions

- [ ] Confirmar que el seed de PhaseDefinition ya tiene registros para `avance` y `cierre` en el ciclo 2026 — si no, el seed de ninebox debe crearlos.
- [ ] Definir colores hex por defecto para los 9 cuadrantes (proposal menciona DaisyUI classes actuales; need mapping: `bg-success/20` → `#22C55E`, etc.).
- [ ] `EvaluationCompetency` tiene `profile_id` — confirmar cómo distinguir self_rating vs hr_rating. Probablemente por `Evaluation.state` o por el actor que creó el rating.
