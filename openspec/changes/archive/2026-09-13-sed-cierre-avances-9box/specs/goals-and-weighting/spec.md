# goals-and-weighting — Delta spec (sed-cierre-avances-9box)

## MODIFIED Requirements

### Requirement: Restricciones de edición por fase (decisión #3)

El sistema SHALL restringir la edición de metas según la fase del ciclo activo. En medio de año (`avance`), el sistema SHALL prohibir eliminar metas. El registro de avances SHALL estar permitido en `avance` (incl. alias `medio-anio`) y en `cierre`; en `cierre` el resto de campos de meta siguen siendo de solo lectura.

#### Scenario: Inicio de año — CRUD completo

- GIVEN ciclo en fase `asignacion`
- WHEN empleado edita sus metas
- THEN puede: crear, editar (todos los campos), eliminar metas y categorías, modificar ponderaciones, vincular/desvincular KPIs

#### Scenario: Medio de año — solo edición parcial

- GIVEN ciclo en fase `avance`
- WHEN empleado edita sus metas
- THEN puede: editar campos de meta (nombre, descripción, targetValue, KPIs), registrar avances
- AND NO puede: crear metas nuevas, eliminar metas, crear/eliminar categorías, modificar ponderaciones

#### Scenario: Fin de año — sin edición de metas

- GIVEN ciclo en fase `cierre`
- WHEN empleado accede a sus metas
- THEN las metas son de solo lectura, excepto el registro de avances que sigue permitido
- AND solo puede realizar autoevaluación (calificar competencias 1–5)

#### Scenario: Cierre — avances permitidos, resto solo lectura

- GIVEN ciclo en fase `cierre`
- WHEN empleado registra avance en su meta
- THEN el avance persiste
- AND NO puede: crear/editar/eliminar metas, categorías, ponderaciones ni KPIs

#### Scenario: Asignación — avances bloqueados

- GIVEN ciclo en fase `asignacion`
- WHEN empleado intenta registrar avance
- THEN el sistema lo rechaza (fase no permite progreso)

## ADDED Requirements

### Requirement: Separación total por fase avance/cierre

El sistema SHALL tratar `avance` y `cierre` como evaluaciones independientes: progreso, snapshot y comentarios de una fase SHALL NOT contaminar la otra. `phaseKind` deriva de la fase del ciclo (`EmployeeEvaluationDetail`); `GoalClosureCard` SHALL renderizar single card de la fase actual.

#### Scenario: Fases independientes

- WHEN se registra avance o comentario en `cierre`
- THEN `avance` permanece inmutable y viceversa.

### Requirement: Comentarios de metas separados por fase y rol

Los comentarios de metas SHALL separarse por fase (`avance`/`cierre`) y rol (empleado/jefa). Solo la fase actual SHALL ser editable; la fase previa SHALL ser inmutable (solo lectura).

#### Scenario: Fase actual editable, previa inmutable

- GIVEN ciclo en `cierre`
- WHEN empleado o jefa accede a comentarios de `avance`
- THEN los ve en solo lectura y NO puede editarlos; solo los comentarios de `cierre` son editables.

### Requirement: Permisos jefa en comentarios de metas

Jefa SHALL ver comentarios de metas de sus evaluados en `avance` y `cierre`; SHALL escribir comentarios de jefa solo en la fase actual. Empleado SHALL ver/escribir solo sus comentarios en fase actual (código actual = verdad).

#### Scenario: Matriz de permisos metas

- GIVEN rol `jefa` y fase actual `cierre`
- WHEN accede a comentarios
- THEN ve comentarios empleado+jefa de `avance` (inmutables) y ve/escribe comentarios de `cierre`.
- GIVEN rol `empleado` y fase actual `cierre`
- WHEN accede a comentarios
- THEN ve comentarios de `avance` (inmutables) y escribe solo sus comentarios de `cierre`.

### Requirement: Cierre activo hasta nuevo ciclo (sin finished_at)

`UpdatePhase` SHALL NOT setear `finished_at` al entrar a `cierre`. `finished_at` SHALL setearse solo al crear el nuevo ciclo anual en `asignacion`. El ciclo en `cierre` SHALL seguir resoluble por `GetActiveCycleID WHERE finished_at IS NULL`. Botón de cierre manual fuera de scope.

#### Scenario: Cierre sigue activo

- GIVEN ciclo en fase `cierre` con `finished_at IS NULL`
- WHEN se llama `GetActiveCycleID`
- THEN retorna el ciclo (válido para avances, autoeval y 9-box).
- WHEN se crea el nuevo ciclo anual en `asignacion`
- THEN el ciclo anterior se marca `finished_at` y deja de ser activo.
