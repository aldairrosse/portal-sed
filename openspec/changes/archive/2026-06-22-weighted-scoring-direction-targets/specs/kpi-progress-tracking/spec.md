# kpi-progress-tracking Specification (NEW)

## Purpose

Definir el **tracking de progreso de KPIs** con `currentValue`, `direction`, y endpoint de **actualización externa**. Los KPIs se actualizan una vez por etapa desde fuentes externas, no por los usuarios.

## Data Model

| Entity | New Fields | Notes |
|--------|-----------|-------|
| **KPI** | `currentValue` (float, nullable), `direction` (`ascendente` \| `descendente`) | `currentValue` se actualiza vía PATCH externo. `direction` default `ascendente`. |

## Requirements

### Requirement: Campo currentValue en KPIs

Cada KPI SHALL tener un campo `currentValue` que representa el valor actual del indicador. Este campo se actualiza externamente, no por los usuarios de la UI.

#### Scenario: KPI con valor actual

- GIVEN KPI "NPS clientes" con direction `ascendente` y targetValue `80`
- WHEN se actualiza `currentValue` a `85`
- THEN el KPI muestra su valor actual `85` junto al target `80`
- AND se calcula el cumplimiento del KPI

#### Scenario: KPI sin valor actual

- GIVEN KPI recién creado sin `currentValue`
- WHEN se renderiza en la UI
- THEN el KPI muestra "Sin datos" o badge neutral
- AND NO se calcula cumplimiento

### Requirement: Dirección de KPI

Los KPIs SHALL tener un campo `direction` con la misma lógica que las metas: `ascendente` (más es mejor) o `descendente` (menos es mejor).

#### Scenario: KPI ascendente

- GIVEN KPI "Ingresos trimestrales" con direction `ascendente`
- WHEN currentValue `500000` supera targetValue `400000`
- THEN el cumplimiento es `100%` (clamp)
- AND se muestra badge `+100000` verde

#### Scenario: KPI descendente

- GIVEN KPI "Tasa de rotación" con direction `descendente`, baselineValue `15` y targetValue `5`
- WHEN currentValue es `8`
- THEN el cumplimiento es `70%` ((15 - 8) / (15 - 5) * 100)
- AND se muestra badge verde indicando mejora

### Requirement: Endpoint PATCH para actualización externa de KPIs

El sistema SHALL exponer un endpoint `PATCH /api/v1/kpis/{kpiId}/value` para que fuentes externas actualicen el `currentValue` de un KPI.

#### Scenario: Actualización exitosa

- GIVEN KPI "NPS clientes" con id `kpi-123`
- WHEN se envía `PATCH /api/v1/kpis/kpi-123/value` con body `{ "current_value": 85 }`
- THEN el KPI se actualiza con `currentValue = 85`
- AND se retorna el KPI actualizado

#### Scenario: KPI no existe

- GIVEN no existe KPI con id `kpi-999`
- WHEN se envía `PATCH /api/v1/kpis/kpi-999/value`
- THEN el sistema retorna error 404 con code `KPI_NOT_FOUND`

#### Scenario: Valor negativo

- GIVEN KPI "NPS clientes"
- WHEN se envía `current_value: -5`
- THEN el sistema rechaza con error 400 y message "current_value must be non-negative"

#### Scenario: Actualización idempotente

- GIVEN KPI "NPS clientes" con currentValue `80`
- WHEN se envía dos veces `PATCH` con `current_value: 85`
- THEN el KPI queda con `currentValue = 85` (no duplica ni acumula)

### Requirement: Cumplimiento de KPI reutiliza lógica de dirección

El cálculo de cumplimiento de KPIs SHALL usar la misma función `progressPercent()` que las metas. No se duplica la lógica.

#### Scenario: KPI ascendente parcial

- GIVEN KPI con direction `ascendente`, targetValue `100` y currentValue `75`
- WHEN se calcula el cumplimiento
- THEN resultado = `75%`

#### Scenario: KPI descendente parcial

- GIVEN KPI con direction `descendente`, baselineValue `20`, targetValue `5` y currentValue `10`
- WHEN se calcula el cumplimiento
- THEN resultado = `66.67%` ((20 - 10) / (20 - 5) * 100)

### Requirement: Indicadores +/- en KPIs

El frontend SHALL mostrar indicadores +/- en KPIs con la misma lógica que las metas: cuando `currentValue` supera o empeora con respecto al target.

#### Scenario: KPI ascendente supera target

- GIVEN KPI con direction `ascendente`, targetValue `80` y currentValue `92`
- WHEN se renderiza el KPI
- THEN se muestra badge `+12` verde

#### Scenario: KPI descendente mejora

- GIVEN KPI con direction `descendente`, baselineValue `20`, targetValue `5` y currentValue `7`
- WHEN se renderiza el KPI
- THEN se muestra badge `+13` verde (mejoró 13 desde baseline)

#### Scenario: KPI sin currentValue

- GIVEN KPI sin `currentValue` definido
- WHEN se renderiza el KPI
- THEN NO se muestra indicador +/- (solo el badge del KPI)

## Non-goals

- **Autenticación del endpoint PATCH**: se implementará cuando auth esté listo.
- **Batch update de KPIs**: el endpoint es por un solo KPI; batch se puede agregar después.
- **Historial de valores de KPI**: no se almacena历史; solo se guarda el último valor.
- **Dashboard de KPIs**: los KPIs se muestran inline en metas, no en pantalla dedicada.
