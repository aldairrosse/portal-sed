# phase-9box Specification

## Purpose
TBD - created by archiving change sed-avances-evaluaciones-9box. Update Purpose after archive.

## Requirements

### Requirement: Fase avance cubre medio-año (3 valores canónicos)

El sistema SHALL persistir solo `phase ∈ {asignacion, avance, cierre}`; `avance` cubre medio-año (display `Medio año`). `medio-anio` es alias solo-lectura equivalente a `avance` en `IsMidYearPhase/SamePhaseForWrite`, nunca persistido. El sistema SHALL permitir una evaluación/matriz por `(employee_id, cycle_id, phase)` con índice `(cycle_id, phase)`.

#### Scenario: Doble 9-box por ciclo

- **WHEN** se guardan matrices para el mismo evaluador en `medio-anio` y `cierre` del mismo ciclo
- **THEN** ambas persisten sin error de clave duplicada

### Requirement: Gate de escritura por fase actual

Toda escritura SHALL exigir `request.phase == cycle.current_phase`; en caso contrario SHALL retornar `409`.

#### Scenario: Escritura fuera de fase

- **WHEN** se envía una evaluación con `phase=avance` mientras `current_phase=cierre`
- **THEN** el sistema retorna 409 y no persiste

### Requirement: Revert editable cierre a avance

`PhaseService` SHALL permitir a RH revertir `cierre→avance` de forma auditada (quién/cuándo), dejando el ciclo editable en avance.

#### Scenario: RH revierte cierre

- **WHEN** RH invoca revert en un ciclo en `cierre`
- **THEN** `current_phase` vuelve a `avance` y queda registro de auditoría

### Requirement: CSV filtrado por fase en medio-año

El export CSV SHALL aceptar `?phase=avance` (y alias legacy `?phase=medio-anio` solo-lectura) y SHALL exportar solo filas de esa fase.

#### Scenario: Export de medio-año

- **WHEN** RH exporta CSV con `phase=avance` (o alias `phase=medio-anio`)
- **THEN** el archivo contiene solo evaluaciones de avance/medio-año
