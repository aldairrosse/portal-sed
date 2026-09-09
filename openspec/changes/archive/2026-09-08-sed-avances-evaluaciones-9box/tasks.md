# Tasks: SED Avances, Evaluaciones y 9-Box por Fase

- [x] Migración: `phase` en Cycle/Evaluation/NineBoxMatrix + `UNIQUE(employee_id,cycle_id,phase)` + índice `(cycle_id,phase)` + backfill
- [x] Ent schemas + `go generate ./...` + consistencia con goose
- [x] Gate `phase == current_phase` en escrituras (409/403) + tests
- [x] `PhaseService`: revert cierre→avance auditable (solo RH) + tests
- [x] Doble 9-box por cycle (medio-anio + cierre) sin duplicate key + tests
- [x] CSV con filtro `?phase=` habilitado en medio-año + tests
- [x] OpenAPI `cycle.yaml` + `evaluations-and-9x9.yaml`: enum `medio-anio`, gates, revert
- [x] Validación: convención `medio-anio` interno / `medio año` display; `openspec validate --all`; `go test ./...`
