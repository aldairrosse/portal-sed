## Context

Ver `proposal.md` (Why) y `specs/` (evaluation-lifecycle, goals-and-weighting). Estado actual: el repo es la verdad — `EmployeeEvaluationDetail.svelte` define permisos reales de jefa (ver/escribir comentarios de metas/competencias); `evaluation_repo.go` hoy reusa fila sin importar fase y pisa competencias entre fases. Stack: Svelte 5 + DaisyUI en `web/`, Go + Chi en `api/`, contratos OpenAPI 3.1 en `api/openapi/`.

## Goals / Non-Goals

**Goals:**

- Fila propia por (employee,cycle,phase): avance escribe avance, cierre escribe cierre.
- Comentarios y ratings por rol y fase separados, independientes, editables en fase actual incluso tras revert.
- Revert cierre→avance conserva ambas fases, sin borrados.
- Gates por fase claros y `ErrorState` con dueño único en este change.

**Non-Goals:**

- Backfill por fase (no se ejecuta aquí; no contradecirlo). Quitar `finished_at` en cierre (solo referencia). Ver proposal Non-goals.

## Decisions

- **Q1 código es verdad**: `EvaluationDetail` y permisos de jefa (ver/escribir comentarios metas/competencias) se toman del repo; si spec y código difieren, se corrige la spec. Alternativa descartada: redefinir permisos en spec (rompe comportamiento verificado).
- **Q2 separación por rol y fase**: avances/ratings + comentarios se almacenan y editan por (rol,fase) de forma independiente; tras revert cierre→avance, la fase `avance` sigue editable sin afectar la data de `cierre`. Alternativa descartada: fila única compartida (pisa competencias entre fases — bug actual).
- **Q3 backfill compatible**: el modelo por fase deja hueco para un backfill futuro sin migraciones destructivas; no se implementa ni se bloquea aquí.
- **Q4 `finished_at` solo referencia**: no se toca en este change; se menciona para no reintroducirlo por accidente.
- **Q5 tasks trazables**: cada task mapea archivo→resultado esperado; incluye Ensure/FindBy(employee,cycle,phase), revert conserva ambas fases, gates por fase y `ErrorState` dueño único. Alternativa descartada: tasks genéricas sin archivo (no verificables).
- **Q6 validación por caso**: tester valida avance-escribe-avance, cierre-escribe-cierre y revert-editable como casos independientes.
- **Auto-crear Evaluation en backend al abrir cierre**: idempotente por (cycle_id, employee_id, phase); el 404 desaparece sin creación manual. Alternativa descartada: crear desde frontend (duplicados en concurrencia).
- **EmptyState con gate `assignmentStatus` sin redirect auto**: el usuario decide vía "Ir a metas". Alternativa descartada: redirect automático (rompe deep-links).
- **Avatar conserva iniciales, fix de nombre visible**: sin migración S3. Labels por fase en frontend derivados de la fase del ciclo. Textarea 1 línea + resize; toasts con `error.code` (contratos existentes `code, message, details[], trace_id`).

## Risks / Trade-offs

- [Auto-creación concurrente duplica Evaluation] → Idempotencia por (cycle, employee, phase) + índice único.
- [Revert interpretado como borrado] → Revert solo cambia fase activa; jamás borra filas de ninguna fase.
- [Ocultar/colorear por rol solo en UI filtra por error] → RBAC en service layer; la UI solo presenta (permisos de jefa según código).
