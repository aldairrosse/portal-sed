# Design: SED Avances, Evaluaciones y 9-Box por Fase

## Context

`nine_box_matrices` tiene unique `(cycle_id, evaluator_id)` sin `phase_id` → la segunda matriz por fase falla con duplicate key (mismo hallazgo que `ninebox-computed-matrix` D4). `Cycle.current_phase` solo rota avance/cierre.

## Decisions

### D1. Schemas Ent: phase canónica (3 valores)

- `Cycle` / `Evaluation` / `NineBoxMatrix`: persistido `phase ∈ {asignacion, avance, cierre}`, default `avance`. `medio-anio` es alias solo-lectura equivalente a `avance` vía `IsMidYearPhase/SamePhaseForWrite`, nunca persistido.
- `UNIQUE(employee_id, cycle_id, phase)` donde aplique (evaluaciones) + índice `(cycle_id, phase)` para listados/CSV.

### D2. Migración

- Goose: `ADD COLUMN phase`, backfill (`avance` si `current_phase` nulo/legacy), crear unique + índice; idempotente con `IF NOT EXISTS`.

### D3. PhaseService + gates

- `PhaseService.revert(cierre→avance)`: transición editable auditada (quién/cuándo), solo RH.
- Gates de escritura: si `request.phase != cycle.current_phase` → `409 Conflict` (o 403 según auth); lecturas filtran por `phase` con default `current_phase`.

### D4. Contratos `cycle.yaml` / `evaluations-and-9x9.yaml`

- Enum `phase` canónico `{asignacion, avance, cierre}` + alias solo-lectura `medio-anio≡avance` en listados/CSV; documentar gate, revert y doble matriz por ciclo.

### D5. CSV por phase + convención ñ

- Export CSV acepta `?phase=avance` (alias legacy `?phase=medio-anio` solo-lectura); persistido siempre `avance`, presentación `Medio año`. Mapping único en helper compartido (`state`/cycle.ts).
