# Tasks: sed-snapshots-metas-9box

> Specs/plan solo (no código app). Q1: código es verdad (`EvaluationDetail`, permisos jefa). Q2: separación total por fase, editable solo fase actual, previa inmutable. Q3: backfill por fase activa. Q4: cierre usa fase activa, no `finished_at`. 9-box por fase. Competency-framework no se toca en este paso.

## Implementación (5 tasks)

- [x] 1. Reconciliar 000048 + backfill por fase activa
  - Archivo → resultado: reconciliar `api/cmd/server/migrations/000048_*.sql` (ya existe/committed con backfill incondicional a `avance_progress`) → backfill condicional por `current_phase` (`avance`→`avance_progress=current_value`, `cierre`→`cierre_progress=current_value`); renumerar nueva migración si aplica (Q3 no es crear 000048).
  - Spec: `specs/evaluations/spec.md` (Requirement Backfill por fase activa).
  - Validación tester: meta con `current_value=60` en fase `avance` ⇒ `avance_progress=60, cierre=NULL`; misma meta en fase `cierre` ⇒ `cierre_progress=60, avance=NULL`.
- [x] 2. Snapshot por fase + sincronización `current_value`
  - Archivo → resultado: handler/service `UpdateGoalProgress` → escribe solo snapshot de fase activa (`cycles.current_phase`), `current_value = cierre_progress ?? avance_progress`, previa inmutable.
  - Spec: `specs/evaluations/spec.md` (Snapshot + Sincronización + Separación por fase).
  - Validación tester: escribir `75` en `avance` no toca `cierre_progress`; leer con ambos prioriza `cierre`; escribir en `cierre` no toca `avance_progress`.
- [x] 3. `PUT goal-state` (cierre) valor directo por `unit`
  - Archivo → resultado: `PUT goal-state` → persiste snapshot de fase activa + sincroniza `current_value`; NO deriva `final_rating`; UI muestra valor por `unit`; fase determinada por `current_phase`, no por `finished_at`.
  - Spec: `specs/evaluations/spec.md` (Cierre con valor directo).
  - Validación tester: cierre `90 porcentaje` ⇒ snapshot cierre `90`, `current_value=90`, sin `final_rating`; `unit` no soportada → `400`, fase desconocida → `400`, sin sesión → `401`.
- [x] 4. 9-box por fase + ponderado
  - Archivo → resultado: `ninebox_repo.go: GetGoalProgressByEmployee` → lee snapshot de la fase pedida (`avance_progress` en avance, `cierre_progress` en cierre); `RecomputeMatrix` con ponderado `0.8 RH / 0.2 self` ignorando ausentes + escalado `direction` → % (clamp 0–100) → tier 1–3.
  - Spec: `specs/ninebox-computed/spec.md` (Fuente por fase + Escalado + RecomputeMatrix).
  - Validación tester: matriz avance usa solo `avance_progress`, matriz cierre usa solo `cierre_progress`; solo-RH promedia solo RH; `RecomputeMatrix == ComputeMatrixView` en matriz dorada.
- [x] 5. Comentarios por fase con autor/fecha
  - Archivo → resultado: `goalcomment` (+ mig manager-comment) → autor desde sesión + fecha/hora visibles, separados por fase, editable solo fase actual.
  - Spec: `specs/evaluations/spec.md` (Comentarios + Separación por fase).
  - Validación tester: comentario en `avance` no aparece editable en `cierre` y viceversa; crear persiste autor/fecha/hora; lectura los devuelve visibles.

## Notas

- OpenAPI + tipos TS (`PUT goal-state`, snapshots, comentarios) se actualizan dentro de cada task que cambia contrato (no task separado).
- `finished_at`, `EvaluationDetail` y permisos jefa quedan fuera: no contradecir código existente.
