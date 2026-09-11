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
