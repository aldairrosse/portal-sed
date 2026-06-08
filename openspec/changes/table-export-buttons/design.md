# Design: Table Export Buttons

## Technical Approach

Tres cambios localizados: (1) utilidad compartida `export.ts`, (2) botón en `EmployeeEvaluationTable.svelte`, (3) botón en asignación anual. Sin nuevas dependencias, sin cambios en stores, sin nuevas rutas. La utilidad `toCsv` vive en `web/src/lib/utils/` y se importa donde se necesite.

## Architecture Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Ubicación utilidad | `web/src/lib/utils/export.ts` | Carpeta estándar para utilidades compartidas; co-located con el feature que la usa |
| Formato archivo | CSV con BOM + `;` | UTF-8 con BOM garantiza que Excel abra correctamente en locale español; `;` evita conflicto con comas decimales |
| Trigger descarga | Blob + `<a>` click programático | Sin dependencias; funciona en todos los browsers modernos; patrón estándar |
| Filtrar datos visibles | Usar `filteredEmployees` / `selectedEmployeeId` existente | La tabla ya computa los datos filtrados; exportar solo lo visible cumple expectativa del usuario |
| Plano vs anidado en metas | Aplanar: una fila por meta | El spec requiere tabla plana compatible con análisis downstream |

## Data Flow

```
EmployeeEvaluationTable                   Asignación anual page
  filteredEmployees ($derived)              categories, goals, getKpisForGoal
        │                                           │
        ▼                                           ▼
  ┌─────────────────┐                    ┌──────────────────────────────┐
  │ toCsv()         │                    │ flattenGoalsToRows()         │
  │                 │                    │  (inline, mapea cat→goal→kpi)│
  │ Blob → <a> click│                    │                              │
  └─────────────────┘                    └──────┬───────────────────────┘
        │                                           │
        └──────────────────────┬────────────────────┘
                               ▼
                    ┌─────────────────────┐
                    │ Descarga .csv       │
                    │ UTF-8 + BOM         │
                    │ Separador ;         │
                    └─────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `web/src/lib/utils/export.ts` | **New** | Función compartida `toCsv()` |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modify | Botón "Exportar CSV" en cabecera, llama a toCsv con `filteredEmployees` |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modify | Botón "Exportar CSV" en header actions, aplana datos y llama a toCsv |

## Implementation Details

### 1. `web/src/lib/utils/export.ts`

```ts
export function toCsv(
  rows: Record<string, string | number | null>[],
  filename: string
): void {
  if (rows.length === 0) {
    console.warn('toCsv: no rows to export');
    return;
  }

  const headers = Object.keys(rows[0]);

  const escape = (val: string): string =>
    val.includes(';') || val.includes('"') || val.includes('\n') || val.includes('\r')
      ? `"${val.replace(/"/g, '""')}"`
      : val;

  const lines = rows.map((row) =>
    headers
      .map((h) => {
        const v = row[h];
        return escape(v == null ? '' : String(v));
      })
      .join(';')
  );

  const bom = '\uFEFF';
  const csv = bom + headers.join(';') + '\r\n' + lines.join('\r\n') + '\r\n';
  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `${filename}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}
```

### 2. EmployeeEvaluationTable — botón + handler

Agregar en el `<div class="flex flex-col gap-6">`, dentro del bloque `{#if !selectedEmployeeId}`, después del search input:

```svelte
<button
  type="button"
  class="btn btn-outline btn-sm"
  onclick={handleExportCsv}
  disabled={filteredEmployees.length === 0}
>
  Exportar CSV
</button>
```

Y en `<script>`:

```ts
function handleExportCsv() {
  const rows = filteredEmployees.map((emp) => ({
    Empleado: emp.employeeName,
    Perfil: getProfileLabel(emp.employeeId),
    'Progreso global %': progressMap.get(emp.employeeId),
    Estado: getStatusLabel(getStatus(emp.employeeId)),
  }));
  toCsv(rows, `evaluados-${new Date().toISOString().slice(0, 10)}`);
}

function getStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    pending: 'Pendiente',
    'in-progress': 'En progreso',
    completed: 'Completado',
  };
  return labels[status] ?? status;
}
```

### 3. Asignación anual — botón + flatten

En el header actions `<div class="flex items-center gap-2 mt-3">`, antes del grupo de botones condicionales:

```svelte
<button
  type="button"
  class="btn btn-outline btn-sm"
  onclick={handleExportCsv}
  disabled={categories.length === 0}
>
  Exportar CSV
</button>
```

Y en `<script>`:

```ts
import { toCsv } from '$lib/utils/export';

function handleExportCsv() {
  const rows: Record<string, string | number | null>[] = [];
  for (const cat of categories) {
    const catGoals = getGoalsByCategory(cat.id);
    for (const goal of catGoals) {
      const kpis = getKpisForGoal(goal.id).map((k) => k.name).join(', ');
      rows.push({
        Categoría: cat.name,
        'Peso categoría %': cat.weight,
        Meta: goal.name,
        Descripción: goal.description,
        Unidad: goal.unit,
        'Peso meta %': goal.weight,
        'Valor objetivo': goal.targetValue,
        KPIs: kpis || '',
      });
    }
  }
  toCsv(rows, `asignacion-${targetEmployeeName}-${new Date().toISOString().slice(0, 10)}`);
}
```

### 4. CSV Format per Table

**Evaluaciones:**
```csv
Empleado;Perfil;Progreso global %;Estado
"María García";Jefe;72;En progreso
"Juan Pérez";Colaborador;;Pendiente
```

**Asignación anual:**
```csv
Categoría;Peso categoría %;Meta;Descripción;Unidad;Peso meta %;Valor objetivo;KPIs
Ventas;50;Cierre mensual;Cierre de caja mensual;numero;30;100000;Tasa cierre
Ventas;50;Rotacion;Rotacion de personal;porcentaje;20;15;
```

### 5. Reactivity Notes

- `filteredEmployees` en `EmployeeEvaluationTable` ya es `$derived` — el CSV siempre exporta los datos visibles actuales.
- En asignación anual, `categories`, `goals`, y `getKpisForGoal` leen `$state` — el botón siempre exporta el estado actual sin necesidad de `$derived` adicional.
- `toCsv` es una función pura sin estado — no afecta reactividad.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit | `toCsv` genera CSV válido | Manual: probar BOM, headers, escaping, archivos descargables |
| Visual | Botón visible en tabla y asignación | Manual: verificar presencia y estado disabled |
| Visual | Botón disabled cuando no hay datos | Manual: filtro sin resultados, categorías vacías |
| Edge case | Caracteres especiales (comillas, saltos de línea) | Manual: ver CSV se abre correctamente en Excel |
| Edge case | Exportación sin metas/KPIs | Manual: KPIs vacíos en CSV |
| Integration | Empleado filtrado se refleja en CSV | Manual: aplicar filtro, exportar, verificar filas |

## Open Questions

- ¿El botón debería estar siempre visible o solo en ciertos ciclos (ej. solo en `fin-anio`)? Decisión inicial: visible siempre que haya datos. Puede ajustarse por feedback.
