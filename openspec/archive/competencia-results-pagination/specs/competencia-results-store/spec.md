# competencia-results-store Specification

## Purpose

Store reactivo Svelte 5 para la lista paginada de resultados de competencias. Sigue el patrón de `rhEvaluadosStore.svelte.ts`, consumiendo `GET /api/v1/evaluations/competency-results?cycle_id=&offset=&limit=&q=&scope=`.

**File**: `web/src/lib/stores/competencyResultsStore.svelte.ts`

## TypeScript Types

```ts
interface CompetencyResultItem {
  id: string;
  name: string;
  profileName: string;
  selfRatingAvg: number | null;
  rhRatingAvg: number | null;
  status: 'completada' | 'autoevaluacion' | 'pendiente' | 'sin-datos';
}

interface CompetencyResultsResponse {
  data: CompetencyResultItem[];
  meta: { hasMore: boolean; limit: number; offset: number; total: number };
}
```

## Requirements

### Requirement: Reactive module-level state

Store SHALL exponer `$state` fields: `items`, `loading`, `error`, `currentPage`, `apiTotal`, `hasMore`, `hasPrev`, `currentQ`, `scopeFilter`. SHALL exportar getters con el mismo naming que `rhEvaluadosStore`: `getItems`, `isLoading`, `getError`, `hasMoreItems`, `hasPrevItems`, `getCurrentPage`, `getTotalCount`.

#### Scenario: Initial state

- GIVEN store cargado pero `init()` no llamado aún
- WHEN cualquier getter es leído
- THEN `items` es `[]`, `loading` es `false`, `error` es `null`, `currentPage` es `0`, `scopeFilter` es `'all'`

### Requirement: init(cycleId) sets cycle context

`init(cycleId: string)` SHALL almacenar el ID del ciclo activo para todas las llamadas subsecuentes. SHALL llamarse una vez desde `onMount` de la página.

### Requirement: load() fetches paginated data

`load()` SHALL llamar `GET /api/v1/evaluations/competency-results?cycle_id={cycleId}&offset={currentPage * PAGE_SIZE}&limit=50&q={currentQ}&scope={scopeFilter}`. En éxito SHALL setear `items`, `hasMore`, `apiTotal`, `hasPrev`. En fallo SHALL setear `error` y limpiar `items`. `loading` SHALL ser `true` durante el fetch.

#### Scenario: Successful page load

- GIVEN `currentPage=1`, `currentQ=undefined`, `scopeFilter='all'`
- WHEN `load()` completa exitosamente
- THEN `items` contiene la página 2 de empleados, `apiTotal` refleja el total del servidor, `hasPrev` es `true`

### Requirement: next() / prev() paginate

`next()` SHALL incrementar `currentPage` y llamar `load()` cuando `hasMore` es `true`. `prev()` SHALL decrementar y llamar `load()` cuando `currentPage > 0`. Ambas SHALL ser no-ops cuando la condición de guarda falla.

### Requirement: search(query) with 300ms debounce

`search(query)` SHALL aplicar debounce de 300ms, resetear `currentPage` a `0`, setear `currentQ`, y llamar `load()`. Pulsaciones rápidas SHALL cancelar el timer previo y re-armar.

#### Scenario: Debounce coalesces keystrokes

- GIVEN usuario tipea "M", "a", "r" con 50ms de separación
- WHEN 300ms pasan desde la última pulsación
- THEN solo UNA llamada API se ejecuta con `q=Mar`

### Requirement: setScope(scope) changes team filter

`setScope(scope: 'all' | 'team')` SHALL setear `scopeFilter`, resetear `currentPage` a `0`, y llamar `load()`.

### Requirement: PAGE_SIZE constant

`PAGE_SIZE` SHALL ser `50`. El offset de página SHALL computarse como `currentPage * PAGE_SIZE`.
