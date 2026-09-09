# Tasks: SED Medio Cierre 9Box

## Fase 1 — OpenAPI

- [ ] Contratos 3.1 para comments (GET/POST por phase), competency (source+phase), redact (viewerMode+phase), ninebox (`phase_id` obligatorio)
- [ ] Generar tipos TS (`openapi-typescript`) en `web/src/lib/api/`

## Fase 2 — Migración

- [ ] Migración `goal_comments.phase ENUM(asignacion,avance,cierre) NOT NULL DEFAULT cierre` + índice `(goal_id,phase)` + backfill documentado
- [ ] Migración unique `EvaluationCompetency` → `(evaluation_id,competency_id,source[,phase])` o evaluaciones separadas por phase

## Fase 3 — RBAC / redact

- [ ] Gates rol × fase en handlers (colaborador `self` en avance, 403 ajeno)
- [ ] `RedactDetailForSelf` por `viewerMode` + phase; colaborador solo comentarios

## Fase 4 — Frontend gates

- [ ] Alias `medio-anio` solo lectura centralizado en `cycle.ts` + `cycle.svelte.ts`
- [ ] Gates UI por fase; sin enviar `medio-anio` al backend

## Fase 5 — 9Box

- [ ] `ComputeMatrixView`/`RecomputeMatrix` con `phase_id` obligatorio; vistas avance vs cierre separadas

## Fase 6 — Validación

- [ ] `rtk proxy openspec validate --all` en verde
- [ ] Tests de tabla Go (RBAC, redact, unique) + `go test ./...`
