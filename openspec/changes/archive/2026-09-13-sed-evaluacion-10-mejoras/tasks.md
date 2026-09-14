# Tasks: sed-evaluacion-10-mejoras

> Solo specs/plan (no código app). Código del repo es verdad (EvaluationDetail, permisos jefa). Tester valida por caso: avance-escribe-avance, cierre-escribe-cierre, revert-editable. Backfill y `finished_at` solo referencia (no implementar, no contradecir). Máx 5 archivos app.

## Implementación (8 pasos)

- [x] 1. `api/internal/repository/evaluation/evaluation_repo.go` → `Ensure`/`FindBy(employee,cycle,phase)` con fila propia por fase
  - Resultado: avance y cierre tienen filas propias; escribir en una fase no altera la otra; `cierre` sin fila se auto-crea idempotente (sin 404, concurrente OK).
- [x] 2. `api/internal/service/evaluation/evaluation_service.go` → revert `cierre`→`avance` conserva ambas fases
  - Resultado: tras revert, snapshot/data de avance y cierre intactas; fase activa `avance` editable. Depende de: 1.
- [x] 3. `api/internal/service/evaluation/evaluation_service.go` → gates por fase + auth RH
  - Resultado: self/manager-RH escriben en `avance`/`medio-anio`/`cierre`; Submit/Finalize solo `cierre` (si no → `409 PHASE_NOT_ADVANCEABLE`); RH escribe con `eval:rh` o jefe asignado (si no → `403`; sin sesión → `401`). Depende de: 1.
- [x] 4. `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte` → comentarios y ratings por (rol,fase) independientes
  - Resultado: cada rol/fase edita lo suyo sin pisar otras fases; permisos jefa ver/escribir según código actual; comentario enviado persiste visible en fase actual. Depende de: 1.
- [x] 5. `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte` → labels por fase + avance con progreso/autoeval y comentarios jefe/RH visibles
  - Resultado: `avance` muestra "Evaluación de avance de medio año / Guardar avance", `cierre` "Evaluación de cierre de año / Guardar cierre"; en avance, progreso+autoeval habilitados y comentarios jefe/RH visibles.
- [x] 6. `web/src/routes/mi-evaluacion/+page.svelte` → EmptyState con gate `assignmentStatus` + botón "Ir a metas"
  - Resultado: categorías sin metas o metas no enviadas muestran empty con botón que navega a metas (sin redirect auto).
- [x] 7. `web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte` → avatar iniciales correctas + textarea 1 línea con resize + comentarios con botón enviar
  - Resultado: avatar nunca en blanco (iniciales del nombre correcto, sin S3); textarea 1 línea con resize usable; comentario se envía y persiste.
- [x] 8. (DUEÑO ÚNICO) `web/src/lib/components/ui/ErrorState.svelte` → `ErrorState` único + toasts con `error.code`
  - Resultado: todo fallo muestra `error.code`, éxito muestra confirmación; `ErrorState` vive solo aquí (no duplicar en otros changes). Depende de: 1–7.

## Validación por caso (tester)

- Caso A avance-escribe-avance: escribir en `avance` solo cambia fila `avance` (cierre intacto).
- Caso B cierre-escribe-cierre: escribir en `cierre` solo cambia fila `cierre` (avance intacto).
- Caso C revert-editable: tras revert `cierre`→`avance`, ambas filas conservan data y `avance` acepta escritura.
