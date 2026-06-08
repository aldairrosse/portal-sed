# Proposal: table-export-buttons

## Why

Los jefes y RH necesitan descargar la información de evaluaciones y metas para analizarla fuera del sistema (reportes locales, filtros avanzados, respaldos). Hoy no existe ningún mecanismo de exportación en la plataforma; la única forma de extraer datos es copiar manualmente desde la pantalla.

## What Changes

- Crear `web/src/lib/utils/export.ts` con utilidad `toCsv` que toma un array de registros y genera un CSV con BOM (compatible con Excel/Sheets).
- Agregar botón **"Exportar CSV"** en la cabecera del componente `EmployeeEvaluationTable.svelte` (cubre las pantallas "Mis evaluados" `/mis-evaluados` y "Evaluaciones RH" `/rh/evaluaciones`).
- Agregar botón **"Exportar CSV"** en la cabecera de la página de asignación anual `/objetivos/asignacion`.
- Para la exportación de asignación anual, aplanar datos de categorías y metas en filas planas (una fila por meta con nombre de categoría, peso, KPI vinculados).
- No se agregan nuevas dependencias npm; cero cambios en stores existentes.

## Capabilities

### New Capabilities

- `table-export`: Utilidad compartida de exportación CSV y botones de descarga en tablas de evaluaciones y metas.

### Modified Capabilities

- Ninguna — el comportamiento existente de las specs no cambia, solo se agrega UI de exportación.

## Approach

1. **Utilidad `export.ts`**: función `toCsv(rows: Record<string, string | number>[], filename: string): void` que:
   - Convierte cada objeto a línea CSV escapando comillas y separando por `;` (locale español).
   - Prefija el contenido con BOM (`\uFEFF`) para que Excel detecte UTF-8.
   - Crea un Blob y dispara un `<a>` click programático para descargar.

2. **EmployeeEvaluationTable**: agregar botón "Exportar CSV" en el `<div class="flex flex-col gap-6">` después del search input (solo cuando `selectedEmployeeId` está vacío). El CSV incluye columnas: `Empleado, Perfil, Progreso global %, Estado`.

3. **Asignación anual (`+page.svelte`)**: agregar botón "Exportar CSV" en el header action area (junto al botón "Guardar asignación"/"Solicitar cambio"). El CSV aplanado incluye columnas: `Categoría, Peso categoría %, Meta, Descripción, Unidad, Peso meta %, Valor objetivo, KPIs`.

4. **Columnas visibles filtradas**: exportar solo los datos de la tabla visible actual (incluyendo filtro de búsqueda en EmployeeEvaluationTable y empleado seleccionado en asignación anual).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `web/src/lib/utils/export.ts` | New | Función compartida `toCsv()` |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modified | Botón "Exportar CSV" en cabecera |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modified | Botón "Exportar CSV" en header actions |

## Dependencies

- Ninguna. Sin nuevas dependencias npm. Sin cambios en stores o fixtures.

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| CSV con `;` no abra correctamente en Excel de ciertos locales | Low | Usar BOM + `;` (locale español); probar en Excel ES, Google Sheets |
| Datos con comillas o saltos de línea rompan el CSV | Low | `toCsv` escapa `"` a `""` y wrap en quotes todo campo con caracteres especiales |
| Rendimiento: miles de metas en asignación anual | Low | Los datos ya están en memoria; generar CSV es síncrono y O(n) |

## Success Criteria

- [ ] `toCsv()` genera CSV válido con BOM, cabeceras, y escapes correctos
- [ ] Botón "Exportar CSV" aparece en EmployeeEvaluationTable (Mis evaluados)
- [ ] Botón "Exportar CSV" aparece en Evaluaciones RH
- [ ] Botón "Exportar CSV" aparece en Asignación anual
- [ ] CSV de evaluaciones incluye columnas: Empleado, Perfil, Progreso, Estado
- [ ] CSV de asignación anual incluye columnas planas: Categoría, Peso cat %, Meta, Descripción, Unidad, Peso meta %, Valor objetivo, KPIs
- [ ] Exportación respeta filtro de búsqueda activo en EmployeeEvaluationTable
- [ ] Exportación en asignación anual usa el empleado seleccionado actualmente
- [ ] `pnpm run lint` sin errores
- [ ] `tsc` sin errores
