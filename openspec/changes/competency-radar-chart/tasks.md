# Tasks: Radar Chart for Competency Visualization

## Delivery Strategy

| Field | Value |
|-------|-------|
| Estimated changed lines | ~250 |
| 400-line budget risk | Low |
| Chained PRs | No — single PR |
| Decision needed before apply | No |

Single PR: todos los cambios son autocontenidos (1 dependencia, 1 componente nuevo, 1 modificación menor). No hay dependencias externas ni cambios en stores/types/rutas.

## Tasks

### 1. Install chart.js dependency

- [x] `pnpm add chart.js` en `web/`
- [x] Verificar que `import { Chart } from 'chart.js'` resuelve tipos correctamente
- [x] No `@types/chart.js` — Chart.js 4 incluye tipos nativos

**Acceptance:** `pnpm run check` pasa; `Chart` se importa sin errores

---

### 2. Create RadarPillarGroup and RadarCompetencyPoint types

- [x] Crear `RadarPillarGroup` y `RadarCompetencyPoint` en archivo de types del radar (ej. `web/src/lib/types/radar-chart.ts`)
- [x] Exportar ambos tipos

**Acceptance:** Tipos compilan; un test manual de creación de `RadarPillarGroup` funciona

---

### 3. Create RadarChart.svelte — component shell

- [x] Crear `web/src/lib/components/evaluation/RadarChart.svelte`
- [x] Props: `pillarGroups: RadarPillarGroup[]`, `employeeName?: string`
- [x] Template: contenedor `<div>` con `<canvas>` + empty state condicional
- [x] Canvas con `role="img"` y `aria-label` dinámico
- [x] Render condicional: si no hay datos, mostrar mensaje "No hay datos de competencias para {name}" con `role="status"` y `aria-live="polite"`
- [x] Si hay datos, mostrar canvas

**Acceptance:** Componente renderiza sin errores; empty state se muestra con `pillarGroups=[]`

---

### 4. Implement Chart.js radar instance

- [x] Importar y registrar Chart.js modules: `RadarController`, `RadialLinearScale`, `PointElement`, `LineElement`, `Tooltip`
- [x] Guard `let chart: Chart<'radar'> | undefined`
- [x] `onMount`: crear instancia Chart.js en el canvas con configuración base (scale 1–5, stepSize 1)
- [x] `onDestroy`: `chart?.destroy()`
- [x] `$effect` con guard `if (!browser) return;`: actualizar `chart.data.labels` y `chart.data.datasets` cuando cambie `pillarGroups`, llamar `chart.update('none')`

**Acceptance:** Al montar el componente con datos, el canvas muestra un radar con ejes y escala 1–5; al desmontar, no hay memory leaks (console no logged)

---

### 5. Implement dataset transformation

- [x] `allCompetencies = pillarGroups.flatMap(g => g.competencies)`
- [x] `labels = allCompetencies.map(c => c.competencyName)`
- [x] Dataset "Autoevaluación": `data = allCompetencies.map(c => c.selfRating)`, color teal-600 (`rgb(13, 148, 136)`), fill `rgba(13, 148, 136, 0.15)`
- [x] Dataset "RH": `data = allCompetencies.map(c => c.rhRating)`, color amber-500 (`rgb(245, 158, 11)`), fill `rgba(245, 158, 11, 0.15)`
- [x] No incluir dataset si todas sus values son `null` (evaluación condicional `hasSelf` / `hasRh`)

**Acceptance:** Con datos mixtos (auto sí, RH no) se muestra 1 dataset; con ambos, 2 datasets superpuestos

---

### 6. Add custom HTML legend

- [x] Agregar `<div>` con `aria-label="Leyenda del radar"` debajo del canvas
- [x] Mostrar círculo de color (DaisyUI `rounded-full`) + etiqueta por dataset presente
- [x] Autoevaluación: teal (DaisyUI `bg-teal-600`)
- [x] RH: amber (DaisyUI `bg-amber-500`)
- [x] Si solo 1 dataset presente, mostrar solo esa leyenda
- [x] `display: false` en plugins.legend de Chart.js

