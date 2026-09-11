## Context

Ver `proposal.md` (Why) y `specs/` (evaluation-lifecycle, goals-and-weighting, ninebox) para motivación y requisitos. Estado actual: fricciones verificadas en mi-evaluación y flujo avance/cierre (vacío sin salida, avatar en blanco, breadcrumb/columna confusos, 404 en cierre, sin ponderado global). Stack: Svelte 5 + DaisyUI en `web/`, Go + Chi en `api/`, contratos OpenAPI 3.1 en `api/openapi/`.

## Goals / Non-Goals

**Goals:**

- Cerrar el ciclo medio/cierre sin trabajo manual: empty con salida, auto-creación en cierre, revert con data conservada.
- Etiquetado y visibilidad por fase/rol decididos en frontend con peso global provisto por backend.

**Non-Goals:**

- Migración de avatares a S3; rediseño completo 9x9; cambios de RBAC o esquema de pesos (ver proposal Non-goals).

## Decisions

- **EmptyState con gate `assignmentStatus` sin redirect auto**: el usuario decide navegar vía "Ir a metas". Alternativa descartada: redirect automático (desorienta y rompe deep-links).
- **Avatar conserva iniciales, fix de nombre visible**: sin migración S3; se corrige la derivación de iniciales del nombre correcto. Alternativa descartada: migrar a S3 (fuera de alcance).
- **Breadcrumb y columna por rol en frontend**: `Mi evaluación > Rol` (self), `Evaluaciones > Rol` (rh), `Mis evaluados > Rol` (jefe/director); columna `Evaluación` oculta si `employeeId === session.user.employeeId`; `sized cols` uniformes. RBAC real sigue en backend. Alternativa descartada: lógica de visibilidad solo en backend (latencia y complejidad innecesarias para presentación).
- **Labels por fase en frontend**: `Evaluación de avance de medio año / Guardar avance` vs `Evaluación de cierre de año / Guardar cierre`, derivados de la fase del ciclo.
- **Auto-crear Evaluation en backend al abrir cierre**: idempotente por (cycle_id, employee_id); el 404 desaparece sin creación manual. Alternativa descartada: crear desde frontend (riesgo de duplicados).
- **Revert cierre→avance conserva snapshot/data**: la fase vuelve a `avance` editable sin borrar nada. Alternativa descartada: revert destructivo (pérdida de auditoría).
- **Brecha ponderada con peso global del backend**: self-view `self vs expectedLevel`; manager/RH `promedio ponderado (auto + rh) vs expectedLevel`; radar oculta dataset RH en self-view. Alternativa descartada: ponderar en frontend (divergencia de cálculo).
- **Textarea 1 línea + resize visible; toasts con `error.code`**: contratos de error existentes (`code, message, details[], trace_id`) en `api/openapi/`.

## Risks / Trade-offs

- [Auto-creación concurrente duplica Evaluation] → Idempotencia por (cycle_id, employee_id) + índice único.
- [Peso global mal configurado distorsiona brecha] → Backend valida pesos; frontend muestra fallback si falta.
- [Ocultar columna solo en UI filtra por error] → RBAC en service layer; la UI solo presenta.
