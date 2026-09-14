# Tasks: sed-cierre-avances-9box

> Plan-only (specs). 5 archivos. T4 = quitar `CASE cierre THEN NOW()`; T8 = verificación. Sin código app en este change.

## Specs/plan (este change, ya aplicado)

- [x] T1. `proposal.md` → Q2/Q4/permisos documentados; cierre activo hasta nuevo ciclo; botón cierre fuera scope.
- [x] T2. `design.md` → Decisión 1: `UpdatePhase` no setea `finished_at`; `GetActiveCycleID` válido sin cambios; Decisión 2/3: separación por fase + permisos jefa.
- [x] T3. `specs/goals-and-weighting/spec.md` → separación total, comentarios por fase/rol, permisos jefa, cierre activo sin `finished_at`.
- [x] T4. `specs/ninebox/spec.md` → snapshot por fase + `ResolvePhaseID` solo `avance`/`cierre` + cierre activo + permisos competencias.
  - NOTA impl futura (no en este change): quitar `CASE cierre THEN NOW()` en `UpdatePhase`/`TransitionPhase` (`cycle_repo.go`); NO aplicar `OR current_phase='cierre'` en `GetActiveCycleID`.
- [x] T5. `tasks.md` (este archivo) → archivo→resultado claro; T8 verificación definida.

## Verificación futura (no implementar aquí)

- [x] T6. `api/internal/repository/cycle/cycle_repo.go:277` → Quitar `CASE cierre THEN NOW()` en `UpdatePhase`/`TransitionPhase`
  - Resultado: entrar a `cierre` deja `finished_at IS NULL` (cierre activo); `finished_at` solo al crear nuevo ciclo en `asignacion`.
- [x] T8. Verificación end-to-end (requiere impl fuera de este change).
  - `UpdatePhase` a `cierre` → `finished_at IS NULL` (cierre activo).
  - `GetActiveCycleID` retorna ciclo en `cierre`; avance + autoeval + 9-box `cierre` operan sin tocar `avance`.
  - Comentarios fase actual editables, previa inmutable, según matriz jefa/empleado.
  - Backfill `medio-anio`→`avance` idempotente (Q3 no contradicho).
