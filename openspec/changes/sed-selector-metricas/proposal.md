---
title: "SED Selector y Métricas"
status: draft
change: sed-selector-metricas
---

# Proposal: SED Selector y Métricas

## Why

El `GlobalGoalCreateForm` usa un buscador plano que no escala a miles de empleados ni refleja la jerarquía real; RH no puede asignar por equipo/rama. Las métricas RH mezclan fases (avance vs cierre) y no separan equipo + jefe, por lo que los tableros no cuadran con la fase actual del ciclo.

## What

- Reemplazar el buscador por un selector DaisyUI: menú paginado anidado (paged + nested) que navega el árbol org por ramas con paginación.
- Métricas RH alineadas a fase: `metrics_service.go` computa por `phase` separando tabla equipo + fila jefe, consistente con `cycle.current_phase`.

## Impact

- `web/src/lib/components/goals/GlobalGoalCreateForm.svelte`: nuevo selector (sin buscador plano), orden por `ltree path` de `orgtree`.
- `api/internal/service/.../metrics_service.go`: queries filtradas por `phase`; contrato OpenAPI de métricas con `phase` requerido/opcional con default.
- Sin cambios en RBAC (reusa scope team/all) ni en matriz 9×9.
