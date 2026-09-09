## ADDED Requirements

### Requirement: Competencias por fase con auto separado

El sistema SHALL permitir autoevaluación de competencias en `avance` y `cierre` por separado: el unique de `EvaluationCompetency` SHALL pasar de `(evaluation_id,competency_id)` a `(evaluation_id,competency_id,source[,phase])`, o las evaluaciones SHALL separarse por phase vía el unique existente `(employee_id,cycle_id,phase)` en `Evaluation`. Solo `source=auto` se duplica por fase.

#### Scenario: Auto en avance y cierre

- **WHEN** un colaborador guarda auto en `avance` y luego en `cierre`
- **THEN** ambos registros coexisten sin conflicto de unique

#### Scenario: Heteruevaluación única por fase

- **WHEN** el jefe guarda la evaluación de competencias de `cierre`
- **THEN** no altera la de `avance`
