# Design: sed-cierre-avances-9box

## Contexto

Verdad de código (no reinvestigar): `EmployeeEvaluationDetail` deriva `phaseKind` `avance`/`cierre`; `GoalClosureCard` es single card; permisos de jefa ya permiten ver/escribir comentarios de metas/competencias. Este change solo alinea specs a código.

## Decisión 1 (Q4 CRÍTICO): `cierre` no finaliza el ciclo

- `UpdatePhase`/`TransitionPhase` NO debe setear `finished_at=NOW()` al entrar a `cierre`.
- `finished_at` se setea únicamente al crear el nuevo ciclo anual en `asignacion` (cierre del ciclo anterior = apertura del siguiente).
- `GetActiveCycleID WHERE finished_at IS NULL` queda válido sin cambios; NO aplicar parche `OR current_phase='cierre'`.
- `cierre` queda activo hasta nuevo ciclo; botón de cierre manual fuera de scope.

## Decisión 2 (Q2): separación total por fase

- `avance` y `cierre` son evaluaciones independientes: progreso, snapshot 9-box y comentarios no se mezclan.
- Comentarios (metas y competencias) separados por fase y rol; solo la fase actual es editable; la fase previa es inmutable (solo lectura).
- `ResolvePhaseID` acepta solo `avance`/`cierre` (+aliases `medio-anio`/`medio_anio` → `avance`); resto `400`. `ComputeMatrixView`/`RecomputeMatrix` exigen `phase_id`.
- Q3: backfill con fase no aplica aquí; no contradecir (migración `000046` idempotente).

## Decisión 3: permisos jefa (código = verdad)

- Jefa ve comentarios de metas/competencias de sus evaluados en `avance` y `cierre`; escribe comentarios de jefa solo en fase actual.
- Empleado ve/escribe sus comentarios solo en fase actual; fase previa inmutable para ambos roles.
- Detalle rol × fase en `specs/goals-and-weighting/spec.md` y `specs/ninebox/spec.md`.

## No-goals

- Sin código app ni tests en este change; sin botón de cierre; sin cambios de ponderación (dueño `sed-snapshots-metas-9box`); sin `ErrorState`.
