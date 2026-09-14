## Why

`EmployeeEvaluationDetail` deriva `phaseKind` (`avance`/`cierre`), `GoalClosureCard` es single card por fase, y los permisos de jefa ya permiten ver/escribir comentarios de metas/competencias en código (código = verdad). Las specs estaban desalineadas: trataban `cierre` como ciclo finalizado (`finished_at=NOW()` en `UpdatePhase`) y no documentaban separación por fase ni permisos por rol.

## What Changes

- Q2 — Separación total por fase: `avance` vs `cierre` independientes; comentarios separados por fase y rol; solo la fase actual es editable, la previa es inmutable.
- Q4 CRÍTICO — `UpdatePhase` NO setea `finished_at` al entrar a `cierre`; `finished_at` solo se setea al crear el nuevo ciclo anual en `asignacion`. `cierre` queda activo hasta nuevo ciclo. `GetActiveCycleID WHERE finished_at IS NULL` queda válido sin parche `OR current_phase='cierre'`.
- Q4 — `cierre` no finaliza hasta nuevo ciclo; botón de cierre manual fuera de scope.
- Q3 — Backfill con fase no aplica directo aquí; no contradecir (migración `000046` idempotente `medio-anio` → `avance`).
- Permisos jefa documentados en spec (qué rol ve/escribe qué comentario en qué fase, según código actual).
- `CanUpdateProgress` (`avance` + alias `medio-anio` + `cierre`) y `ResolvePhaseID` (solo `avance`/`cierre`, resto `400`) se mantienen.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `goals-and-weighting`: avances permitidos en `avance` y `cierre`; comentarios de metas separados por fase/rol; `cierre` activo hasta nuevo ciclo (sin `finished_at`).
- `ninebox`: snapshot 9-box única por ciclo/fase sin mezcla; `cierre` activo; `ResolvePhaseID` solo `avance`/`cierre` → resto `400`.

## Impact

- Specs/plan solo (este change): 5 archivos listados en tasks.
- Implementación futura (fuera de este change): `UpdatePhase`/`TransitionPhase` (quitar `CASE cierre THEN NOW()`), `cycle_repo.go` sin cambios.
- Contratos: `goals-api.yaml`, `evaluations-and-9x9.yaml` ya documentan `NineBoxPhase`; no se tocan aquí.
