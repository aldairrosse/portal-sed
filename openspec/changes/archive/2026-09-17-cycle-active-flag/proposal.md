# Proposal: cycle-active-flag

## Why
Varios ciclos conviven sin marca de vigencia y las escrituras de metas, competencias, evaluaciones y 9-box pueden caer en ciclos ya cerrados. Se introduce el flag `is_active` estricto: un solo ciclo activo por organización, y todo el sistema escribe y visualiza sobre ese ciclo.

## What Changes
- Flag `is_active boolean` en `Cycle` (default `false`); unique partial index por organización donde `is_active = true`.
- Activar un ciclo pone en `false` a los demás de la organización (transacción + lock); endpoint `POST /cycles/{id}/activate` (solo RH).
- Metas, competencias y evaluaciones escriben solo en el ciclo activo: guard en backend (error 409) + validación en stores frontend.
- Gestión de ciclos muestra badge Activo/Cerrado; `cycleStore.loadCurrent()` expone el ciclo activo y lo reutilizan los stores de metas/competencias/evaluación.
- 9-box opera solo sobre el ciclo activo, con tabs avance/cierre.

## Capabilities
- **New Capabilities**: `cycle-activation` (endpoint de activación, `cycleStore.loadCurrent`, badges Activo/Cerrado, guards en stores, 9-box sobre ciclo activo con tabs avance/cierre).
- **Modified Capabilities**: `evaluation-lifecycle` (flag `is_active` en el modelo Cycle, invariante un-activo-por-organización, guard de escritura solo en ciclo activo).

## Impact
`api/internal/schema/cycle.go`, `api/internal/repository/cycle/cycle_repo.go` (`GetCurrent`/`GetActiveCycleID`), `api/internal/service/cycle/cycle_service.go` (activate en tx + lock), handler de ciclos + `POST /cycles/{id}/activate`, `api/openapi/cycle.yaml`, `web/src/lib/stores/cycleStore.svelte.ts` (`loadCurrent`), vista de gestión de ciclos (badges), stores de goals/competency/evaluation (validación ciclo activo), vista 9-box (tabs avance/cierre).
