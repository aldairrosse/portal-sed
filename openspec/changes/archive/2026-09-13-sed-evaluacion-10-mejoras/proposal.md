## Why

La ruta mi-evaluación y el flujo avance/cierre tienen fricciones verificadas: vacío sin salida a metas, cierre con 404 sin auto-creación, y mezcla de data entre fases (una sola fila acumula competencias de avance y cierre). Se corrigen ahora para cerrar el ciclo medio/cierre sin trabajo manual y sin pisar data entre fases.

Fuente de verdad: el código del repo (EvaluationDetail y permisos de jefa para ver/escribir comentarios de metas/competencias) prevalece sobre cualquier descripción previa.

## What Changes

- Evaluación por fase con fila propia: `Ensure`/`FindBy(employee,cycle,phase)`; avance escribe avance, cierre escribe cierre; nunca se pisan.
- Avances/ratings + comentarios por rol y fase separados e independientes; editables en fase actual incluso tras revert cierre→avance.
- Cierre 404: auto-crear Evaluation de fase `cierre` si no existe (idempotente, concurrente OK).
- Revert cierre→avance conserva ambas fases (no borra snapshot ni data de ninguna fase); la fase actual queda editable.
- Gates por fase: escritura self/manager-RH en `avance`/`medio-anio`/`cierre`; Submit/Finalize solo `cierre` → `409 PHASE_NOT_ADVANCEABLE` fuera de fase.
- Mi-evaluación vacía: EmptyState con gate `assignmentStatus` + botón "Ir a metas" (sin redirect auto).
- Labels por fase: "Evaluación de avance de medio año / Guardar avance" vs "Evaluación de cierre de año / Guardar cierre".
- Avance: progreso + autoeval habilitados, comentarios jefe/RH visibles.
- Avatar: conserva iniciales; fix nombre correcto visible (hoy blanco sin inicial); sin migración S3.
- Textarea 1 línea con resize visible; toasts con `error.code` + éxito; `ErrorState` con dueño único en este change.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `evaluation-lifecycle`: fila propia por (employee,cycle,phase), auto-crear en cierre, revert conserva ambas fases, gates por fase, comentarios por rol/fase independientes y editables tras revert.
- `goals-and-weighting`: empty con gate, labels por fase, progreso+autoeval y comentarios en avance, avatar fix sin S3, textarea, toasts y `ErrorState` dueño único.

## Impact

- Frontend (verdad): `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte`, `web/src/routes/mi-evaluacion/+page.svelte`.
- Backend: `api/internal/repository/evaluation/evaluation_repo.go` (Ensure/FindBy por fase), `api/internal/service/evaluation/` (gates + revert).
- Contratos: `api/openapi/evaluations-and-9x9.yaml`, `goals-api.yaml` (labels fase, error.code) — solo documentar.

## Non-goals

- Backfill por fase: no se ejecuta aquí; este change no lo contradice (filas por fase lo permiten a futuro).
- Quitar `finished_at` en cierre: solo referencia, no se implementa aquí.
- Migración de avatares a S3; rediseño 9x9 (spec ninebox omitida a propósito); cambios de RBAC o esquema de pesos.
