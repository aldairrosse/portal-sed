# evaluation-lifecycle — Delta spec (sed-evaluacion-10-mejoras)

## ADDED Requirements

### Requirement: Evaluation con fila propia por employee, cycle y phase

El sistema SHALL buscar y crear Evaluations por la tupla (employee,cycle,phase) mediante `Ensure`/`FindBy(employee,cycle,phase)`. Escribir en `avance` SHALL afectar solo la fila `avance`; escribir en `cierre` SHALL afectar solo la fila `cierre`. El sistema SHALL NOT reutilizar la fila de otra fase.

#### Scenario: Avance escribe solo avance

- **GIVEN** existen filas `avance` y `cierre` para (employee,cycle)
- **WHEN** se escribe avance/ratings en fase `avance`
- **THEN** solo la fila `avance` cambia y la fila `cierre` queda intacta

#### Scenario: Cierre escribe solo cierre

- **GIVEN** existen filas `avance` y `cierre` para (employee,cycle)
- **WHEN** se escribe en fase `cierre`
- **THEN** solo la fila `cierre` cambia y la fila `avance` queda intacta

#### Scenario: Auto-creación de Evaluation al abrir cierre

- **GIVEN** no existe Evaluation para (employee,cycle,`cierre`)
- **WHEN** se abre la vista de cierre
- **THEN** el sistema crea la fila `cierre` automáticamente (idempotente, concurrente OK) y la muestra sin error 404

### Requirement: Avances, ratings y comentarios por rol y fase separados e independientes

Los avances/ratings y los comentarios SHALL almacenarse por (rol,fase) de forma separada e independiente (self, jefe, RH × avance/cierre). Editar el comentario de un rol en una fase SHALL NOT alterar el de otro rol ni el de otra fase. Los permisos de ver/escribir comentarios de metas/competencias de la jefa SHALL regirse por el código del repo (`EmployeeEvaluationDetail`).

#### Scenario: Comentario de jefa en avance no toca cierre

- **WHEN** la jefa escribe un comentario de meta/competencia en fase `avance`
- **THEN** el comentario de fase `cierre` del mismo rol queda intacto

#### Scenario: Permisos de jefa según código

- **GIVEN** el código de `EmployeeEvaluationDetail` define ver/escribir comentarios para la jefa
- **WHEN** hay conflicto entre spec y código
- **THEN** prevalece el código y la spec se corrige

### Requirement: Revert cierre→avance conserva ambas fases y deja editable la fase actual

El sistema SHALL permitir revertir de `cierre` a `avance` conservando snapshot y data de ambas fases (sin borrados). Tras el revert, la fase `avance` SHALL quedar editable (avances/ratings + comentarios de la fase actual editables).

#### Scenario: Revert conserva ambas filas

- **GIVEN** una Evaluation con filas `avance` y `cierre` con data
- **WHEN** RH revierte `cierre`→`avance`
- **THEN** ambas filas conservan su data y la fase activa es `avance` editable

#### Scenario: Editable tras revert

- **GIVEN** se revirtió `cierre`→`avance`
- **WHEN** se escribe avance o comentario en fase `avance`
- **THEN** el sistema acepta la escritura en la fila `avance` sin error de fase

### Requirement: Gates de escritura por fase

La escritura de competencias (self y manager/RH) SHALL estar permitida en fases `avance`/`medio-anio` y `cierre`; Submit/Finalize SHALL seguir exigiendo `cierre`. Fuera de fases escribibles el sistema SHALL responder 409 PHASE_NOT_ADVANCEABLE. La escritura RH (`POST`/`PUT /evaluations/{id}/rh-evaluation`) SHALL estar autorizada solo a RH (permiso `eval:rh`) o al jefe asignado; sin sesión SHALL responder 401, no autorizado SHALL responder 403.

#### Scenario: Self 1-5 en avance OK

- **WHEN** el empleado actualiza su autoevaluación con rating 1-5 en fase `avance`
- **THEN** el sistema acepta la escritura sin error de fase

#### Scenario: Submit en avance responde 409

- **WHEN** se intenta Submit (self o RH) o Finalize en fase `avance`
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE (operación solo `cierre`)

#### Scenario: Fuera de fase escribible responde 409

- **WHEN** se actualiza self o RH fuera de `avance`/`medio-anio`/`cierre` (p. ej. `asignacion`)
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE

#### Scenario: RH o jefe asignado escribe OK; tercero 403; sin sesión 401

- **WHEN** un usuario con permiso `eval:rh` o el manager del evaluado escribe la evaluación RH
- **THEN** el sistema acepta (sujeto a fase)
- **WHEN** un tercero autenticado no autorizado intenta escribirla
- **THEN** responde 403; sin sesión responde 401

### Requirement: Avance con progreso, autoeval y comentarios visibles

En fase `avance`, el sistema SHALL habilitar registro de progreso y autoevaluación del empleado, y SHALL mostrar los comentarios del jefe y de RH.

#### Scenario: Progreso y autoevaluación habilitados en avance

- **WHEN** el empleado abre su evaluación en fase `avance`
- **THEN** puede registrar progreso en sus metas y completar su autoevaluación

#### Scenario: Comentarios de jefe y RH visibles en avance

- **GIVEN** existen comentarios del jefe o de RH
- **WHEN** el empleado consulta su evaluación en fase `avance`
- **THEN** los comentarios del jefe y de RH son visibles

## Notes (referencia, no implementar aquí)

- Backfill por fase: fuera de alcance; este modelo por fase no lo contradice ni lo impide.
- Quitar `finished_at` en cierre: solo referencia; no se toca en este change.
