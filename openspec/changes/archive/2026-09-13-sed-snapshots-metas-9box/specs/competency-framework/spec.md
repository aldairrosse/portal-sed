## ADDED Requirements

### Requirement: Promedio ponderado de competencias ignorando ausentes

El sistema SHALL calcular el puntaje agregado de competencias como promedio ponderado `0.8 RH / 0.2 self`, ignorando las calificaciones ausentes en vez de tratarlas como 0.

#### Scenario: Ponderado RH/self con ambos presentes

- **WHEN** una competencia tiene RH `3` y self `5`
- **THEN** el agregado es `3*0.8 + 5*0.2 = 3.4`

#### Scenario: Solo RH presente

- **WHEN** una competencia solo tiene calificación RH `4` (self ausente)
- **THEN** el agregado es `4`
- **AND** la ausencia de self no aporta 0 al cálculo
