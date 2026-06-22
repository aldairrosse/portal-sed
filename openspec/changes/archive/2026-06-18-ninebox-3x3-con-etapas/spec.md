# Spec: ninebox-3x3-con-etapas

## 1. Resumen ejecutivo

Se reduce la matriz de 9×9 manual (81 celdas, scores 1–9 por eje, 18 definiciones de escala) a una matriz 3×3 con 9 cuadrantes configurados por RH y tiers 1–3 derivados automáticamente: desempeño desde avance de metas, potencial desde promedio autoevaluación + evaluación RH de competencias. La matriz se instancia por evaluador × etapa del ciclo (medio año, fin de año).

## 2. Modelo de datos

| Entidad | Cambio |
|---------|--------|
| **NineBoxMatrix** | +`phaseId` (FK a PhaseDefinition). Unique `(cycleId, evaluatorId, phaseId)`. Una matriz por evaluador por etapa. |
| **NineBoxEntry** | Eliminar `performanceScore`, `potentialScore` (1–9 manuales). +`performanceTier` (1–3), +`potentialTier` (1–3). `quadrant` (1–9) sigue calculado automático. |
| **NineBoxQuadrant** | +`title`, +`description` (texto libre RH), +`colorHex` (reemplaza `color` DaisyUI class). `quadrant` sigue 1–9 fijo. |
| **NineBoxScale** | Sin cambios. Se conserva para compatibilidad histórica. |
| **PhaseDefinition** | +Edge → `NineBoxMatrix` (nuevo). |

## 3. Reglas de cálculo

### 3.1 Eje X — Performance Tier (1–3)

```
avgProgress = AVG(GoalAssignment.progress) WHERE progress > 0 del evaluatee
Si no hay metas con avance → tier 2 (default)
Tier = 1 si avgProgress ≤ 33%
Tier = 2 si 34% ≤ avgProgress ≤ 66%
Tier = 3 si avgProgress ≥ 67%
```

### 3.2 Eje Y — Potential Tier (1–3)

```
avgPotential = AVG(selfRating + hrRating) // ambas escala 1–5
Si falta una → usar la disponible
Si no hay ninguna → tier 2 (default)
Tier = 1 si 1.0 ≤ avgPotential ≤ 2.33
Tier = 2 si 2.34 ≤ avgPotential ≤ 3.66
Tier = 3 si 3.67 ≤ avgPotential ≤ 5.0
```

### 3.3 Cuadrante

`quadrant = (potentialTier − 1) × 3 + performanceTier` — genera valor 1–9.

## 4. Comportamiento por etapa

| Etapa | Quién | Qué |
|-------|-------|-----|
| **Medio año** | RRHH | Evalúa competencias (avance). Labels UI dicen "Avance". Se calcula potential tier si hay datos de competencias. |
| **Medio año** | Sistema | Crea `NineBoxMatrix` por evaluador al cerrar etapa. Calcula performance tier desde avance de metas. |
| **Fin de año** | Empleado + RH | Autoevaluación + evaluación RH completas. |
| **Fin de año** | Sistema | Recalcula placements con datos completos de ambas evaluaciones. |
| **Ambas etapas** | RH | Puede editar title, description, colorHex de cada cuadrante. |

Empleados sin datos suficientes para alguno de los dos ejes → aparecen como "pendientes de ubicación" (sin cuadrante asignado).

## 5. API endpoints

| Método | Ruta | Descripción |
|--------|------|-------------|
| `GET` | `/api/v1/ninebox/matrix/{cycleId}/{evaluatorId}/{phase}` | Devuelve matriz con placements automáticos |
| `PUT` | `/api/v1/ninebox/quadrant/{id}` | Actualiza title, description, colorHex de un cuadrante |
| `POST` | `/api/v1/ninebox/recompute/{cycleId}/{phase}` | Recalcula placements de todos los evaluadores |

## 6. UI / Components

| Componente | Cambio |
|-----------|--------|
| `NineBoxMatrix.svelte` | Grilla 3×3 con 9 celdas coloreadas por `colorHex`. Empleados como dots posicionados por tiers. Sin sliders. |
| `NineBoxEntryCard.svelte` | Muestra nombre, performanceTier, potentialTier, cuadrante, comentarios. Solo lectura. |
| `NineBoxCellConfig.svelte` | **Nuevo**. Modal RH: editar título, descripción, color hex de celda. |
| `+page.svelte` (9×9) | Selector de etapa. Label: "Avance" (medio año) / "Evaluación" (fin de año). |

## 7. Criterios de aceptación

- [x] 9 cuadrantes seed con title, description, colorHex (no 7)
- [x] Performance tier calculado desde AVG(GoalAssignment.progress) escalado 1–3
- [x] Potential tier calculado desde AVG(selfRating, hrRating) escalado 1–3
- [ ] Empleados ubicados automáticamente al cerrar etapa; pendientes si faltan datos
  — Parcial: ubicación automática funciona (RecomputeMatrix + ComputePerformanceTier/PotentialTier).
    Pendiente: empleados sin datos suficientes no aparecen como "pendientes de ubicación";
    actualmente se les asigna tier 2 por default. Esto requiere un estado "pending" en NineBoxEntry.
