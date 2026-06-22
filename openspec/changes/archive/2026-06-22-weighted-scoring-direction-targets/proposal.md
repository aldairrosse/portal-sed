## Why

El sistema actual de metas tiene un `targetValue` y un `currentValue`, pero carece de dos capacidades fundamentales para una evaluación de desempeño seria:

1. **Sin scoring ponderado completo**: no existe un cálculo que combine el peso de categoría × peso de meta × porcentaje de cumplimiento para obtener un score numérico final del empleado. Solo hay un promedio simple por categoría en el frontend.
2. **Sin dirección de meta**: todas asumen que "más es mejor". Metas como "reducir ausentismo" o "reducir quejas" se rompen porque `currentValue > targetValue` se interpreta como "exceso" cuando en realidad es el opuesto del cumplimiento. Los KPIs tampoco tienen dirección ni tracking de progreso.
3. **KPIs sin progreso**: los KPIs se vinculan a metas pero no se actualizan externamente con valores reales. No hay forma de ver el avance de un KPI independiente de la meta.

Este cambio cierra el gap entre "asignar metas" y "evaluar cumplimiento real con un número".

## What Changes

- **Campo `direction`** en `Goal`: `'ascendente' | 'descendente'`. Ascendente = más es mejor (default). Descendente = menos es mejor. Radio input en formulario de asignación.
- **Cálculo de cumplimiento por dirección**: para ascendente, `progress% = min(current/target * 100, 100)`. Para descendente, se invierte: el `currentValue` inicial (inicio de año) es el 0% y el `targetValue` es el 100%.
- **Campo `baselineValue`** en `Goal`: valor inicial al inicio de año para metas descendentes. Obligatorio cuando `direction === 'descendente'`. Se registra una sola vez en fase `asignacion`.
- **Scoring ponderado completo**: `score = Σ(categoría_weight/100 × Σ(meta_weight/100 × meta_progress%))`. Se calcula en backend y frontend.
- **Campo `direction`** en `KPI`: `'ascendente' | 'descendente'`. Los KPIs también tienen dirección propia.
- **Campo `currentValue`** en KPI (nuevo endpoint): los KPIs se actualizan externamente una vez por etapa. Nuevo endpoint `PATCH /kpis/{kpiId}/value`.
- **Cumplimiento de KPIs**: reutiliza la misma lógica de dirección que las metas. Muestra badge +/- cuando supera o queda por debajo del target.
- **Indicadores +/-** en frontend: cuando `currentValue` supera o empeora con respecto al target, se muestra la diferencia numérica con signo `+` o `-` (solo visual, no afecta scoring).
- **Migración de BD**: campo `direction` y `baseline_value` en tabla `goals`; campo `direction` y `current_value` en tabla `kp_is`.

## Capabilities

### New Capabilities
- `goal-direction-scoring`: Dirección de metas (ascendente/descendente), baseline value, scoring ponderado completo, indicadores +/- de cumplimiento.
- `kpi-progress-tracking`: Progreso de KPIs con currentValue, dirección, endpoint de actualización externa, y cálculo de cumplimiento reutilizando lógica de dirección.

### Modified Capabilities
- `goals-and-weighting`: Agregar campo `direction` al data model de Goal, agregar `baselineValue`, actualizar reglas de ponderación para incluir scoring. Agregar escenarios de dirección.
- `goal-assignment-ui`: Radio input de dirección en formulario de meta, indicadores +/- en GoalRow, cálculo de score por categoría y global en UI.
- `evaluation-lifecycle`: El scoring ponderado se calcula en fase `cierre` para determinar el rating final de metas.

## Impact

- **Backend (Go)**:
  - Migración Ent: `goals` agrega `direction` (enum) y `baseline_value` (float64 nullable); `kp_is` agrega `direction` (enum) y `current_value` (float64).
  - Nuevo handler: `PATCH /api/v1/kpis/{kpiId}/value` para actualización externa de KPIs.
  - Nuevo servicio: `ScoringService` que calcula el score ponderado por empleado.
  - Modificar `GoalService.CreateGoal` para requerir `baselineValue` cuando `direction === 'descendente'`.
  - Modificar `ProgressService.UpdateGoalProgress` para manejar inversión descendente.

- **Frontend (Svelte)**:
  - Formulario de meta: agregar radio `direction` (ascendente/descendente), input condicional `baselineValue`.
  - `GoalRow`: indicador +/- cuando `currentValue` supera/empeora target.
  - `KpiBadge` o nuevo componente: mostrar progreso de KPI con dirección y cumplimiento.
  - Store: calcular score ponderado en `getCategoryProgressAverage` y nuevo `getWeightedScore`.
  - Fixture KPIs: agregar campo `direction` y `currentValue`.

- **API OpenSpec**:
  - Actualizar schema de Goal para `direction` y `baseline_value`.
  - Actualizar schema de KPI para `direction` y `current_value`.
  - Nuevo endpoint PATCH para KPI value update.

- **Dependencias**: No nuevas. Todo con stack actual (Go + Ent, Svelte 5 + DaisyUI).
