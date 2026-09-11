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