- [x] Matriz por evaluador por etapa (una a medio año, otra a fin de año)
- [x] RH edita title, description, colorHex de cuadrantes vía modal
- [x] Tests unitarios para ComputePerformanceTier y ComputePotentialTier con casos borde
- [x] GET matrix incluye phase; PUT quadrant persiste configuración RH

## 8. Escenarios clave

### SC-01: Cálculo completo de ubicación
- GIVEN empleado con 3 metas (progress 60%, 80%, 40%) → avgProgress = 60% → tier 2
- AND selfRating = 4, hrRating = 4 → avgPotential = 4.0 → tier 3
- THEN quadrant = (3−1)×3 + 2 = 8
- AND empleado aparece en celda (performance=2, potential=3) con color del cuadrante 8

### SC-02: Medio año sin datos de potencial
- GIVEN ciclo en fase `avance`, empleado con metas pero sin autoevaluación ni ev. RH
- THEN potentialTier = 2 (default), performanceTier calculado normalmente
- AND label UI muestra "Avance"

### SC-03: Sin datos suficientes
- GIVEN empleado sin metas ni evaluaciones
- THEN aparece como "pendiente de ubicación"
- AND no se le asigna cuadrante

### SC-04: RH edita cuadrante
- GIVEN cuadrante 1 ("Bajo desempeño, bajo potencial") con colorHex `#EF4444`
- WHEN RH cambia colorHex a `#DC2626` y description a "Requiere plan de acción inmediato"
- THEN PUT persiste los cambios
- AND la grilla refleja el nuevo color inmediatamente

### SC-05: Matriz duplicada por etapa
- GIVEN evaluador "Carlos" con matriz en fase `avance` del ciclo 2025
- WHEN se crea matriz para fase `cierre` del mismo ciclo
- THEN existen dos matrices distintas para Carlos en 2025 (avance y cierre)
- AND cada una tiene sus propios placements calculados con los datos disponibles en esa etapa

---

## Delta: MODIFIED manager-9x9

### MODIFIED Requirements

#### Requirement: Calificación de desempeño y potencial

El sistema SHALL calcular automáticamente el performance tier (1–3) desde el avance de metas del evaluatee y el potential tier (1–3) desde el promedio de autoevaluación + evaluación RH de competencias. Los tiers son de solo lectura para el jefe. La calificación ya no es manual con escala 1–9.

(Previously: El jefe calificaba manualmente desempeño y potencial en escala 1–9 con sliders.)

#### Scenario: Tiers calculados automáticamente

- GIVEN evaluatee con avance de metas 60% y promedio competencias 4.0
- WHEN jefe abre la matriz 3×3
- THEN performance tier = 2, potential tier = 3
- AND los tiers NO son editables por el jefe

#### Scenario: Cálculo es independiente de la intervención del jefe

- GIVEN evaluatee con datos de metas y competencias completos
- WHEN se recalcula la matriz al cerrar etapa
- THEN los tiers se derivan exclusivamente de datos existentes (metas, autoevaluación, ev. RH)
- AND el jefe no puede modificar los tiers manualmente

#### Requirement: Cálculo automático de cuadrante

El sistema SHALL calcular el cuadrante automáticamente como `(potentialTier − 1) × 3 + performanceTier`. El cuadrante se recalcula al cambiar los tiers (por cambio en datos fuente).

(Previously: El cuadrante se calculaba desde scores manuales 1–9 de desempeño y potencial.)

#### Scenario: Cuadrante recalculado al actualizar datos fuente

- GIVEN evaluatee con performance tier 2, potential tier 3 → cuadrante 8
- WHEN se registra nuevo avance de metas que eleva performance tier a 3
- THEN el cuadrante se recalcula a 9

#### Scenario: Cambio en evaluación de competencias

- GIVEN evaluatee con potential tier 2
- WHEN RH completa evaluación de competencias y avgPotential sube a 4.2 → tier 3
- THEN el cuadrante se recalcula inmediatamente al recomputar la matriz

#### Requirement: Vista de matriz por evaluador (3×3)

El sistema SHALL renderizar la matriz como grilla 3×3 con 9 cuadrantes coloreados según configuración de `NineBoxQuadrant`. Cada empleado se posiciona según sus tiers calculados. Los cuadrantes son editables por RH (title, description, colorHex).

(Previously: Grilla 9×9 con scores manuales 1–9 y cuadrantes con colores DaisyUI fijos.)

#### Scenario: Grilla 3×3 con empleados automáticos

- GIVEN matriz de Carlos con 5 evaluatees en fase `cierre`
- WHEN se renderiza
- THEN 9 celdas coloreadas con colorHex configurado por RH
- AND cada evaluatee aparece como dot en su celda correspondiente
- AND clic en dot abre card de solo lectura con tiers y cuadrante

### REMOVED Requirements

#### Requirement: Sliders de desempeño y potencial

(Reason: Los tiers son calculados automáticamente. Ya no se requiere entrada manual del jefe para scores 1–9.)

#### Requirement: Definiciones de escala por eje (NineBoxScale usage)

(Reason: La escala 1–9 se conserva en BD pero ya no se usa en la matriz activa. Los tiers 1–3 tienen umbrales fijos derivados de los datos fuente, no de definiciones de catálogo.)
