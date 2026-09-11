# Tasks: sed-cierre-avances-9box

> Volcado desde `status --change sed-cierre-avances-9box --json` + `show --type change sed-cierre-avances-9box --diff`. Trabajo ya implementado en el árbol (4 iteraciones); estas tareas lo trazan para validación y eventual archive.

## Backend

- [x] 1.1 `api/internal/service/goal/phase_check.go` — `CanUpdateProgress` permite `PhaseAvance`, `PhaseMedioAnio`, `PhaseCierre`
- [x] 1.2 `api/internal/service/evaluation/ninebox_service.go` — guard `isNineBoxPhase` (avance/cierre + aliases) en `ResolvePhaseID`, `400` en resto
- [x] 1.3 Migración `000046_ninebox_phase_check` (up/down)
  - Nota R3: "DOWN 000046 es NOOP intencional — data no productiva, nine-entries recalculados por consulta, seguro borrar"

## Contratos

- [x] 2.1 `api/openapi/goals-api.yaml` — `NineBoxPhase` (`avance`/`cierre`) en avances
- [x] 2.2 `api/openapi/evaluations-and-9x9.yaml` — `NineBoxPhase` en 9-box

## Frontend

- [x] 3.1 Gates de avance con `isMedioAnio || isFinAnio` en `EmployeeEvaluationDetail.svelte`, `mis-evaluados/+page.svelte`, `rh/evaluaciones/+page.svelte`

## Cierre

- [ ] 4.1 `validate --all --strict --json` en verde
- [ ] 4.2 `archive sed-cierre-avances-9box -y` tras confirmar (trabajo ya en árbol)
