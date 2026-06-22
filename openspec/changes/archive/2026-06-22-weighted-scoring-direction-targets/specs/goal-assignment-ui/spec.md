# goal-assignment-ui Specification (DELTA)

## MODIFIED Requirements

### Requirement: Edición de metas dentro de una categoría

El sistema SHALL permitir crear, editar y eliminar metas dentro de cada categoría. Cada meta SHALL tener: `name`, `description`, `unit` (`porcentaje` | `moneda` | `numero`), `weight` (0..100), `targetValue` (`> 0`) y 0..N KPIs vinculados. **ADDED**: Cada meta SHALL tener un campo `direction` (`ascendente` | `descendente`) con radio input en el formulario, y un campo condicional `baselineValue` que solo se muestra cuando `direction === 'descendente'`.

#### Scenario: Crear meta ascendente

- GIVEN categoría con peso asignado
- WHEN hace clic en "Nueva meta", selecciona direction `ascendente` (default), completa campos y confirma
- THEN la meta aparece en la tabla con badge de dirección ascendente (↑)
- AND el `baselineValue` NO se muestra ni se valida

#### Scenario: Crear meta descendente

- GIVEN categoría con peso asignado
- WHEN hace clic en "Nueva meta", selecciona direction `descendente`, completa campos incluyendo `baselineValue` y confirma
- THEN la meta aparece en la tabla con badge de dirección descendente (↓)
- AND el `baselineValue` se muestra como valor inicial

#### Scenario: BaselineValue condicional

- GIVEN formulario de meta abierto
- WHEN selecciona direction `ascendente`
- THEN el input de `baselineValue` se oculta
- AND al seleccionar `descendente`, el input aparece con helper text "Valor al inicio de año (punto de partida)"

#### Scenario: Validar `targetValue` positivo

- GIVEN formulario inline de meta abierto
- WHEN intenta guardar con `targetValue <= 0`
- THEN el formulario inline muestra `alert-error` y no guarda

#### Scenario: Validar baselineValue para descendente

- GIVEN formulario de meta con direction `descendente`
- WHEN intenta guardar sin `baselineValue` o con `baselineValue <= targetValue`
- THEN el formulario muestra `alert-error` "El valor inicial debe ser mayor al objetivo"

#### Scenario: Sin botón eliminar meta fuera de inicio-anio

- GIVEN `cyclePhase !== 'inicio-anio'`
- WHEN se renderiza cualquier `GoalRow`
- THEN NO existe botón "Eliminar" en las acciones

### Requirement: Vinculación de KPIs a metas (N:M)

El sistema SHALL permitir crear una librería de KPIs (nombre + unidad + descripción) y vincular 0..N KPIs a cada meta. Un mismo KPI SHALL poder alimentar 1..N metas. **ADDED**: Los KPIs en la librería SHALL mostrar su `currentValue` y dirección. El `KpiBadge` SHALL calcular y mostrar el cumplimiento del KPI reutilizando la lógica de dirección.

#### Scenario: Vincular KPI a meta

- GIVEN meta en edición
- WHEN selecciona uno o más KPIs de la lista (checkboxes inline en `GoalRow`)
- THEN los KPIs aparecen como `KpiBadge` chips en la fila de la meta
- AND se crea/elimina un `GoalKpiLink` correspondiente en el store

#### Scenario: KPI reutilizado en varias metas

- GIVEN KPI "Ingresos trimestrales" vinculado a 2 metas
- WHEN edita la meta 1
- THEN la meta 2 sigue mostrando el `KpiBadge` "Ingresos trimestrales"
- AND la lista de "KPIs en uso" en la librería de KPIs lo refleja

#### Scenario: KpiBadge muestra cumplimiento

- GIVEN KPI "NPS clientes" con currentValue `85`, targetValue `80` y direction `ascendente`
- WHEN se renderiza el `KpiBadge` en una meta
- THEN el badge muestra "NPS clientes 100%" con color verde
- AND se muestra badge pequeño `+5` indicando exceso

#### Scenario: KpiBadge sin datos

