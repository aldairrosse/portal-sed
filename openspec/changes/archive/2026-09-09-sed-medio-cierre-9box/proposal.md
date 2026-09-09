---
title: "SED Medio Cierre 9Box"
status: draft
change: sed-medio-cierre-9box
---

# Proposal: SED Medio Cierre 9Box

## Why

El ciclo SED necesita distinguir medio-año (avance) de cierre: hoy los comentarios de objetivos no están faseados (`goal_comments` sin columna `phase`), las competencias se guardan una sola vez por evaluación (unique `evaluation_id,competency_id` impide auto vs cierre separados), el detalle para `self` no redacta por fase, y la matriz 9×9 se computa sin `phase_id` mezclando medio-año con cierre.

## What

- Comentarios de objetivos faseados por `asignacion/avance/cierre` con API GET/POST por phase, autor y fecha.
- Competencias por fase: auto en avance + cierre separados (unique incluye `source` y fase o evaluaciones separadas por phase).
- Redacción de detalle para `self` según `viewerMode` + phase; colaborador solo ve comentarios.
- Matriz 9×9 separada: medio-año (avance) vs cierre, con `phase_id` obligatorio en cómputo/recompute.

## Out of scope

- Renombrar el enum canónico backend `asignacion/avance/cierre` (definido en `api/internal/schema/evaluation.go:30-37`, `phasedefinition.go:120-124`, `machine.go:35-98`). `medio-anio`/`medio_anio` queda solo como alias de lectura en frontend (`web/src/lib/types/cycle.ts`, `cycle.svelte.ts:2,16`).
- Cambios de cómputo de tiers o de sync Mobonet.

## Impact

- `api/`: migración `goal_comments.phase`, unique `EvaluationCompetency`, gates RBAC por rol/fase, `ComputeMatrixView`/`RecomputeMatrix` con `phase_id`.
- `api/openapi/`: contratos 3.1 primero para comments, competency, redact, ninebox.
- `web/`: gates por fase, alias `medio-anio` solo lectura, vistas 9×9 separadas.
