# evaluation-lifecycle Specification (DELTA)

## MODIFIED Requirements

### Requirement: Restricciones de edición en medio de año (decisión #3)

En fase `avance`, el sistema SHALL permitir editar metas existentes y registrar avances, pero SHALL bloquear la eliminación de metas y la creación de metas nuevas. **ADDED**: En fase `avance`, los campos `direction`, `baselineValue`, `weight` y `targetValue` quedan congelados. Solo se permiten cambios en nombre, descripción, KPIs y `currentValue` (avance).

#### Scenario: Editar meta en medio año

- GIVEN ciclo en fase `avance`, meta en estado `en-seguimiento`
- WHEN empleado edita nombre, descripción o KPIs de la meta
- THEN los cambios se persisten
- AND la meta mantiene su estado `en-seguimiento`

#### Scenario: Registrar avance en meta

- GIVEN ciclo en fase `avance`, meta en estado `en-seguimiento`
- WHEN empleado registra un valor de avance (`currentValue`)
- THEN el avance se actualiza
- AND el semáforo/indicador de avance se recalcula según dirección

#### Scenario: Bloquear cambio de dirección en medio año

- GIVEN ciclo en fase `avance`, meta descendente con baselineValue `100`
- WHEN empleado intenta cambiar `direction` a `ascendente`
- THEN la acción está bloqueada
- AND el campo `direction` es read-only

#### Scenario: Bloquear cambio de targetValue en medio año

- GIVEN ciclo en fase `avance`, meta con targetValue `100`
- WHEN empleado intenta cambiar `targetValue`
- THEN la acción está bloqueada
- AND el campo `targetValue` es read-only

#### Scenario: Bloquear eliminación de meta en medio año

- GIVEN ciclo en fase `avance`
- WHEN empleado intenta eliminar una meta
- THEN la acción está bloqueada (botón deshabilitado o no renderizado)

#### Scenario: Bloquear creación de meta en medio año

- GIVEN ciclo en fase `avance`
- WHEN empleado intenta crear una meta nueva
- THEN la acción está bloqueada (botón "Nueva meta" no disponible)

### Requirement: Vías paralelas en fin de año (decisión #4)

En fase `cierre`, el sistema SHALL soportar tres vías de evaluación en paralelo. **ADDED**: Durante el cierre, el scoring ponderado SHALL calcularse automáticamente para determinar el rating final de metas. El score se usa como dato de entrada para la evaluación, no como reemplazo de la evaluación humana.

#### Scenario: Autoevaluación del empleado

- GIVEN ciclo en fase `cierre`, empleado con metas y competencias asignadas
- WHEN empleado completa su autoevaluación
- THEN registra calificación 1–5 por competencia y comentarios de cierre de metas
- AND su evaluación pasa a estado `completada`
- AND el score ponderado se calcula automáticamente y se muestra como referencia

#### Scenario: Score como referencia en cierre

- GIVEN empleado en fase `cierre` con score ponderado `78`
- WHEN se renderiza la pantalla de cierre
- THEN se muestra "Score de metas: 78/100" como dato informativo
- AND el empleado puede agregar comentarios sobre el score (no calificar el score directamente)

#### Scenario: Jefe califica 9×9

- GIVEN ciclo en fase `cierre`, jefe con evaluados
- WHEN jefe abre la matriz 9×9
- THEN puede calificar desempeño y potencial de cada evaluado
- AND las calificaciones 9×9 son independientes del score de metas

#### Scenario: RH evalúa formalmente

- GIVEN ciclo en fase `cierre`, RH con empleados asignados
- WHEN RH completa la evaluación formal de un empleado
- THEN registra calificación de competencias y cierre
- AND la evaluación formal es la definitiva para el empleado
- AND el score de metas es un dato de referencia, no determinante

## ADDED Requirements

### Requirement: Score ponderado calculado en cierre

El sistema SHALL calcular automáticamente el score ponderado del empleado al entrar en fase `cierre`. El score se almacena como referencia en la evaluación.

#### Scenario: Score calculado al inicio de cierre

- GIVEN empleado con categorías y metas en fase `avance`
- WHEN el ciclo transiciona a `cierre`
- THEN se calcula el score ponderado = `Σ(cat_weight/100 × Σ(goal_weight/100 × goalProgressPercent))`
- AND el score se almacena en la evaluación como referencia

#### Scenario: Score se recalcula si cambia progreso

- GIVEN empleado en fase `cierre` con score `78`
- WHEN (si aplica) se actualiza algún `currentValue`
- THEN el score se recalcula automáticamente
- AND el nuevo valor se muestra en la UI

### Requirement: KPIs actualizados por etapa

Los KPIs se actualizan externamente una vez por etapa (no por los usuarios). El sistema SHALL soportar la actualización de `currentValue` de KPIs vía endpoint PATCH.

#### Scenario: KPIs se actualizan entre fases

- GIVEN KPI "NPS clientes" sin `currentValue`
- WHEN fuentes externas envían `PATCH /kpis/{id}/value` con `current_value: 85`
- THEN el KPI se actualiza
- AND el KPI vinculado a metas refleja el nuevo valor en la UI

#### Scenario: KPIs no se editan en UI

- GIVEN empleado en cualquier fase
- WHEN accede a la pantalla de metas
- THEN los KPIs se muestran con su `currentValue` actual
- AND NO hay inputs para editar `currentValue` de KPIs (solo el endpoint PATCH)
