## ADDED Requirements

### Requirement: Avance phase gates and auto-create
En fase `avance`, el sistema SHALL habilitar el registro de progreso y la autoevaluación del empleado, y SHALL mostrar los comentarios del jefe y de RH. En fase `cierre`, el sistema SHALL auto-crear la Evaluation si no existe al abrir el cierre, en lugar de devolver 404. El sistema SHALL permitir revertir de `cierre` a `avance` conservando la data existente (snapshot y actualizaciones previas).

#### Scenario: Progreso y autoevaluación habilitados en avance
- **WHEN** el empleado abre su evaluación en fase `avance`
- **THEN** puede registrar progreso en sus metas y completar su autoevaluación

#### Scenario: Comentarios de jefe y RH visibles en avance
- **GIVEN** existen comentarios del jefe o de RH sobre la evaluación
- **WHEN** el empleado consulta su evaluación en fase `avance`
- **THEN** los comentarios del jefe y de RH son visibles

#### Scenario: Auto-creación de Evaluation al abrir cierre
- **GIVEN** no existe Evaluation para el empleado en el ciclo de cierre
- **WHEN** se abre la vista de cierre
- **THEN** el sistema crea la Evaluation automáticamente y la muestra sin error 404

#### Scenario: Revert de cierre a avance conserva data
- **GIVEN** una Evaluation en fase `cierre` con snapshot y data registrada
- **WHEN** RH revierte la fase a `avance`
- **THEN** la data existente se conserva y la evaluación queda editable en `avance`

### Requirement: Competency write allowed in avance/medio-anio
La escritura de competencias (self y manager/RH) SHALL estar permitida en fases `avance`/`medio-anio` y `cierre`; Submit/Finalize SHALL seguir exigiendo `cierre`. Fuera de `avance`/`medio-anio`/`cierre` el sistema SHALL responder 409 PHASE_NOT_ADVANCEABLE.

#### Scenario: Self 1-5 en avance OK
- **WHEN** el empleado actualiza su autoevaluación con rating 1-5 en fase `avance`
- **THEN** el sistema acepta la escritura sin error de fase

#### Scenario: Jefe vía RH en avance OK
- **WHEN** un usuario con permiso de RH actualiza la evaluación RH en fase `avance`
- **THEN** el sistema acepta la escritura sin error de fase

#### Scenario: Submit en avance responde 409
- **WHEN** se intenta Submit (self o RH) o Finalize en fase `avance`
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE (operación solo `cierre`)

    #### Scenario: Fuera de fase escribible responde 409
- **WHEN** se actualiza self o RH fuera de `avance`/`medio-anio`/`cierre` (p. ej. `asignacion`)
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE

### Requirement: RH evaluation write authorized for RH or assigned manager
La escritura de la evaluación RH (`POST`/`PUT /evaluations/{id}/rh-evaluation`) SHALL estar autorizada solo a RH (permiso `eval:rh`) o al jefe asignado (manager del evaluado). Sin sesión el sistema SHALL responder 401; usuario autenticado no autorizado SHALL responder 403.

#### Scenario: RH con permiso eval:rh escribe OK
- **WHEN** un usuario con permiso `eval:rh` actualiza la evaluación RH
- **THEN** el sistema acepta la escritura (sujeto a reglas de fase)

#### Scenario: Jefe asignado escribe OK
- **WHEN** el manager del evaluado actualiza la evaluación RH
- **THEN** el sistema acepta la escritura (sujeto a reglas de fase)

#### Scenario: Usuario no autorizado responde 403
- **WHEN** un usuario autenticado sin permiso `eval:rh` y que no es el jefe asignado intenta escribir la evaluación RH
- **THEN** el sistema responde 403

#### Scenario: Sin sesión responde 401
- **WHEN** se intenta escribir la evaluación RH sin sesión válida
- **THEN** el sistema responde 401
