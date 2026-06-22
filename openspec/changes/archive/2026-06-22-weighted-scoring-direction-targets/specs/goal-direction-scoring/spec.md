# goal-direction-scoring Specification (NEW)

## Purpose

Definir la **dirección de metas** (ascendente/descendente), el cálculo de **cumplimiento por dirección**, el **scoring ponderado completo** del empleado, y los **indicadores +/-** de exceso/déficit.

## Data Model

| Entity | New Fields | Notes |
|--------|-----------|-------|
| **Goal** | `direction` (`ascendente` \| `descendente`), `baselineValue` (float, nullable) | `direction` default `ascendente`. `baselineValue` requerido solo para descendente. |
| **KPI** | `direction` (`ascendente` \| `descendente`) | Misma lógica que Goal. Default `ascendente`. |

## Requirements

### Requirement: Dirección de meta

Cada meta SHALL tener un campo `direction` con valores `ascendente` (default) o `descendente`. Ascendente significa "más es mejor" (default actual). Descendente significa "menos es mejor" (ej: reducir ausentismo, reducir quejas).

#### Scenario: Crear meta ascendente

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Alcanzar ventas" con direction `ascendente` y targetValue `1000000`
- THEN la meta se muestra con dirección ascendente (flecha hacia arriba o badge "↑")
- AND el cumplimiento se calcula como `currentValue / targetValue * 100`

#### Scenario: Crear meta descendente

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Reducir ausentismo" con direction `descendente`, baselineValue `100` y targetValue `3`
- THEN la meta se muestra con dirección descendente (flecha hacia abajo o badge "↓")
- AND el cumplimiento se calcula como `(baselineValue - currentValue) / (baselineValue - targetValue) * 100`

#### Scenario: Default ascendente

- GIVEN meta creada sin especificar direction
- WHEN se guarda
- THEN `direction` se establece en `ascendente` automáticamente

### Requirement: Baseline value para metas descendentes

Las metas descendentes SHALL tener un `baselineValue` que representa el valor inicial al inicio de año. Este valor es el punto de partida (0% de cumplimiento). El `targetValue` es el punto deseado (100% de cumplimiento).

#### Scenario: Baseline requerido para descendente

- GIVEN empleado creando meta descendente
- WHEN intenta guardar sin `baselineValue`
- THEN el sistema rechaza con error "El valor inicial es requerido para metas descendentes"

#### Scenario: Baseline ignorado para ascendente

- GIVEN empleado creando meta ascendente
- WHEN envía `baselineValue` junto con la meta
- THEN el sistema ignora `baselineValue` (no se persiste o se guarda como NULL)

#### Scenario: Baseline mayor que target

- GIVEN meta descendente con baselineValue `100` y targetValue `3`
- WHEN se calcula el cumplimiento
- THEN baselineValue (100) > targetValue (3) es válido
- AND el rango de progreso es de 100 → 3

### Requirement: Cálculo de cumplimiento por dirección

El sistema SHALL calcular el porcentaje de cumplimiento de una meta según su dirección.

#### Scenario: Cumplimiento ascendente parcial

- GIVEN meta ascendente con targetValue `1000000` y currentValue `650000`
- WHEN se calcula el cumplimiento
- THEN el resultado es `65%` (650000 / 1000000 * 100)

#### Scenario: Cumplimiento ascendente completo

- GIVEN meta ascendente con targetValue `100` y currentValue `120`
- WHEN se calcula el cumplimiento
- THEN el resultado es `100%` (clamp a 100, no puede exceder)

#### Scenario: Cumplimiento descendente parcial

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `50`
- WHEN se calcula el cumplimiento
- THEN el resultado es `51.02%` ((100 - 50) / (100 - 3) * 100)

#### Scenario: Cumplimiento descendente completo

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `3`
- WHEN se calcula el cumplimiento
- THEN el resultado es `100%`

#### Scenario: Cumplimiento descendente sin progreso

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `100`
- WHEN se calcula el cumplimiento
- THEN el resultado es `0%` (no ha mejorado desde baseline)

#### Scenario: Cumplimiento descendente peor que baseline

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `120`
- WHEN se calcula el cumplimiento
- THEN el resultado es `0%` (clamp a 0, no puede ser negativo)
- AND se muestra indicador `-20` (empeoró 20 unidades desde baseline)

### Requirement: Scoring ponderado completo

El sistema SHALL calcular un score numérico ponderado que combine categoría × meta × cumplimiento. El score SHALL estar en rango 0–100.

#### Scenario: Score de empleado con dos categorías

- GIVEN empleado con categoría A (peso 60%) con 2 metas (pesos 50/50, cumplimiento 80%/60%) y categoría B (peso 40%) con 1 meta (peso 100%, cumplimiento 90%)
- WHEN se calcula el score
- THEN score = 0.6 * (0.5 * 80 + 0.5 * 60) + 0.4 * (1.0 * 90) = 0.6 * 70 + 0.4 * 90 = 42 + 36 = 78

#### Scenario: Score con categorías vacías

- GIVEN empleado con 3 categorías pero solo 2 con metas
- WHEN se calcula el score
- THEN las categorías vacías contribuyen 0 al score
- AND el score se calcula solo con las categorías que tienen metas

#### Scenario: Score siempre en rango 0–100

- GIVEN empleado con cumplimiento 120% en una meta (ascendente, clamp a 100)
- WHEN se calcula el score
- THEN el score NO excede 100

### Requirement: Indicadores +/- de exceso/déficit

El frontend SHALL mostrar indicadores visuales cuando `currentValue` supera o empeora con respecto al target. Los indicadores son puramente visuales y NO se persisten.

#### Scenario: Indicador de exceso en meta ascendente

- GIVEN meta ascendente con targetValue `100` y currentValue `115`
- WHEN se renderiza la meta
- THEN se muestra badge `+15` en color verde (superó el target)

#### Scenario: Indicador de déficit en meta ascendente

- GIVEN meta ascendente con targetValue `100` y currentValue `82`
- WHEN se renderiza la meta
- THEN se muestra badge `-18` en color ámbar (no llegó al target)

#### Scenario: Indicador de mejora en meta descendente

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `5`
- WHEN se renderiza la meta
- THEN se muestra badge `+95%` en color verde (mejoró 95 puntos desde baseline)

#### Scenario: Indicador de empeoramiento en meta descendente

- GIVEN meta descendente con baselineValue `100`, targetValue `3` y currentValue `110`
- WHEN se renderiza la meta
- THEN se muestra badge `-10` en color rojo (empeoró 10 unidades desde baseline)

#### Scenario: Sin indicador cuando cumple exacto

- GIVEN meta con currentValue == targetValue
- WHEN se renderiza la meta
- THEN NO se muestra indicador +/- (solo el badge de cumplimiento 100%)

## Non-goals

- **Persistencia de indicadores**: los +/- son cálculos frontend-only.
- **Scoring en backend de cierre**: el score se calcula on-demand, no se persiste como campo en evaluación.
- **Alertas de exceso**: no se envían notificaciones cuando un empleado supera el target.
