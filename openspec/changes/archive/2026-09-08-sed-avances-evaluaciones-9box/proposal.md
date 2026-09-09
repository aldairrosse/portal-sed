---
title: "SED Avances, Evaluaciones y 9-Box por Fase"
status: draft
change: sed-avances-evaluaciones-9box
---

# Proposal: SED Avances, Evaluaciones y 9-Box por Fase

## Why

Hoy el ciclo persiste solo {asignacion,avance,cierre} y la matriz 9×9 colisiona al guardar la segunda fase (índice sin `phase_id`). `medio-anio` es alias solo-lectura de `avance` (nunca persistido) vía `IsMidYearPhase/SamePhaseForWrite`: sin él no hay doble lectura 9-box (avance/medio-año vs cierre), ni gate que bloquee edición fuera de fase, ni revert de cierre→avance, ni CSV de medio-año para RH.

## What

- Canónico persistido `phase ∈ {asignacion,avance,cierre}` en `Cycle/Evaluation/NineBoxMatrix` (display `Medio año` para `avance`); `medio-anio` alias solo-lectura equivalente a `avance`, nunca persistido.
- Doble 9-box por cycle: una vista en `avance` (incluye alias `medio-anio` solo-lectura) y otra en `cierre`, con `UNIQUE(employee_id,cycle_id,phase)` e índice `(cycle_id,phase)`.
- Gate `SamePhaseForWrite(request.phase, cycle.current_phase)` para escrituras (409); revert editable (cierre→avance) vía `PhaseService`; CSV habilitado en medio-año filtrado por `?phase=avance` (alias legacy `?phase=medio-anio` solo-lectura).

## Impact

- `api/`: Ent schemas + migración, `PhaseService`, gates en handlers, CSV con filtro `phase`.
- `api/openapi/cycle.yaml` + `evaluations-and-9x9.yaml`: enums canónicos + alias `medio-anio` solo-lectura en listados/CSV, 409/403 fuera de fase.
- `web/`: badges y filtros por fase, doble vista 9-box por ciclo.
- Convención: persistido siempre `avance`; UI/display `Medio año`; `medio-anio` alias solo-lectura (ASCII sin ñ) nunca persistido.
