## ADDED Requirements

### Requirement: Escalado de metas por direction en 9-box

El sistema SHALL escalar cada meta según su `direction` (ascendente/descendente) a % completado con clamp 0–100 y luego a escala 1–3 (`<34=1, <67=2, else 3`).

#### Scenario: Meta ascendente escala a tier

- **WHEN** una meta ascendente tiene 50% completado
- **THEN** el tier de desempeño es `2`

#### Scenario: Meta descendente invierte el cálculo

- **WHEN** una meta descendente con base `100` y actual `20` se evalúa
- **THEN** el % completado refleja la reducción (clamp 0–100)
- **AND** el tier resultante sigue la escala `<34=1, <67=2, else 3`

### Requirement: RecomputeMatrix con promedio ponderado ignorando ausentes

`RecomputeMatrix` SHALL usar el promedio ponderado `0.8 RH / 0.2 self` (igual que `ComputeMatrixView`) e SHALL ignorar los valores ausentes en vez de tratarlos como 0.

#### Scenario: Ponderado coincide con ComputeMatrixView

- **WHEN** un empleado tiene calificación RH `4` y self `2`
- **THEN** `RecomputeMatrix` calcula `4*0.8 + 2*0.2 = 3.6`
- **AND** el resultado coincide con `ComputeMatrixView` para los mismos insumos

#### Scenario: Valores ausentes se ignoran

- **WHEN** un empleado solo tiene calificación RH `4` (self ausente)
- **THEN** el promedio es `4` (solo RH)
- **AND** no se promedia con 0 por la ausencia de self
