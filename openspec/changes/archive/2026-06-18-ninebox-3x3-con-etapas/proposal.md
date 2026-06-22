# Proposal: ninebox-3x3-con-etapas

## Intent

Reducir la matriz de 9×9 (manual, 9 niveles por eje) a una matriz 3×3 donde los ejes se calculan automáticamente: desempeño desde avance de metas, potencial desde promedio de autoevaluación + evaluación RH. La matriz se instancia por etapa (medio año, fin de año) y sus 9 celdas son editables en título, descripción y color.

## Motivation

1. **Complejidad innecesaria**: 9 niveles manuales por eje (18 registros de escala) generan fricción sin valor. RRHH debe mantener definiciones para 18 niveles que el usuario apenas distingue.
2. **Alinear con etapas reales**: la matriz actual es una sola por ciclo anual. No refleja que medio año y fin de año son evaluaciones distintas con scopes diferentes.
3. **Automatizar ejes**: el desempeño y potencial son derivables de datos existentes (metas, autoeval, ev. RH). Calificarlos manualmente es redundante y subjetivo.
4. **Celdas editables**: los 9 cuadrantes deben reflejar la cultura de la organización, no valores hardcodeados.

## Scope

### In Scope
- Matriz 3×3 con tiers 1–3 por eje (9 cuadrantes físicos, no 81)
- Una `NineBoxMatrix` por evaluador **por etapa** (`PhaseDefinition`) — se crea al cerrar etapa
- Eje X (Desempeño): calculado desde `goal` progress escalado a tier 1–3
- Eje Y (Potencial): calculado desde promedio de `self_rating` (autoevaluación) + `hr_rating` (evaluación RH), escalado a tier 1–3
- `NineBoxQuadrant` editable: título, descripción, color hex por celda (1–9)
- Ubicación automática de empleados con datos suficientes en la matriz de su etapa
- Empleados sin datos suficientes aparecen como "pendientes de ubicación"

### Out of Scope
- Modificar la evaluación formal de competencias (A5/C6) — se lee solo
- Eliminar `NineBoxScale` de la base de datos — se mantiene por compatibilidad con datos históricos (seed existente)
- Vista de comparaciones entre etapas o ciclos históricos
- Cambios en el modelo de `Cycle` o `PhaseDefinition` — se agregan edges nada más
- Cálculo de potencial para medio año (fase `avance`) — solo fin de año tiene autoevaluación y ev. RH

## Capabilities

### New Capabilities
- `ninebox-3x3`: Matriz 3×3 con tiers derivados; ubicación automática de empleados; celdas editables
- `ninebox-per-stage`: Una instancia de matriz por evaluador por etapa; se calcula/cierra al completar la etapa

### Modified Capabilities
- `manager-9x9`: Cambia de 9×9 manual con scores 1–9 a 3×3 automático con tiers 1–3 derivados de datos existentes. Se modifica el requirement de "Calificación de desempeño y potencial" — ya no es manual.

## Approach

### Schema (Ent)

| File | Change |
|------|--------|
| `nineboxmatrix.go` | Agregar `phase` field (fk a PhaseDefinition). Unique en (cycle, evaluator, phase). |
| `nineboxentry.go` | Eliminar `performance_score` y `potential_score`. Agregar `performance_tier` y `potential_tier` (1–3). `quadrant` sigue calculado de tiers. |
| `nineboxquadrant.go` | Expandir: `title`, `description`, `color_hex` (reemplaza `color` DaisyUI class). `quadrant` pasa de 1–9 fijo a 1–9 configurable. |
| `nineboxscale.go` | Sin cambios — mantener seed histórico. |

### Computation (Backend)

**Performance tier** = escalado del avance global de metas del empleado en la etapa:
- Recorrer `GoalAssignment` con `progress` > 0 del evaluatee
- Promedio de `progress` → mapear a tier: 0–33% → 1, 34–66% → 2, 67–100% → 3
- Si no hay metas con avance → tier 2 (valor por defecto)

**Potential tier** = promedio de autoevaluación y evaluación RH, escalado:
- `self_rating` (1–5) + `hr_rating` (1–5) → promedio → mapear a tier: 1.0–2.33 → 1, 2.34–3.66 → 2, 3.67–5.0 → 3
- Si falta una de las dos → usar la disponible
- Si no hay ninguna → tier 2 (valor por defecto)

**Quadrant** = `(potential_tier-1)*3 + performance_tier` (sin cambios en la fórmula, solo los inputs son derivados)

### Service Layer

- `ninebox_service.go`: modificar `ComputePlacement` para invocar los cálculos de performance y potential. Nuevos métodos `GetOrCreateMatrixByPhase`, `RecomputeAllPlacements`.
- Agregar edge `PhaseDefinition` → `NineBoxMatrix` en el schema.

### Handler / Routes