**Acceptance:** Leyenda se renderiza correctamente para 1 o 2 datasets; coincide con colores del radar

---

### 7. Configure chart tooltips

- [x] Tooltip muestra `{dataset.label}: {value}` en cada punto
- [x] Formato: "Autoevaluación: 4" o "RH: 3"
- [x] Tooltip se muestra en hover sobre puntos del radar
- [x] No hay tooltip en áreas vacías del canvas

**Acceptance:** Hover sobre punto de autoevaluación muestra "Autoevaluación: 4"; hover sobre RH muestra "RH: 3"

---

### 8. Modify CompetencyNetworkView — add tabs-box and radar

- [x] Importar `RadarChart.svelte`, `RadarPillarGroup` type
- [x] Estado `activeTab: 'table' | 'radar'` = `'table'` (default)
- [x] Computed `pillarGroups: RadarPillarGroup[]` derivado de `pillars` + `ratings`
- [x] Agregar `tabs-box` arriba con dos radio inputs + Lucide icons (`Table`, `Network`)
- [x] Labels: "Vista actual" (tabla) y "Gráfica radar"
- [x] Wrap tabla existente en `{#if activeTab === 'table'}`, radar en `{:else}`
- [x] Keyboard: manejar ArrowLeft/ArrowRight para cambiar tabs
- [x] Accesibilidad: `role="tablist"`, `role="tab"`, `aria-selected`, `aria-controls`, `tabindex`

**Acceptance:** Vista actual se muestra por defecto; al seleccionar "Gráfica radar" aparece el radar; teclado (ArrowLeft/ArrowRight) cambia tabs; `pnpm run check` pasa

---

### 9. Style tabs-box with DaisyUI utilities

- [x] `tabs-box` wrapper con DaisyUI `tabs` / `tabs-bordered` o utility classes equivalentes
- [x] Tab activo: `tab-active` class
- [x] Radio inputs ocultos (`hidden peer/...`), labels estilizados como tabs vía `peer-checked:` + DaisyUI
- [x] Separación visual de tabs: borde inferior en tab activo
- [x] Consistencia de espaciado con el resto de la UI (gap, padding)

**Acceptance:** Tabs visualmente consistentes con diseño del sistema; tab activo se distingue claramente

---

### 10. Handle partial data edge cases

- [x] Auto sí + RH no → 1 dataset, 1 ítem en leyenda
- [x] Auto no + RH sí → 1 dataset, 1 ítem en leyenda
- [x] Auto no + RH no → empty state ("No hay datos de competencias...")
- [x] Todas las competencias sin rating en uno de los dos orígenes → dataset omitido

**Acceptance:** Los 4 casos se renderizan correctamente sin errores en consola de Chart.js

---

### 11. Final verification

- [x] `pnpm run check` pasa sin errores
- [x] `pnpm run lint` pasa sin warnings nuevos
- [ ] Revisión manual en navegador: navegar a `/evaluacion/9x9/competencias/:id`, cambiar tabs, verificar radar se renderiza
- [ ] Revisión de accesibilidad: `<canvas>` tiene `aria-label`, tabs navegables con teclado
- [ ] Verificar que la tabla original no perdió funcionalidad (scroll, columnas, badges de brecha)

**Acceptance:** Lint + typecheck pasan; radar se renderiza con datos mock; tabla sigue funcionando igual que antes

---

## Summary

| # | Task | Type | Est. lines |
|---|------|------|:----------:|
| 1 | Install chart.js | dependency | 1 |
| 2 | Create radar-chart types | type | 15 |
| 3 | RadarChart.svelte shell | component | 25 |
| 4 | Chart.js instance | integration | 45 |
| 5 | Dataset transformation | logic | 30 |
| 6 | Custom HTML legend | template | 20 |
| 7 | Tooltip configuration | config | 10 |
| 8 | Tabs-box + conditional render | component | 55 |
| 9 | DaisyUI tabs styling | style | 20 |
| 10 | Partial data edge cases | logic | 15 |
| 11 | Final verification | qa | — |
| | **Total** | | **~236** |
