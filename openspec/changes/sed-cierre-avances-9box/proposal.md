## Why

En `cierre` los avances de metas quedaban bloqueados (`CanUpdateProgress` solo permitía `avance`), lo que impedía registrar el progreso final antes del cierre del ciclo. Además, la matriz 9-box aceptaba cualquier fase sin guard, mezclando snapshots entre fases.

## What Changes

- `CanUpdateProgress` permite registrar avances en `avance` (incl. alias `medio-anio`) y `cierre`; sigue bloqueado en `asignacion`.
- `NineBoxService.ResolvePhaseID` rechaza fases fuera de `avance`/`cierre` (con aliases `medio-anio`/`medio_anio`) con `400`; snapshot 9-box única por ciclo/fase, sin mezcla entre `avance` y `cierre`.
- OpenAPI `goals-api.yaml` y `evaluations-and-9x9.yaml`: documentan `NineBoxPhase` (`avance`/`cierre`) en avances y 9-box.
- Frontend (3 archivos): gates de avance visibles con `isMedioAnio || isFinAnio`.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `goals-and-weighting`: avances de meta permitidos en `avance` y `cierre` (antes `cierre` era solo-lectura total).
- `ninebox`: guard de fase — solo `avance`/`cierre` generan snapshot 9-box; otras fases retornan `400`.

## Impact

- Backend: `api/internal/service/goal/phase_check.go`, `api/internal/service/evaluation/ninebox_service.go`, migración `000046` (up/down).
- Contratos: `api/openapi/goals-api.yaml`, `api/openapi/evaluations-and-9x9.yaml`.
- Frontend: `EmployeeEvaluationDetail.svelte`, `mis-evaluados/+page.svelte`, `rh/evaluaciones/+page.svelte`.
