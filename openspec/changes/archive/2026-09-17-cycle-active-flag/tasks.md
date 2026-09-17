# Tasks: cycle-active-flag

Orden: a→b→c→d→e→f. Sin commits sin pedido explícito.

- [ ] (a) schema + migración ciclo activo
  Files: `api/internal/schema/cycle.go`, migración versionada (`is_active boolean NOT NULL DEFAULT false` + unique partial index `(organization_id) WHERE is_active = true`)
  Acept: migrar arriba/abajo en limpio; segundo activo por org falla por índice. Dep: design. Primero.

- [ ] (b) repo + service activate transaccional
  Files: `api/internal/repository/cycle/cycle_repo.go` (`GetCurrent`, `GetActiveCycleID`), `api/internal/service/cycle/cycle_service.go` (`Activate` tx + `SELECT ... FOR UPDATE`, apaga resto de la org)
  Acept: activar viejo deja resto en `false`; doble activación concurrente → un solo activo. Dep: a.

- [ ] (c) handler + contrato OpenAPI
  Files: handler de ciclos (`POST /cycles/{id}/activate` solo RH, `GET /cycles/current`), guard 409 `cycle-not-active` en escrituras de metas/competencias/evaluaciones, `api/openapi/cycle.yaml`
  Acept: RH activa → 200 con `is_active=true`; no-RH → 403; escritura en cerrado → 409 sin persistir. Dep: b.

- [ ] (d) cycleStore + gestión con badges
  Files: `web/src/lib/stores/cycleStore.svelte.ts` (`loadCurrent`), vista gestión de ciclos (badge Activo/Cerrado, acción activar solo RH)
  Acept: gestión muestra Activo/Cerrado; sin activo → vacío con mensaje y CTA. Dep: c.

- [ ] (e) guards en stores + 9-box tabs
  Files: stores goals/competency/evaluation (validan ciclo activo, consumen `loadCurrent`), vista 9-box (solo activo, tabs avance/cierre)
  Acept: sin activo no se emite POST; 409 backend → error legible; 9-box no consulta cerrados. Dep: d.

- [ ] (f) tests + validate estricto
  Files: `*_test.go` repo/service/handler (tabla), tests stores/tabs
  Acept: `rtk proxy openspec validate --all --strict --json` verde; cobertura del invariante un-activo-por-org. Dep: a-e. Cierra change.
