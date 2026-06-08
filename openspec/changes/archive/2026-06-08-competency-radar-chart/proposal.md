## Why

La vista actual de competencias en `evaluacion/9x9/competencias/:id` usa una tabla HTML pura (ComparisonTable) que muestra cada competencia como fila con puntuaciones numéricas. Al cruzar múltiples pilares y comparar autoevaluación vs RH, el usuario pierde la capacidad de identificar patrones visualmente. Un radar chart permite ver de un vistazo el perfil completo de competencias, las brechas entre auto y RH, y los pilares donde hay fortalezas o debilidades.

## What Changes

- Install **chart.js** como dependencia del frontend (peer dep opcional para types)
- Crear `RadarChart.svelte` que renderiza un canvas con Chart.js usando las competencias como ejes y escala 1–5
- Agregar **tabs-box** con radio inputs e iconos Lucide (Table, Network) en la parte superior de `CompetencyNetworkView.svelte`
- Tab activo por defecto: "Vista actual" (tabla existente)
- Tab secundario: "Gráfica radar" con el nuevo componente
- Mostrar dos datasets superpuestos en el radar: autoevaluación (un color) y RH (otro color)
- Agregar **color indicator / legend** en la parte inferior del radar identificando cada dataset
- No se modifican rutas, stores, tipos existentes ni la tabla actual

## Capabilities

### New Capabilities
- `competency-radar-visualization`: Visualización tipo radar (Chart.js) para comparar perfiles de competencias entre autoevaluación y RH, con ejes por competencia, escala 1–5, y selector de vista (tabla/radar) mediante tabs-box. Aplica exclusivamente a la ruta `evaluacion/9x9/competencias/:id`.

### Modified Capabilities
<!-- Ninguna — los cambios son puramente de presentación. No se modifican requisitos de specs existentes. -->

## Impact

- **Dependencia nueva**: `chart.js` en `web/package.json`
- **Componente nuevo**: `web/src/lib/components/evaluation/RadarChart.svelte`
- **Componente modificado**: `web/src/lib/components/evaluation/CompetencyNetworkView.svelte` — se agrega el wrapper de tabs y la lógica de toggle entre tabla y radar
- **Sin impacto en**: stores, tipos (`CompetencyRating`, `Competency`, `Pillar`), rutas, API, ni otros módulos