- `routes.go`: nuevo endpoint `GET /api/v1/ninebox/matrix/:cycleId/:evaluatorId/:phase` — devuelve matriz con placements automáticos para la etapa.
- `PUT /api/v1/ninebox/quadrant/:id` — actualizar título, descripción, color de celda.

### Frontend

| File | Change |
|------|--------|
| `nine-box.ts` | Tipos: `PerformanceTier`, `PotentialTier` (1–3), `NineBoxMatrix` con `phase`, `NineBoxQuadrant` con `title`, `description`, `colorHex` |
| `nineBoxStore.svelte.ts` | Leer placements del API; ya no hay sliders de score manual |
| `NineBoxMatrix.svelte` | Grilla 3×3 con 9 celdas; cada celda muestra empleados como dots; clic abre `NineBoxEntryCard` |
| `NineBoxEntryCard.svelte` | Muestra nombre, tier desempeño, tier potencial, cuadrante; ya no hay sliders editables |
| `NineBoxCellConfig.svelte` | Nuevo: modal para que RH edite título, descripción, color hex de cada celda |

### Seed

- Actualizar `seed/ninebox.go`: crear 9 `NineBoxQuadrant` con valores por defecto (labels Genéricos: "Alto desempeño/Alto potencial", etc.) y colores hex por defecto.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/schema/nineboxmatrix.go` | Modified | +phase field, unique (cycle,evaluator,phase) |
| `api/internal/schema/nineboxentry.go` | Modified | -performance_score, -potential_score; +performance_tier, potential_tier (1–3) |
| `api/internal/schema/nineboxquadrant.go` | Modified | +title, +description, +color_hex; remove DaisyUI class coupling |
| `api/internal/schema/cycle.go` | Modified | Add edge to PhaseDefinition |
| `api/internal/pkg/quadrant/compute.go` | Modified | ComputeQuadrant stays same; add ComputePerformanceTier, ComputePotentialTier |
| `api/internal/service/evaluation/ninebox_service.go` | Modified | Auto-compute tiers on placement; per-phase matrix access |
| `api/internal/repository/evaluation/ninebox_repo.go` | Modified | Queries by phase; auto-placement computation |
| `api/internal/handler/evaluation/routes.go` | Modified | +GET matrix by phase, +PUT quadrant config |
| `web/src/lib/types/nine-box.ts` | Modified | New tier types, phase-aware matrix, editable quadrant |
| `web/src/lib/stores/nineBoxStore.svelte.ts` | Modified | Read auto-placements; no manual score editing |
| `web/src/lib/components/nine-box/NineBoxMatrix.svelte` | Modified | 3×3 grid (not 9×9) |
| `web/src/lib/components/nine-box/NineBoxEntryCard.svelte` | Modified | Show tiers, no sliders |
| `web/src/lib/components/nine-box/NineBoxCellConfig.svelte` | New | RH config modal for cell title/desc/color |
| `web/src/routes/evaluacion/9x9/+page.svelte` | Modified | Phase selector; auto-placement view |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Auto-cálculo de tiers produce valores inesperidos | Medium | Validar con datos reales de test; agregar logging del cálculo |
| Medio año sin autoevaluación — potencial no disponible | High | Potential tier = default 2 hasta que existan datos de cierre |
| RRHH edita color hex a contraste bajo | Low | Validar luminosidad mínima del hex o usar preview |
| Frontend 3×3 rompe expectativas de usuarios accustomed a 9×9 | Medium | Comunicar el cambio; mostrar tier labels en hover de celda |
| NineBoxScale sigue existiendo con datos huérfanos | Low | Mantener seed; no se elimina tabla ni datos |

## Rollback Plan

1. Revertir `nineboxentry.go`: restaurar `performance_score`, `potential_score` como campos
2. Revertir `nineboxmatrix.go`: remover `phase` field
3. Revertir `nineboxquadrant.go`: restaurar `color` (DaisyUI class) y remover `title`, `description`, `color_hex`
4. Revertir service y repo para usar scores manuales
5. Revertir frontend: restaurar sliders y grilla 9×9
6. Regenerar migrations Ent

## Dependencies

- `evaluation-lifecycle` (etapas: medio año `avance`, fin de año `cierre`)
- `competency-framework` (autoevaluación y ev. RH — se leen solo)
- `goals-and-weighting` (progress de metas para cálculo de desempeño)

## Success Criteria

- [ ] Matriz 3×3 visible con 9 cuadrantes editables por RH
- [ ] Empleados con avance de metas ≥1 aparecen ubicados automáticamente al cerrar etapa
- [ ] Performance tier refleja el promedio de avance de metas (escala 1–3)
- [ ] Potential tier refleja el promedio de self_rating + hr_rating (escala 1–3)
- [ ] RRHH puede editar título, descripción y color hex de cada celda
- [ ] Una matriz por etapa (medio año + fin de año) por evaluador
- [ ] Tests de ComputePerformanceTier y ComputePotentialTier con casos borde (sin datos, datos parciales)