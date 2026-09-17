# Design: cycle-active-flag

## Context
Ver `proposal.md - Why`. El modelo `Cycle` actual (`evaluation-lifecycle`) tiene `id`, `year`, `currentPhase` sin noción de vigencia: nada impide escribir en ciclos cerrados. Este change añade `is_active` estricto con un solo activo por organización y propaga el ciclo activo a backend y frontend.

## Goals / Non-Goals
- Goals: flag `is_active` con unique partial index; activación transaccional que desactiva el resto; `GET /cycles/current` + `POST /cycles/{id}/activate`; `cycleStore.loadCurrent()` reutilizado por stores; badges Activo/Cerrado; 9-box solo activo con tabs avance/cierre.
- Non-Goals: transiciones de fase (siguen lineales según `evaluation-lifecycle`), multi-activo por organización, scheduler de activación automática, cambios de RBAC fuera del guard RH del endpoint.

## Decisions
1. **Schema `is_active boolean NOT NULL DEFAULT false`:** columna en `api/internal/schema/cycle.go` + migración versionada con unique partial index `(organization_id) WHERE is_active = true`. Alternativa check-constraint: descartada (no expresa unicidad condicional en Postgres).
2. **Repo `GetCurrent`/`GetActiveCycleID`:** queries en `cycle_repo.go` filtrando `organization_id + is_active = true`; `GetActiveCycleID` proyección ligera para guards de escritura.
3. **Service `Activate` en tx + `SELECT ... FOR UPDATE`:** `cycle_service.go` bloquea la fila del ciclo a activar, pone `false` a los demás de la org y `true` al objetivo; el índice parcial es la red de seguridad ante concurrencia (una gana, la otra recibe unique-violation → 409).
4. **Handler + rutas:** `POST /cycles/{id}/activate` (RH) y `GET /cycles/current` en el handler de ciclos; error estándar `cycle-not-active` 409 en escrituras fuera del activo; contrato en `api/openapi/cycle.yaml` (OpenAPI 3.1 fuente de verdad).
5. **Frontend `cycleStore.loadCurrent()`:** `web/src/lib/stores/cycleStore.svelte.ts` carga y cachea el activo; stores de goals/competency/evaluation lo consumen y validan antes de POST/PUT; gestión de ciclos renderiza badges; 9-box filtra por activo con tabs avance/cierre (DaisyUI, sentence case, sin box-shadow).

## Risks / Trade-offs
- [Doble activación concurrente] → lock + índice parcial; test de concurrencia.
- [Órgano sin activo tras deploy] → migración deja todo en `false`; gestión muestra vacío con CTA activar (aceptado, sin backfill inventado).
- [Stores con ciclo stale] → `loadCurrent` refresca al montar gestión/9-box; `ponytail:` sin realtime, refetch manual si hiciera falta.

## Migration Plan
1. Merge change → migración añade columna + índice (segura, default `false`). 2. RH activa el ciclo vigente desde gestión. 3. Rollback: revert de código; columna ignorada sin lecturas viejas que la usen.

## Open Questions
- Ninguna bloqueante; mensaje exacto del vacío "sin ciclo activo" se cierra en review UI.
