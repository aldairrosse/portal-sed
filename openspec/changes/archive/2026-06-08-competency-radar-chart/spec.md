# competency-radar-chart — Spec

> Delta spec. Modifica `competency-framework` (visualización) y la ruta `9x9/competencias`.

## Purpose

Agregar visualización tipo radar (Chart.js) en la vista de competencias de un empleado (`/evaluacion/9x9/competencias/:id`), permitiendo alternar entre la tabla actual y una gráfica radial que compara autoevaluación vs RH por competencia agrupada en pilares.

## Requirements

### Requirement: Toggle between table and radar view

Sistema SHALL mostrar un tabs-box con dos opciones: "Vista actual" (tabla) y "Gráfica radar". El tab por defecto SHALL ser "Vista actual".

#### Scenario: Switch from table to radar

- GIVEN vista actual de competencias en modo tabla
- WHEN usuario selecciona tab "Gráfica radar"
- THEN la tabla se oculta y se renderiza el radar chart
- AND el canvas del radar es focusable con teclado

#### Scenario: Return to table from radar

- GIVEN radar chart visible
- WHEN usuario selecciona tab "Vista actual"
- THEN el radar se oculta y la tabla se muestra
- AND no hay pérdida de estado en la tabla

### Requirement: Radar chart with dual datasets

Sistema SHALL renderizar un radar chart con un dataset por evaluador (autoevaluación, RH). Cada eje SHALL representar una competencia. La escala SHALL ser 1–5.

#### Scenario: View radar with complete data

- GIVEN empleado con 6 competencias evaluadas (auto + RH)
- WHEN usuario cambia a "Gráfica radar"
- THEN el radar muestra 6 ejes con nombres de competencia
- AND dataset "Autoevaluación" con valores coloreados (azul/teal)
- AND dataset "RH" con valores coloreados (naranja/amber)
- AND cada eje escala de 1 a 5
- AND leyenda en parte inferior indicando qué color corresponde a cada dataset

#### Scenario: Radar with partial data (auto only)

- GIVEN empleado con autoevaluación completa y sin calificación RH
- WHEN usuario cambia a "Gráfica radar"
- THEN el radar muestra solo el dataset de autoevaluación
- AND el dataset RH no se renderiza (no muestra ceros)
- AND la leyenda muestra solo el indicador de autoevaluación

#### Scenario: Radar with partial data (RH only)

- GIVEN empleado sin autoevaluación pero con calificación RH
- WHEN usuario cambia a "Gráfica radar"
- THEN el radar muestra solo el dataset RH
- AND el dataset de autoevaluación no se renderiza
- AND la leyenda muestra solo el indicador RH

#### Scenario: Radar with no data

- GIVEN empleado sin autoevaluación ni calificación RH en ninguna competencia
- WHEN usuario cambia a "Gráfica radar"
- THEN el radar NO se renderiza
- AND se muestra mensaje "No hay datos de competencias para {nombre}"
- AND el tabs-box sigue visible

### Requirement: Grouping by pillar

Sistema SHALL agrupar competencias por pilar en el radar, usando un color de relleno de fondo sutil por pilar o mostrando separación visual entre pilares.

#### Scenario: Competencies from multiple pillars

- GIVEN empleado con competencias en 3 pilares distintos
- WHEN usuario cambia a "Gráfica radar"
- THEN cada pilar se renderiza en el mismo radar (todos los ejes visibles simultáneamente)
- AND visualmente se distingue qué competencias pertenecen a cada pilar

### Requirement: No data modification

El radar SHALL ser solo visualización. No SHALL modificar ratings, ni stores, ni tipos existentes.

#### Scenario: Radar is read-only

- GIVEN radar chart visible con datasets de auto y RH
- WHEN usuario interactúa con el canvas (hover, click)
- THEN no se modifican valores en la BD ni en stores
- AND tooltip de Chart.js muestra: competencia, valor auto, valor RH

## Data Contracts

### CompetencyRating (existente, no se modifica)

```ts
export interface CompetencyRating {
  id: string;
  employeeId: string;
  competencyId: string;
  selfRating?: 1 | 2 | 3 | 4 | 5;
  selfComment?: string;
  rhRating?: 1 | 2 | 3 | 4 | 5;
  rhComment?: string;
}
```

### Estructura de agrupación (nueva, solo para consumo del radar)

```ts
/** Datos transformados para el radar, agrupados por pilar */
export interface RadarPillarGroup {
  pillarId: string;
  pillarName: string;
  competencies: RadarCompetencyPoint[];
}

export interface RadarCompetencyPoint {
  competencyId: string;
  competencyName: string;
  selfRating: number | null;  // null si no hay autoevaluación
  rhRating: number | null;    // null si no hay calificación RH
}
```

### Props del componente RadarChart

```ts
interface RadarChartProps {
  pillarGroups: RadarPillarGroup[];
  employeeName?: string;
}
```

## Non-goals

- **Interactividad en el radar**: no se implementa click para editar valores, solo tooltips de solo lectura.
- **Animaciones narrativas**: Chart.js animaciones por defecto sí, pero no secuencias guiadas.
- **Exportar radar como imagen**: no se agrega botón de descarga en esta iteración.
- **Responsive complex**: el radar se adapta al contenedor vía `responsive: true` de Chart.js; no se definen breakpoints específicos.
- **Soporte para más de 12 competencias**: el radar se renderiza pero ejes múltiples pueden superponerse; no se implementa scroll o paginación de ejes.
