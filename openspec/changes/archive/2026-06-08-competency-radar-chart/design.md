# Design: Radar Chart for Competency Visualization

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│  CompetencyNetworkView.svelte (MODIFIED)                            │
│  ┌─────────────────────────────────────────────────────────────────┐│
│  │  tabs-box (radio inputs + Lucide icons)                        ││
│  │  ┌──────────────┐  ┌──────────────┐                            ││
│  │  │ ○ Table      │  │ ● Network    │   ← tab activo             ││
│  │  └──────────────┘  └──────────────┘                            ││
│  ├─────────────────────────────────────────────────────────────────┤│
│  │  {#if tab === 'table'}       {#if tab === 'radar'}             ││
│  │    <ComparisonTable .../>       <RadarChart .../>              ││
│  │  {/if}                       {/if}                             ││
│  └─────────────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────────────┘
         │                                      ▲
         │ props: pillarGroups                  │
         ▼                                      │
┌────────────────────────────────────────────────┐
│  RadarChart.svelte (NEW)                       │
│                                                │
│  transforms pillarGroups → Chart.js datasets   │
│  renders <canvas> on mount via onMount         │
│  destroys Chart instance onDestroy             │
│  responsive: true, scale: 1-5                 │
│  datasets: auto (teal) + RH (amber)           │
│  legend: color indicator bar at bottom         │
└────────────────────────────────────────────────┘
```

## Architecture Decisions

| Decision | Options | Tradeoff | Choice |
|----------|---------|----------|--------|
| Chart library | D3.js vs Chart.js vs vanilla SVG | Chart.js: simpler API, radar built-in, less bundle overhead (~60 KB gzipped); D3: 3×+ bundle, more flexible but overkill | **Chart.js** — radar chart es un caso de uso directo |
| Tabs implementation | DaisyUI tabs vs native `<input type="radio">` + `<label>` | DaisyUI `tab` component usa estilos predefinidos pero no maneja `aria` nativamente; radio inputs permiten control fino de accesibilidad | **Radio inputs con label** — wrapper `tabs-box` con DaisyUI utility classes |
| Tab icons | SVG inline vs Lucide `Icon` component | `Icon` ya es dependencia del proyecto; consistente con otras vistas | **Lucide** `Table` y `Network` |
| Data flow | Prop drilling vs context vs store | El radar recibe datos ya transformados desde `CompetencyNetworkView`; no necesita store propio | **Props** — `pillarGroups` y `employeeName` |
| Chart reactivity | Full reactive rebuild vs manual update | `$effect` que detecta cambios en `pillarGroups` y llama `chart.update()` solo en datos, no en configuración | **Manual update** — `$effect` con `chart.data.datasets` replacement |

## Directory Structure

```
web/src/lib/components/evaluation/
├── RadarChart.svelte          (NEW)
├── CompetencyNetworkView.svelte (MODIFIED)
├── ComparisonTable.svelte     (unchanged)
└── ...
```

## Dependencies

### `web/package.json`

```json
{
  "dependencies": {
    "chart.js": "^4.4.8"
  }
}
```

Sin tipos separados — Chart.js 4 incluye tipos nativos.

## Components

### CompetencyNetworkView.svelte (MODIFIED)

**Additions:**

1. Import `RadarChart.svelte` y `RadarPillarGroup`
2. Estado local: `let activeTab = $state<'table' | 'radar'>('table')`
3. Computed: `pillarGroups: RadarPillarGroup[]` derivado de `pillars` + `ratings`
4. Template: tabs-box + render condicional

**Template structure:**

```svelte
<div class="flex flex-col gap-6">
  <!-- tabs-box -->
  <div class="tabs-box" role="tablist" aria-label="Selector de vista">
    <input type="radio" name="competency-view"
           id="view-table" value="table" bind:group={activeTab}
           class="hidden peer/view-table" />
    <label for="view-table" class="tab" role="tab"
           aria-controls="panel-table" aria-selected={activeTab === 'table'}
           tabindex={activeTab === 'table' ? 0 : -1}>
      <Icon name="table" class="w-4 h-4" />
      Vista actual
    </label>

    <input type="radio" name="competency-view"
           id="view-radar" value="radar" bind:group={activeTab}
           class="hidden peer/view-radar" />
    <label for="view-radar" class="tab" role="tab"
           aria-controls="panel-radar" aria-selected={activeTab === 'radar'}
           tabindex={activeTab === 'radar' ? 0 : -1}>
      <Icon name="network" class="w-4 h-4" />
      Gráfica radar
    </label>
  </div>

  <!-- Panels -->
  {#if activeTab === 'table'}
    <div id="panel-table" role="tabpanel" aria-labelledby="view-table">
      <!-- existing ComparisonTable loop -->
    </div>
  {:else}
    <div id="panel-radar" role="tabpanel" aria-labelledby="view-radar">
      <RadarChart {pillarGroups} {employeeName} />
    </div>
  {/if}
</div>
```

**Derived data — pillarGroups:**

```ts
const pillarGroups: RadarPillarGroup[] = $derived.by(() => {
  return pillars
    .map((pillar) => {
      const comps = getCompetenciesByPillar(pillar.id);
      const competencies: RadarCompetencyPoint[] = comps.map((c) => {
        const r = ratings.find((r) => r.competencyId === c.id);
        return {
          competencyId: c.id,
          competencyName: c.name,
          selfRating: r?.selfRating ?? null,
          rhRating: r?.rhRating ?? null,
        };
      });
      return { pillarId: pillar.id, pillarName: pillar.name, competencies };
    })
    .filter((g) => g.competencies.length > 0);
});
```

### RadarChart.svelte (NEW)

**Props:**

```ts
interface Props {
  pillarGroups: RadarPillarGroup[];
  employeeName?: string;
}

let { pillarGroups, employeeName = 'Empleado' }: Props = $props();
```

**Lifecycle:**

| Hook | Action |
|------|--------|
| `onMount` | Create Chart.js instance en el canvas |
| `onDestroy` | `chart.destroy()` |
| `$effect` (pillarGroups) | Update datasets + labels via `chart.data.labels = ...`, `chart.data.datasets = ...`, `chart.update('none')` |

**Chart.js config:**

```ts
const config: ChartConfiguration<'radar'> = {
  type: 'radar',
  data: {
    labels: [],        // nombres de competencia (plano, todos los pilares)
    datasets: [],      // 0, 1 o 2 datasets según datos disponibles
  },
  options: {
    responsive: true,
    maintainAspectRatio: true,
    scales: {
      r: {
        min: 1,
        max: 5,
        ticks: {
          stepSize: 1,
          callback: (v) => `${v}`,  // etiquetas numéricas 1-5
        },
        pointLabels: {
          font: { size: 11 },
          color: '#6b7280',  // text-base-content/60
        },
      },
    },
    plugins: {
      legend: { display: false },  // leyenda custom abajo
      tooltip: {
        callbacks: {
          label: (ctx) => {
            const label = ctx.dataset.label ?? '';
            return `${label}: ${ctx.parsed.r}`;
          },
        },
      },
    },
  },
};
```

**Dataset colors (DaisyUI design tokens):**

| Dataset | Fill | Border | DaisyUI token |
|---------|------|--------|---------------|
| Autoevaluación | `rgba(13, 148, 136, 0.15)` | `rgb(13, 148, 136)` | teal-600 |
| RH | `rgba(245, 158, 11, 0.15)` | `rgb(245, 158, 11)` | amber-500 |

**Data transformation — datasets:**

```ts
const allCompetencies = pillarGroups.flatMap((g) => g.competencies);
const labels = allCompetencies.map((c) => c.competencyName);

const datasets: ChartDataset<'radar'>[] = [];

const hasSelf = allCompetencies.some((c) => c.selfRating !== null);
const hasRh = allCompetencies.some((c) => c.rhRating !== null);

if (hasSelf) {
  datasets.push({
    label: 'Autoevaluación',
    data: allCompetencies.map((c) => c.selfRating),
    backgroundColor: 'rgba(13, 148, 136, 0.15)',
    borderColor: 'rgb(13, 148, 136)',
    pointBackgroundColor: 'rgb(13, 148, 136)',
    borderWidth: 2,
    pointRadius: 3,
  });
}

if (hasRh) {
  datasets.push({
    label: 'RH',
    data: allCompetencies.map((c) => c.rhRating),
    backgroundColor: 'rgba(245, 158, 11, 0.15)',
    borderColor: 'rgb(245, 158, 11)',
    pointBackgroundColor: 'rgb(245, 158, 11)',
    borderWidth: 2,
    pointRadius: 3,
  });
}
```

**Legend (HTML custom, inferior):**

```svelte
{#if datasets.length > 0}
  <div class="flex justify-center gap-6 mt-4 text-sm" aria-label="Leyenda del radar">
    {#if hasSelf}
      <span class="flex items-center gap-2">
        <span class="w-3 h-3 rounded-full bg-teal-600" />
        Autoevaluación
      </span>
    {/if}
    {#if hasRh}
      <span class="flex items-center gap-2">
        <span class="w-3 h-3 rounded-full bg-amber-500" />
        RH
      </span>
    {/if}
  </div>
{/if}
```

**Empty state:**

```svelte
{#if !hasSelf && !hasRh}
  <div class="text-center py-8">
    <p class="text-sm text-base-content/40">
      No hay datos de competencias para {employeeName}.
    </p>
  </div>
{/if}
```

**Reactivity — pillarGroups change:**

```ts
$effect(() => {
  const fresh = pillarGroups;  // track dependency
  if (!chart) return;

  const all = fresh.flatMap((g) => g.competencies);
  chart.data.labels = all.map((c) => c.competencyName);

  chart.data.datasets = [];
  if (all.some((c) => c.selfRating !== null)) {
    chart.data.datasets.push({ ...selfDatasetTemplate, data: all.map((c) => c.selfRating) });
  }
  if (all.some((c) => c.rhRating !== null)) {
    chart.data.datasets.push({ ...rhDatasetTemplate, data: all.map((c) => c.rhRating) });
  }

  chart.update('none');
});
```

## Accessibility

| Element | Attribute / Behavior |
|---------|---------------------|
| Tablist | `role="tablist"`, `aria-label="Selector de vista"` |
| Tabs | `role="tab"`, `aria-selected`, `aria-controls`, `tabindex` (0 activo, -1 inactivo) |
| Panels | `role="tabpanel"`, `aria-labelledby` |
| Keyboard | Arrow Left/Right cambia de tab (preventDefault + focus + update state) |
| Canvas | `role="img"`, `aria-label="Gráfica radar de competencias de {employeeName}"` |
| Legend | `aria-label="Leyenda del radar"` |
| Empty state | `role="status"`, `aria-live="polite"` |

**Keyboard tab navigation:**

```ts
function handleTabKeydown(e: KeyboardEvent, tab: 'table' | 'radar') {
  if (e.key === 'ArrowRight' || e.key === 'ArrowLeft') {
    e.preventDefault();
    const next = e.key === 'ArrowRight' ? 'radar' : 'table';
    activeTab = next;
    document.getElementById(`view-${next}`)?.focus();
  }
}
```

## Integration Points

| Dependency | How it's used |
|------------|---------------|
| `competencyStore` | `getPillars()`, `getCompetenciesByPillar()` — ya importadas en `CompetencyNetworkView` |
| `evaluationStore` | `getCompetencyRatings(employeeId)` — ya importada |
| `ComparisonTable` | Se mantiene sin cambios, solo se oculta condicionalmente |
| `chart.js` | Nueva dependencia; import `{ Chart, RadarController, RadialLinearScale, PointElement, LineElement, Tooltip }` |
| Lucide `Icon` | Componente `Icon` ya disponible en `$lib/components/ui/` |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| Chart.js bundle size (~60 KB gzipped) | Tree-shakeable: import solo `RadarController`, `RadialLinearScale`, `PointElement`, `LineElement`, `Tooltip` — evita registros innecesarios |
| Canvas no accesible por defecto | `role="img"` + `aria-label` descriptivo; tab nativo manejado por los radio inputs |
| Overplotting con muchas competencias | Chart.js escala automáticamente el radio del radar; si > 12 ejes el texto puede superponerse — se documenta como non-goal en spec |
| Chart instance no destruido en Svelte 5 | `onDestroy(() => chart?.destroy())` garantiza cleanup |
| `$effect` corre en servidor (SSR) | Guard `if (!browser) return;` + `onMount` para crear chart solo en cliente |
