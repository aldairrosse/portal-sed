# evaluation-lifecycle — delta

## ADDED Requirements

### Requirement: Ciclo activo único por organización
El sistema SHALL mantener exactamente un ciclo con `is_active = true` por organización (o ninguno antes de la primera activación). Activar un ciclo SHALL poner `is_active = false` en todos los demás ciclos de la misma organización dentro de la misma transacción, bajo lock pesimista de la fila del ciclo a activar. La unicidad SHALL estar respaldada por un unique partial index `(organization_id) WHERE is_active = true`.

#### Scenario: activar un ciclo viejo desactiva el actual
- GIVEN organización con ciclo 2025 activo y ciclo 2024 cerrado
- WHEN RH activa el ciclo 2024
- THEN el ciclo 2024 queda con `is_active = true`
- AND el ciclo 2025 queda con `is_active = false`

#### Scenario: activación concurrente deja un solo ganador
- GIVEN dos peticiones simultáneas de activación sobre ciclos distintos de la misma organización
- WHEN ambas se procesan
- THEN exactamente un ciclo termina con `is_active = true`
- AND la perdedora recibe error o converge al mismo estado sin duplicar activos

#### Scenario: escrituras fuera del ciclo activo son rechazadas
- GIVEN un ciclo cerrado (`is_active = false`)
- WHEN se intenta crear o modificar meta, competencia asignada o evaluación sobre ese ciclo
- THEN el backend rechaza con error 409 y código `cycle-not-active`
- AND no se persiste ningún cambio
