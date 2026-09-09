# Design: SED Selector y Métricas

## Context

`GlobalGoalCreateForm.svelte` hoy filtra en memoria; `orgtree` expone `ltree path` apto para orden jerárquico. `metrics_service.go` agrega sin discriminar `phase`, mezclando avance/cierre.

## Decisions

### D1. Selector DaisyUI paged nested

- Reemplazar input search por menú DaisyUI anidado: nivel = hijos del nodo actual, paginado (cursor, ej. 50/página), breadcrumb de rama.
- Orden server-side por `orgtree path` (ltree); búsqueda opcional solo dentro de rama con `ILIKE` + límite.
- Emite selección múltiple de empleados/nodos; payload sin cambios de forma (ids).

### D2. Orden ltree path orgtree

- Listado jerárquico `ORDER BY path` para que cada rama quede contigua; paginación por cursor sobre ese orden.

### D3. `metrics_service.go` alineado por phase

- Toda métrica RH recibe `phase` (default `cycle.current_phase`); agrega tabla equipo (descendientes) + fila jefe por separado.
- Queries con `WHERE cycle_id AND phase`; índice existente `(cycle_id, phase)` reutilizado (ver change 9box-fases si lo crea).
