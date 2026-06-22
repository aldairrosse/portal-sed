# goals-and-weighting Specification (DELTA)

## MODIFIED Requirements

### Requirement: Unidades de medida para metas

Cada meta SHALL tener una `unit` que define su tipo de medida: `porcentaje` (0–100), `moneda` (monto con símbolo) o `numero` (entero/decimal). La unidad determina cómo se expresa el `targetValue` y el avance. **ADDED**: Cada meta SHALL tener un campo `direction` (`ascendente` | `descendente`, default `ascendente`) que define si "más es mejor" o "menos es mejor". Para metas descendentes, se REQUIERE un `baselineValue` que representa el valor inicial al inicio de año.

#### Scenario: Crear meta con unidad porcentaje

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Reducir tasa de rotación" con unit `porcentaje`, targetValue `15`, direction `descendente` y baselineValue `25`
- THEN la meta se muestra con sufijo "%" y dirección descendente (flecha ↓)
- AND el avance se calcula como `(25 - currentValue) / (25 - 15) * 100`

#### Scenario: Crear meta con unidad moneda

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Incrementar ingresos" con unit `moneda`, targetValue `500000` y direction `ascendente`
- THEN la meta se muestra con formato monetario y dirección ascendente
- AND el avance se registra como monto

#### Scenario: Unidad es independiente del KPI

- GIVEN KPI "Ingresos trimestrales" con unit `moneda` y direction `ascendente`
- WHEN se vincula a meta "Crecimiento sostenible" con unit `porcentaje` y direction `descendente`
- THEN la vinculación es válida
- AND la meta mantiene su unidad y dirección propias, no hereda las del KPI

#### Scenario: Baseline requerido para descendente

- GIVEN empleado creando meta descendente
- WHEN intenta guardar sin `baselineValue`
- THEN el sistema rechaza con error "El valor inicial es requerido para metas descendentes"

#### Scenario: Baseline ignorado para ascendente

- GIVEN empleado creando meta ascendente
- WHEN envía `baselineValue` junto con la meta
- THEN el sistema ignora `baselineValue` (no se persiste)

### Requirement: Doble ponderación 100% (decisión #1)

El sistema SHALL implementar doble ponderación: las categorías suman 100% del empleado, y las metas dentro de cada categoría suman 100% de esa categoría. Ambas sumas SHALL ser validadas independientemente. **ADDED**: El sistema SHALL calcular un **scoring ponderado completo** que combine categoría × meta × cumplimiento según dirección. Score = `Σ(cat_weight/100 × Σ(goal_weight/100 × goalProgressPercent))`. El score SHALL estar en rango 0–100.

#### Scenario: Suma de categorías = 100%

- GIVEN empleado con 3 categorías de pesos 40, 35, 25
- WHEN valida su asignación
- THEN la suma de categorías es 100% ✓
- AND el sistema permite guardar

#### Scenario: Suma de categorías ≠ 100%

- GIVEN empleado con 2 categorías de pesos 60, 30
- WHEN valida su asignación
- THEN la suma es 90%, no cumple
- AND el sistema bloquea el guardado
- AND muestra feedback indicando el déficit (faltan 10%)

#### Scenario: Suma de metas dentro de categoría = 100%

- GIVEN categoría "Resultados de negocio" con peso 40% y 2 metas de pesos 60 y 40
- WHEN valida la categoría
- THEN la suma de metas es 100% ✓

#### Scenario: Suma de metas dentro de categoría ≠ 100%

- GIVEN categoría con 2 metas de pesos 70 y 20
- WHEN valida la categoría
- THEN la suma es 90%, no cumple
- AND el sistema bloquea el guardado de la categoría
- AND muestra feedback indicando el déficit

#### Scenario: Categoría vacía (sin metas)

- GIVEN categoría recién creada sin metas
- WHEN valida la categoría
- THEN la suma de metas es 0%
- AND se muestra badge de advertencia "Sin metas"
- AND la categoría vacía no bloquea la validación global (decisión transitoria)

#### Scenario: Tolerancia flotante

- GIVEN pesos que suman 99.99 o 100.01
- WHEN valida la suma
- THEN el sistema acepta como válido (tolerancia ε = 0.01)

#### Scenario: Scoring ponderado completo

- GIVEN empleado con categoría A (peso 60%) con 2 metas (pesos 50/50, cumplimiento 80%/60%) y categoría B (peso 40%) con 1 meta (peso 100%, cumplimiento 90%)
- WHEN se calcula el score
- THEN score = 0.6 * (0.5 * 80 + 0.5 * 60) + 0.4 * (1.0 * 90) = 78
- AND el score se muestra en la UI de cierre como número 0–100

### Requirement: Vinculación KPIs a metas (decisión #6)

KPIs son **indicadores** reutilizables (numérico, porcentaje o moneda) que pueden vincularse a 1..N metas. Cada meta puede tener 0..N KPIs asociados. **ADDED**: Los KPIs SHALL tener un campo `direction` (`ascendente` | `descendente`) y un campo `currentValue` que se actualiza externamente. El cumplimiento de KPIs reutiliza la misma lógica de dirección que las metas.

#### Scenario: Vincular KPI existente a meta

- GIVEN KPI "NPS clientes" (unit: `porcentaje`, direction: `ascendente`) y meta "Mejorar satisfacción" (direction: `ascendente`)
- WHEN se vincula el KPI a la meta
- THEN la meta muestra el badge del KPI
- AND el KPI puede ser consultado desde la meta y viceversa

#### Scenario: KPI alimenta múltiples metas

- GIVEN KPI "Ingresos trimestrales" vinculado a 3 metas
- WHEN se consulta el KPI
- THEN muestra las 3 metas asociadas
- AND eliminar una meta elimina solo el vínculo, no el KPI

#### Scenario: Meta sin KPI

- GIVEN meta "Desarrollo personal"
- WHEN se guarda sin vincular ningún KPI
- THEN la meta es válida
- AND no muestra badges de KPI

#### Scenario: KPI con progreso externo

- GIVEN KPI "NPS clientes" con currentValue `85` y targetValue `80`
- WHEN se consulta el KPI vinculado a una meta
- THEN el KPI muestra su cumplimiento `100%` (clamp)
- AND se muestra badge `+5` verde si supera el target

### Requirement: Restricciones de edición por fase (decisión #3)

Las reglas de edición de metas dependen de la fase del ciclo activo. En medio de año (`avance`), está **prohibido eliminar** metas. **ADDED**: En fase `asignacion`, el campo `direction` y `baselineValue` son editables. En fase `avance`, `direction` y `baselineValue` quedan congelados (igual que `weight` y `targetValue`).

#### Scenario: Inicio de año — CRUD completo

- GIVEN ciclo en fase `asignacion`
- WHEN empleado edita sus metas
- THEN puede: crear, editar (todos los campos incluyendo `direction` y `baselineValue`), eliminar metas y categorías, modificar ponderaciones, vincular/desvincular KPIs

#### Scenario: Medio de año — solo edición parcial

- GIVEN ciclo en fase `avance`
- WHEN empleado edita sus metas
- THEN puede: editar campos de meta (nombre, descripción, KPIs), registrar avances
- AND NO puede: crear metas nuevas, eliminar metas, crear/eliminar categorías, modificar ponderaciones, modificar `direction`, modificar `baselineValue`, modificar `targetValue`

#### Scenario: Fin de año — sin edición de metas

- GIVEN ciclo en fase `cierre`
- WHEN empleado accede a sus metas
- THEN las metas son de solo lectura
- AND solo puede realizar autoevaluación (calificar competencias 1–5)
