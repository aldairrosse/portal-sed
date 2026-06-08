# table-export Specification

## Purpose

Agregar exportación CSV a las tres tablas principales del sistema: tabla de evaluados (`EmployeeEvaluationTable`), tabla RH de evaluaciones y tabla de asignación anual de metas. La exportación permite a jefes y RH descargar datos para análisis offline sin depender de stores ni nuevas dependencias.

## Requirements

### Requirement: Utilidad compartida `toCsv`

El sistema SHALL proveer una función `toCsv(rows, filename)` en `web/src/lib/utils/export.ts` que:

| Aspecto | Especificación |
|---------|---------------|
| Input | `Record<string, string \| number \| null>[]` — array de objetos planos |
| Output | Archivo `.csv` descargable vía Blob + `<a>` click programático |
| Encoding | UTF-8 con BOM (`\uFEFF`) para que Excel detecte UTF-8 en todos los locales |
| Separador | `;` (punto y coma) — locale español evita conflicto con comas decimales |
| Headers | Primera fila = keys del primer objeto |
| Escaping | Campos con `;`, `"`, `\n` o `\r` envueltos en `"..."` y `"` escapado a `""` |
| Filename | Parámetro `filename` sin extensión — se agrega `.csv` automáticamente |

#### Scenario: Exportación con datos

- GIVEN `rows = [{Nombre: "María", Edad: 30}, {Nombre: "Juan", Edad: 25}]`
- WHEN se llama `toCsv(rows, "reporte")`
- THEN se descarga un archivo `reporte.csv`
- AND el contenido es `\uFEFFNombre;Edad\r\nMaría;30\r\nJuan;25\r\n`

#### Scenario: Campo con caracteres especiales

- GIVEN `rows = [{Nombre: `María "Mía" López`, Notas: `Vendedor; senior`}]`
- WHEN se llama `toCsv(rows, "test")`
- THEN el campo Nombre se envuelve en quotes y se escapa: `"María ""Mía"" López"`
- AND el campo Notas se envuelve en quotes: `"Vendedor; senior"`

#### Scenario: Array vacío

- GIVEN `rows = []`
- WHEN se llama `toCsv([], "vacio")`
- THEN SHALL no descargar archivo (early return)
- AND SHALL hacer `console.warn("toCsv: no rows to export")`

### Requirement: Botón "Exportar CSV" en EmployeeEvaluationTable

El componente `EmployeeEvaluationTable.svelte` SHALL mostrar un botón "Exportar CSV" en el área de cabecera (junto al search input) cuando `selectedEmployeeId` está vacío (vista de lista). El CSV SHALL contener solo los empleados filtrados por `searchQuery`.

#### Scenario: Exportación con filtro activo

- GIVEN tabla con 10 empleados, filtro "María" activo que muestra 2 resultados
- WHEN usuario hace clic en "Exportar CSV"
- THEN CSV contiene solo las 2 filas filtradas
- AND columnas: `Empleado`, `Perfil`, `Progreso global %`, `Estado`

#### Scenario: Exportación sin datos filtrados

- GIVEN tabla con 0 empleados visibles (filtro sin resultados)
- WHEN usuario hace clic en "Exportar CSV"
- THEN no se descarga archivo (no hay filas que exportar)

### Requirement: Botón "Exportar CSV" en Asignación anual

La página `objetivos/asignacion/+page.svelte` SHALL mostrar un botón "Exportar CSV" en el header action area. El CSV SHALL aplanar categorías y metas en filas planas para el empleado actualmente seleccionado.

#### Scenario: Exportación con categorías y metas

- GIVEN empleado con categoría "Ventas" (peso 50%) con meta "Cierre mensual" (peso 30%, target $100k) vinculada a KPI "Tasa cierre"
- WHEN usuario hace clic en "Exportar CSV"
- THEN CSV incluye fila: `Ventas, 50%, Cierre mensual, ..., $100k, Tasa cierre`
- AND se respeta el `selectedEmployeeId` actual para filtrar datos de ese empleado

## Data Contracts

### CSV: EmployeeEvaluationTable

```
Columna            | Tipo    | Fuente
-------------------|---------|-----------------------
Empleado           | string  | emp.employeeName
Perfil             | string  | getProfileLabel()
Progreso global %  | number  | progressMap.get() o null
Estado             | string  | getStatus() — label textual
```

Formato: `Empleado;Perfil;Progreso global %;Estado`

El valor de Estado SHALL ser el texto legible del badge (`pending`, `in-progress`, `completed` o su label `EvaluationStatus`). Progreso SHALL ser `null` formateado como string vacío.

### CSV: Asignación anual (aplanado)

```
Columna            | Tipo    | Fuente
-------------------|---------|-----------------------
Categoría          | string  | category.name
Peso categoría %   | number  | category.weight
Meta               | string  | goal.name
Descripción        | string  | goal.description
Unidad             | string  | goal.unit
Peso meta %        | number  | goal.weight
Valor objetivo     | number  | goal.targetValue
KPIs               | string  | KPI names concatenados con ", "
```

Formato: `Categoría;Peso categoría %;Meta;Descripción;Unidad;Peso meta %;Valor objetivo;KPIs`

Si una meta no tiene KPIs vinculados, la columna KPIs SHALL mostrarse vacía.

## Non-goals

- Exportación a PDF, Excel (.xlsx), o cualquier formato que no sea CSV
- Botones de exportación en tabla de RH específica (usa el mismo `EmployeeEvaluationTable`, cubierto)
- Selección de columnas a exportar
- Exportación batch de múltiples empleados en asignación anual
- Ordenamiento personalizado del CSV (respeta el orden actual de la tabla)
- Dependencias npm nuevas (cero nuevas)
- Cambios en stores, fixtures, o tipos existentes
