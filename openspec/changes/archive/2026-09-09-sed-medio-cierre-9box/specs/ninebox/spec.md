## ADDED Requirements

### Requirement: Matriz 9×9 separada por fase

`ComputeMatrixView` y `RecomputeMatrix` SHALL exigir `phase_id` obligatorio; la vista 9×9 de medio-año (`avance`) y la de `cierre` SHALL computarse y cachearse por separado sin mezclarse.

#### Scenario: Cómputo de medio-año

- **WHEN** se llama `ComputeMatrixView` con `phase_id=avance`
- **THEN** retorna la matriz solo con evaluaciones de `avance`

#### Scenario: Recompute sin phase

- **WHEN** se llama `RecomputeMatrix` sin `phase_id`
- **THEN** retorna 400
