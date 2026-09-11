## Why

La ruta mi-evaluación y el flujo avance/cierre tienen fricciones verificadas: vacío sin salida a metas, avatar con iniciales en blanco, breadcrumb y columna RH confusos, cierre con 404 sin auto-creación, y brecha sin ponderado global. Se corrigen ahora para cerrar el ciclo medio/cierre sin trabajo manual.

## What Changes

- Mi-evaluación vacía: EmptyState con gate `assignmentStatus` + botón "Ir a metas" (sin redirect auto).
- Avatar: conserva iniciales; fix nombre correcto visible (hoy blanco sin inicial); sin migración S3.
- Breadcrumb 9x9 por rol: Mi evaluación > Rol (self), Evaluaciones > Rol (rh), Mis evaluados > Rol (jefe/director); columna RH renombrada a "Evaluación", oculta si `employeeId === session.user.employeeId`; grid con sized cols uniformes.
- Labels por fase: "Evaluación de avance de medio año / Guardar avance" vs "Evaluación de cierre de año / Guardar cierre".
- Cierre 404: auto-crear Evaluation por meta (nadie crea manual); comentarios con botón enviar.
- Revert cierre→avance permitido aunque haya snapshot/data (agregar data o updates fase anterior).
- Avance: progreso + autoeval habilitados, comentarios jefe/RH visibles.
- Textarea 1 línea con resize visible; toasts con `error.code` + éxito; brecha self vs promedio ponderado vs nivel esperado (revisar peso auto/rh en entries 9x9).

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `evaluation-lifecycle`: empty con gate, labels por fase, auto-crear Evaluation en cierre, revert cierre→avance, toasts y textarea.
- `ninebox`: breadcrumb por rol, columna Evaluación con ocultamiento self, grid sized cols, brecha ponderada global backend.
- `goals-and-weighting`: avatar fix sin S3, progreso+autoeval y comentarios en avance, status pendiente/completado distinto avance/cierre.

## Impact

- Frontend: `web/src/routes/mi-evaluacion/+page.svelte`, `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte`, `GoalClosureCard`, avatar/iniciales, 9x9 grid, textarea, toasts.
- Backend: `api/internal/service/evaluation/` (auto-create, revert, brecha ponderada), `ninebox_service.go` (peso auto/rh), `rh/ciclos` revert con refresh labels.
- Contratos: `api/openapi/evaluations-and-9x9.yaml`, `goals-api.yaml` (labels fase, error.code).

## Non-goals

- Migración de avatares a S3; rediseño completo 9x9; cambios de RBAC o esquema de pesos.