- GIVEN KPI sin `currentValue`
- WHEN se renderiza el `KpiBadge`
- THEN el badge muestra solo el nombre del KPI sin porcentaje de cumplimiento

### Requirement: Doble validación 100% en tiempo real

El sistema SHALL mostrar en tiempo real dos `WeightIndicator` independientes. **UNCHANGED** — esta requirement no cambia con este change.

### Requirement: Jerarquía de edición (decisión #8)

El sistema SHALL detectar si el usuario activo es **dueño** de la asignación o **jefe/lector**. En modo lector SHALL ocultar todos los botones de crear/editar/eliminar y SHALL mostrar "Solicitar cambio" por categoría y por meta. **UNCHANGED** — esta requirement no cambia con este change.

## ADDED Requirements

### Requirement: Indicadores +/- en GoalRow

El `GoalRow` SHALL mostrar indicadores visuales de exceso/déficit cuando `currentValue` supera o empeora con respecto al `targetValue`. Los indicadores son puramente visuales.

#### Scenario: Indicador de exceso en meta ascendente

- GIVEN `GoalRow` con meta ascendente, targetValue `100` y currentValue `115`
- WHEN se renderiza la fila
- THEN se muestra badge `+15` en color verde (badge-success) junto al valor actual

#### Scenario: Indicador de déficit en meta ascendente

- GIVEN `GoalRow` con meta ascendente, targetValue `100` y currentValue `82`
- WHEN se renderiza la fila
- THEN se muestra badge `-18` en color ámbar (badge-warning) junto al valor actual

#### Scenario: Indicador de mejora en meta descendente

- GIVEN `GoalRow` con meta descendente, baselineValue `100`, targetValue `3` y currentValue `5`
- WHEN se renderiza la fila
- THEN se muestra badge `+95%` en color verde indicando mejora desde baseline

#### Scenario: Indicador de empeoramiento en meta descendente

- GIVEN `GoalRow` con meta descendente, baselineValue `100`, targetValue `3` y currentValue `110`
- WHEN se renderiza la fila
- THEN se muestra badge `-10` en color rojo (badge-error) indicando empeoramiento

#### Scenario: Sin indicador en valor exacto

- GIVEN `GoalRow` con currentValue == targetValue
- WHEN se renderiza la fila
- THEN NO se muestra indicador +/- (solo el badge de cumplimiento)

### Requirement: Score ponderado en UI

El frontend SHALL calcular y mostrar el score ponderado del empleado. El score se muestra en la pantalla de asignación como número 0–100.

#### Scenario: Score visible en asignación

- GIVEN empleado con categorías y metas configuradas
- WHEN navega a `/objetivos/asignacion`
- THEN se muestra el score ponderado actual como número 0–100
- AND el score se actualiza en tiempo real al editar pesos o progreso

#### Scenario: Score con desglose por categoría

- GIVEN empleado con 3 categorías
- WHEN hace clic en el score o expande la sección
- THEN se muestra el desglose: score por categoría con sus pesos individuales

### Requirement: Radio input de dirección en formulario de meta

El formulario de meta SHALL incluir un radio input para seleccionar la dirección (`ascendente` | `descendente`). El default es `ascendente`.

#### Scenario: Radio direction visible

- GIVEN formulario de meta abierto
- WHEN se renderiza el formulario
- THEN se muestra un radio group "Dirección" con opciones "Ascendente (↑)" y "Descendente (↓)"
- AND "Ascendente" está seleccionado por defecto

#### Scenario: Cambiar a descendente muestra baseline

- GIVEN formulario con direction `ascendente`
- WHEN cambia el radio a `descendente`
- THEN aparece un input "Valor inicial (baseline)" con helper text
- AND el input es obligatorio solo cuando direction es descendente

#### Scenario: Cambiar a ascendente oculta baseline

- GIVEN formulario con direction `descendente` y baselineValue definido
- WHEN cambia el radio a `ascendente`
- THEN el input de baselineValue se oculta
- AND el valor previo se ignora (no se persiste)
