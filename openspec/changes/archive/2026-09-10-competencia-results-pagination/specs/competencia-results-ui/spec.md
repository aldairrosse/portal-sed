# competencia-results-ui Specification

## Purpose

Reescritura de la vista "Resultados de competencias" siguiendo el patrón de `rh/evaluaciones/+page.svelte`: paginación server-side, búsqueda, skeleton solo en carga inicial, y toggle de filtro "Mi equipo".

**File**: `web/src/routes/evaluacion/9x9/competencias/+page.svelte`

## Requirements

### Requirement: Page header with description

La página SHALL mostrar título "Resultados de competencias" con icono `BarChart3` (de lucide-svelte) y descripción: "Promedios de autoevaluación y evaluación RH por empleado en el ciclo activo".

### Requirement: Search input always enabled

Input de búsqueda (`input-bordered input-sm`, `max-w-xs`) SHALL estar **siempre habilitado** — sin `disabled={loading}`. Placeholder: "Buscar empleado...".

### Requirement: Team filter toggle

Toggle "Mi equipo" SHALL usar `input type="checkbox"` con clase `toggle toggle-sm`. Tooltip: "Excluye tu propio registro de la lista". Al activarse, SHALL llamar `setScope('team')`; al desactivarse, `setScope('all')`.

### Requirement: Result count display

A la derecha de la barra de búsqueda, SHALL mostrar:

| Condición | Texto |
|-----------|-------|
| Sin búsqueda activa, hay resultados | "Viendo {items.length} de {apiTotal} empleados" |
| Búsqueda activa, hay resultados | "Viendo {items.length} resultado(s)" |
| Sin resultados | Oculto |

#### Scenario: Search updates counter

- GIVEN 120 empleados totales, 50 por página, sin búsqueda
- WHEN página carga
- THEN barra muestra "Viendo 50 de 120 empleados"

### Requirement: Previous / Next pagination controls

Botones SHALL usar `btn-outline btn-xs` con iconos `ChevronLeft`/`ChevronRight`. Estados deshabilitados: `!hasPrev || loading` para Anterior, `!hasMore || loading` para Siguiente. Número de página SHALL mostrarse como "Pág. {currentPage + 1}". Los controles SHALL estar **siempre visibles** — fuera del condicional de loading.

### Requirement: Skeleton only on initial load

Skeleton (`PageSkeleton variant="table" rows={5}`) SHALL renderizarse SOLO cuando `{#if loading && items.length === 0}`. Durante cargas de páginas subsecuentes, SHALL mantener las filas existentes visibles y mostrar los controles de paginación en estado deshabilitado.

### Requirement: Error state with retry

Componente `ErrorState` SHALL mostrar `storeError` con callback de reintento llamando `load()`. SHALL mostrarse cuando `storeError` es truthy.

### Requirement: Empty state

Cuando `items.length === 0 && !loading`, SHALL mostrar mensaje centrado en cursiva: "Sin resultados de competencias para mostrar".

### Requirement: Competency results table with 6 columns

Tabla SHALL renderizar las siguientes 6 columnas:

| Columna | Contenido | Nota |
|---------|-----------|------|
| Empleado | `item.name` | Nombre completo |
| Perfil | `titleCase(item.profileName)` | Usando `titleCase()` de `$lib/utils/text` |
| Autoevaluación | `item.selfRatingAvg?.toFixed(1) ?? '—'` | Promedio del empleado |
| RH | `item.rhRatingAvg?.toFixed(1) ?? '—'` | Promedio de RH |
| Estado | `item.status` | Etiqueta con clase condicional según valor |
| Acción | Link a `/evaluacion/9x9/competencias/{item.id}` | Texto: "Ver detalle" |

#### Scenario: All 6 columns rendered

- GIVEN `items` contiene resultados con `status='completada'`
- WHEN tabla se renderiza
- THEN las 6 columnas son visibles con sus valores
- AND la columna Perfil aplica `titleCase()`
- AND la columna Acción contiene un link funcional al detalle del empleado

### Requirement: Status label conditional styling

La columna Estado SHALL aplicar clase CSS según el valor:

| `status` | Clase | Etiqueta |
|----------|-------|----------|
| `completada` | `badge badge-success` | Completada |
| `autoevaluacion` | `badge badge-info` | Autoevaluación |
| `pendiente` | `badge badge-warning` | Pendiente |
| `sin-datos` | `badge badge-ghost` | Sin datos |

### Requirement: titleCase for profile labels

Nombres de perfil provenientes de `items[].profileName` (ej. "gerente de tienda") SHALL renderizarse vía `titleCase()` de `$lib/utils/text` en la columna Perfil.

### Requirement: Store integration

Página SHALL importar de `competencyResultsStore.svelte.ts` usando el mismo patrón de getters que `rh/evaluaciones`: `getItems`, `isLoading`, `getError`, `hasMoreItems`, `hasPrevItems`, `getCurrentPage`, `getTotalCount`, `load`, `search`, `next`, `prev`, `setScope`. `onMount` SHALL llamar `init(cycleId)` y luego `load()`. `cycleId` SHALL obtenerse del contexto del ciclo activo (mismo mecanismo que otras vistas 9x9).
