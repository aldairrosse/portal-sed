# Design: SED Medio Cierre 9Box

## Context

Enum canónico backend: `asignacion/avance/cierre` (`api/internal/schema/evaluation.go:30-37`, `phasedefinition.go:120-124`, `machine.go:35-98`). `goal_comments` hoy sin `phase` (`000010_goal_comments.up.sql`); migraciones `000042`/`000043` consolidan `medio_anio` a `avance`. `Evaluation` ya tiene unique `(employee_id,cycle_id,phase)`; `EvaluationCompetency` tiene unique `(evaluation_id,competency_id)` y debe ampliarse. El alias `medio-anio`/`medio_anio` es solo lectura en frontend (`web/src/lib/types/cycle.ts`, `cycle.svelte.ts:2,16`).

## Decisions

### D1. OpenAPI 3.1 primero

Cada delta (comments, competency, redact, ninebox) se contrata en `api/openapi/` antes de handlers o UI. Tipos TS generados con `openapi-typescript` en `web/src/lib/api/`.

### D2. RBAC en backend siempre

Permisos rol × fase se verifican en handlers Go (`Chi`), no solo en UI. Colaborador: `self` en `avance`; detalle ajeno → 403. Redacción `RedactDetailForSelf` por `viewerMode` + phase.

### D3. Ent versionado + migraciones goose

- `goal_comments.phase ENUM(asignacion,avance,cierre) NOT NULL DEFAULT cierre` + índice `(goal_id,phase)` + backfill documentado.
- `EvaluationCompetency` unique → `(evaluation_id,competency_id,source[,phase])` o evaluaciones separadas por phase (decidir en tasks; no duplicar ambos).

### D4. Alias solo frontend

`medio-anio` no entra al backend ni a la BD. Mapper centralizado en `cycle.ts`/`cycle.svelte.ts`; el backend solo acepta `avance` para medio-año.

### D5. 9×9 por fase

`ComputeMatrixView` y `RecomputeMatrix` exigen `phase_id`; las vistas de medio-año y cierre se computan y cachean por separado.
