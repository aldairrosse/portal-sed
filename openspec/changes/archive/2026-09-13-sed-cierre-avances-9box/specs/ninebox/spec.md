# ninebox — Delta spec (sed-cierre-avances-9box)

## MODIFIED Requirements

### Requirement: Matriz 9×9 separada por fase

`ComputeMatrixView` y `RecomputeMatrix` SHALL exigir `phase_id` obligatorio; la vista 9×9 de medio-año (`avance`) y la de `cierre` SHALL computarse y cachearse por separado sin mezclarse. `ResolvePhaseID` SHALL aceptar solo fases `avance` y `cierre` (`medio-anio`/`medio_anio` como alias solo-lectura de `avance`); cualquier otra fase SHALL retornar `400` y no computar matriz.

#### Scenario: Cómputo de medio-año

- **WHEN** se llama `ComputeMatrixView` con `phase_id=avance`
- **THEN** retorna la matriz solo con evaluaciones de `avance`

#### Scenario: Recompute sin phase

- **WHEN** se llama `RecomputeMatrix` sin `phase_id`
- **THEN** retorna 400

#### Scenario: Fase válida cierre

- **WHEN** se llama `ResolvePhaseID` con `phase=cierre`
- **THEN** resuelve el `phase_id` de `cierre` y computa su snapshot sin tocar la de `avance`

#### Scenario: Fase inválida rechazada

- **WHEN** se llama `ResolvePhaseID` con `phase=asignacion`
- **THEN** retorna `400` (`phase must be one of 'avance', 'medio-anio', 'cierre'`) y no persiste nada

## ADDED Requirements

### Requirement: Cierre activo hasta nuevo ciclo

El ciclo en `cierre` SHALL permanecer activo (`finished_at IS NULL`) hasta la creación del nuevo ciclo anual en `asignacion`. `GetActiveCycleID WHERE finished_at IS NULL` SHALL resolver el ciclo en `cierre` sin parche adicional. `UpdatePhase` SHALL NOT setear `finished_at` en `cierre`.

#### Scenario: 9-box en cierre activo

- GIVEN ciclo en `cierre` con `finished_at IS NULL`
- WHEN se computa matriz de `cierre`
- THEN resuelve ciclo y snapshot de `cierre` sin tocar `avance`.

### Requirement: Comentarios de competencias y permisos jefa

Los comentarios de competencias SHALL separarse por fase y rol igual que metas: solo fase actual editable, previa inmutable. Jefa SHALL ver competencias/comentarios de evaluados en ambas fases y escribir solo en fase actual (código actual = verdad).

#### Scenario: Permisos competencias

- GIVEN rol `jefa`, ciclo en `cierre`
- WHEN accede a competencias
- THEN ve `avance` inmutable y escribe comentarios de `cierre`.
