## Purpose

Define los snapshots de metas por fase del ciclo, su sincronización con el valor actual y los comentarios con autor y fecha para metas y competencias.

## ADDED Requirements

### Requirement: Snapshot por fase con valor directo según unit

El sistema SHALL persistir en `evaluation_goals` los campos `avance_progress` y `cierre_progress` (FLOAT NULL) con el valor directo de la meta según su `unit` (porcentaje/moneda/numero/binario) capturado en cada fase.

#### Scenario: Captura de snapshot en avance

- **WHEN** se registra el avance de una meta con `unit=porcentaje` y valor `70` en fase `avance`
- **THEN** `avance_progress` queda en `70` y `cierre_progress` permanece NULL

#### Scenario: Captura de snapshot en cierre

- **WHEN** se registra el cierre de una meta con `unit=moneda` y valor `500000` en fase `cierre`
- **THEN** `cierre_progress` queda en `500000` y `avance_progress` conserva su valor previo

### Requirement: Sincronización de current_value con snapshots

El sistema SHALL resolver `goals.current_value = cierre_progress ?? avance_progress`, y cuando `UpdateGoalProgress` actualice `current_value` SHALL escribir también el snapshot de la fase activa (`cycles.current_phase`: avance o cierre).

#### Scenario: Lectura prioriza cierre sobre avance

- **WHEN** una meta tiene `avance_progress=60` y `cierre_progress=80`
- **THEN** `current_value` es `80`

#### Scenario: Escritura actualiza snapshot de fase activa

- **WHEN** `UpdateGoalProgress` actualiza `current_value=75` con `cycles.current_phase=avance`
- **THEN** `avance_progress` queda en `75` y `cierre_progress` no cambia

### Requirement: Cierre con valor directo sin derivación

`PUT goal-state` (cierre) SHALL persistir el valor directo en el snapshot de la fase activa y sincronizar `current_value`; el sistema SHALL NOT derivar `final_rating = int(finalProgress*5)` y la UI SHALL mostrar el valor según `unit`.

#### Scenario: Cierre persiste valor directo

- **WHEN** se llama `PUT goal-state` en cierre con valor `90` y `unit=porcentaje`
- **THEN** el snapshot de cierre queda en `90`, `current_value` queda en `90`
- **AND** no se calcula ni persiste `final_rating` derivado

### Requirement: Comentarios de metas y competencias con autor y fecha

El sistema SHALL guardar y mostrar los comentarios de metas y competencias (self/jefe/rh) con nombre de autor, fecha y hora, tomando el autor desde la sesión activa.

#### Scenario: Comentario muestra autor y fecha

- **WHEN** un jefe con sesión `jefe-ana` crea un comentario en una meta
- **THEN** el comentario persiste con autor `jefe-ana`, fecha y hora
- **AND** la lectura lo devuelve con nombre de autor, fecha y hora visibles
