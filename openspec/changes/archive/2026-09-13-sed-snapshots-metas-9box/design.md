## Context

Ver `proposal.md` (Why). Estado actual: `evaluation_goals` sin snapshots por fase; `PUT goal-state` deriva `final_rating`; `RecomputeMatrix` y `ComputeMatrixView` divergen en ponderación; ausencias tratadas como 0; comentarios sin autor/fecha uniformes. Restricciones: PostgreSQL + ORM con migraciones versionadas; OpenAPI 3.1 como contrato; RBAC en service. Código es verdad para `EvaluationDetail` y permisos de jefa (Q1): esta spec no los redefine.

## Goals / Non-Goals

**Goals:**

- Persistir `avance_progress`/`cierre_progress` y sincronizar `current_value` según fase activa, con fase previa inmutable.
- Backfill por fase activa: `avance` → `avance_progress=current_value`, `cierre` → `cierre_progress=current_value`.
- Separación total por fase de metas/ratings y comentarios; editable solo fase actual.
- Unificar ponderado 0.8/0.2 ignorando ausentes en matriz y dominio de competencias.

**Non-Goals:**

- Backfill histórico de snapshots; cambios de ponderación de metas; notificaciones email.

## Decisions

- Snapshot como columnas NULL en `evaluation_goals` (no tabla nueva): alternativa tabla de historial descartada por sobre-ingeniería para 2 fases.
- `current_value = cierre ?? avance` como regla de lectura + escritura dual en `UpdateGoalProgress`/cierre: alternativa trigger BD descartada (lógica en service, testeable).
- Backfill por fase activa (Q3): si `cycles.current_phase=avance` → `avance_progress=current_value`; si `cierre` → `cierre_progress=current_value`. Alternativa "siempre avance" descartada (corrompe cierre).
- Inmutabilidad por fase (Q2): escribir solo el snapshot de la fase activa, nunca tocar la otra; metas/ratings y comentarios separados por fase, previa inmutable.
- Cierre usa fase activa, no `finished_at` (Q4): no contradice el change que quitó `finished_at`; `finished_at` queda fuera de este change.
- 9-box por fase: snapshot avance lee evaluaciones de avance, snapshot cierre lee evaluaciones de cierre.
- Eliminar derivación `final_rating`; UI formatea por `unit`: alternativa mantener rating descartada (duplica fuente de verdad).

## Risks / Trade-offs

- [Risk] Escrituras concurrentes avance/cierre → Mitigación: transacción por meta + `updated_at` como guarda.
- [Risk] Cambio de contrato `PUT goal-state` rompe clientes → Mitigación: versionar en `/api/v1/`, documentar en OpenAPI, error 400 con `code/message/details[]`.
