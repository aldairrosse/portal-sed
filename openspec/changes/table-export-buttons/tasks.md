# Tasks: Table Export Buttons

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~80 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No

## Phase 1: Utilidad `toCsv`

- [x] 1.1 Crear `web/src/lib/utils/export.ts` con función `toCsv(rows, filename)` que:
  - Toma `Record<string, string | number | null>[]` y genera CSV con BOM
  - Usa separador `;` y escapa comillas, punto y coma, saltos de línea
  - Descarga vía Blob + `<a>` click programático
  - Early return con `console.warn` si el array está vacío

**Acceptance:** `toCsv([{A: 1, B: "hola"}], "test")` descarga `test.csv` con `\uFEFFA;B\r\n1;hola\r\n`. `toCsv([], "empty")` no descarga nada y emite warn.

## Phase 2: Botón Exportar CSV en EmployeeEvaluationTable

- [x] 2.1 Importar `toCsv` en `EmployeeEvaluationTable.svelte`
- [x] 2.2 Agregar función `handleExportCsv` que mapea `filteredEmployees` a filas planas con columnas: Empleado, Perfil, Progreso global %, Estado
- [x] 2.3 Agregar botón "Exportar CSV" (`btn btn-outline btn-sm`) después del search input, dentro del bloque `{#if !selectedEmployeeId}`
- [x] 2.4 Botón deshabilitado cuando `filteredEmployees.length === 0`

**Acceptance:** Botón visible en Mis evaluados y Evaluaciones RH. Exporta solo empleados visibles (respeta filtro). Botón disabled cuando no hay resultados.

## Phase 3: Botón Exportar CSV en Asignación anual

- [x] 3.1 Importar `toCsv` en `objetivos/asignacion/+page.svelte`
- [x] 3.2 Agregar función `handleExportCsv` que aplana categorías → metas → KPIs en filas con columnas: Categoría, Peso categoría %, Meta, Descripción, Unidad, Peso meta %, Valor objetivo, KPIs
- [x] 3.3 Agregar botón "Exportar CSV" en el header actions `<div class="flex items-center gap-2 mt-3">`
- [x] 3.4 Botón deshabilitado cuando `categories.length === 0`

**Acceptance:** Botón visible en Asignación anual. CSV incluye una fila por meta con datos de categoría padre. KPIs concatenados en una columna. Botón disabled sin categorías.

## Phase 4: Verificación

- [x] 4.1 `pnpm run lint` sin errores nuevos (34 pre-existing, ninguno de este cambio)
- [x] 4.2 `svelte-check` sin errores nuevos (2 pre-existing, ninguno de este cambio)
- [ ] 4.3 Manual: exportar desde EmployeeEvaluationTable con y sin filtro
- [ ] 4.4 Manual: exportar desde Asignación anual con múltiples categorías y metas
- [ ] 4.5 Manual: abrir CSV en Excel — encoding, separadores, caracteres especiales correctos
- [ ] 4.6 Manual: exportar tabla vacía (sin resultados de búsqueda, sin categorías) — botón disabled

## Acceptance Criteria

- [x] `toCsv()` genera CSV válido con BOM, headers, `;` separator, y escaping correcto
- [x] Botón "Exportar CSV" presente en EmployeeEvaluationTable (Mis evaluados y RH evaluaciones)
- [x] Botón "Exportar CSV" presente en Asignación anual
- [x] CSV de evaluaciones incluye columnas: Empleado, Perfil, Progreso global %, Estado
- [x] CSV de asignación anual incluye columnas planas: Categoría, Peso categoría %, Meta, Descripción, Unidad, Peso meta %, Valor objetivo, KPIs
- [x] Exportación respeta filtro de búsqueda activo en EmployeeEvaluationTable (`filteredEmployees`)
- [x] Botón deshabilitado cuando no hay datos que exportar
- [x] `pnpm run lint` sin errores nuevos (34 pre-existing, ninguno de este cambio)
- [x] `svelte-check` sin errores nuevos (2 pre-existing, ninguno de este cambio)
