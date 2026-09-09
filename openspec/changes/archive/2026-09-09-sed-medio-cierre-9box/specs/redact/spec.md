## ADDED Requirements

### Requirement: Redacción de detalle por viewer y fase

`RedactDetailForSelf` SHALL redactar el detalle de evaluación según `viewerMode` + `phase`: un viewer `colaborador` en modo `self` SHALL ver solo comentarios (sin puntajes/tiers de otros), y el acceso a detalle ajeno en `avance` SHALL retornar 403.

#### Scenario: Colaborador ve su avance

- **WHEN** un colaborador abre su propio detalle en `phase=avance`
- **THEN** ve comentarios completos con puntajes propios redactados según regla `self`

#### Scenario: Colaborador abre detalle ajeno

- **WHEN** un colaborador llama `GET /evaluations/{idAjeno}?phase=avance`
- **THEN** recibe 403
