## 1. Empty state de mi-evaluación con acceso a metas

- [ ] 1.1 Mostrar EmptyState con gate `assignmentStatus` y botón "Ir a metas" en `web/src/routes/mi-evaluacion/+page.svelte` y verificar que con categorías sin metas o metas no enviadas se ve el empty con el botón navegando a metas.
- [ ] 1.2 Cubrir escenarios categorías-sin-metas y metas-no-enviadas con test de componente y verificar que `pnpm vitest` pasa en verde.

## 2. Avatar con iniciales correctas

- [ ] 2.1 Corregir derivación de iniciales del nombre correcto del empleado (sin render en blanco, sin migración S3) en componente avatar/iniciales y verificar visualmente que el avatar muestra iniciales con nombre correcto y nunca en blanco.
- [ ] 2.2 Agregar test de iniciales (nombre correcto, vacío, un solo nombre) y verificar que el test pasa.

## 3. Breadcrumb contextual por rol

- [ ] 3.1 Implementar breadcrumb por rol en `EmployeeEvaluationDetail.svelte` (`Mi evaluación > Rol` self, `Evaluaciones > Rol` rh, `Mis evaluados > Rol` jefe/director) y verificar cada rol navegando como self, rh y jefe.
- [ ] 3.2 Agregar tests de breadcrumb por rol y verificar que los tres escenarios pasan.

## 4. Columna Evaluación y grid uniforme

- [ ] 4.1 Renombrar columna `RH` a `Evaluación`, ocultarla si `employeeId === session.user.employeeId`, y aplicar `sized cols` uniformes entre pilares, y verificar self-view sin columna, tercero con columna visible y columnas alineadas entre pilares.
- [ ] 4.2 Agregar tests de visibilidad por contexto y verificar que pasan.

## 5. Labels diferenciados de avance y cierre

- [ ] 5.1 Usar labels por fase (`Evaluación de avance de medio año` / `Guardar avance` vs `Evaluación de cierre de año` / `Guardar cierre`) y verificar que al cambiar de fase el título y botón cambian y RH/ciclos refresca labels tras revert.
- [ ] 5.2 Agregar test de labels por fase y verificar que ambos escenarios pasan.

## 6. Auto-creación de Evaluation en cierre y envío de comentarios

- [ ] 6.1 Auto-crear Evaluation por meta en `api/internal/service/evaluation/` al abrir cierre (idempotente por cycle_id + employee_id, sin 404) y verificar con test de tabla que abrir cierre sin Evaluation la crea una sola vez aun con llamadas concurrentes.
- [ ] 6.2 Habilitar comentarios con botón enviar en `GoalClosureCard` y verificar que el comentario se envía y persiste visible.
- [ ] 6.3 Documentar `error.code` y labels de fase en `api/openapi/evaluations-and-9x9.yaml` y `goals-api.yaml`, regenerar tipos TS en `web/src/lib/api/` y verificar que `pnpm exec openapi-typescript` genera sin errores.

## 7. Revert cierre → avance conservando data

- [ ] 7.1 Permitir revert de `cierre` a `avance` conservando snapshot y data en `api/internal/service/evaluation/` y vista `rh/ciclos`, y verificar con test que tras el revert la data previa sigue presente y la evaluación queda editable en `avance`.

## 8. Fase de avance: progreso, autoevaluación y comentarios visibles

- [ ] 8.1 Habilitar registro de progreso y autoevaluación más comentarios de jefe/RH visibles en fase `avance` y verificar que el empleado registra progreso, completa autoeval y ve comentarios de jefe y RH.
- [ ] 8.2 Agregar test de gates de avance y verificar que pasa.

## 9. Textarea y toasts

- [ ] 9.1 Renderizar textarea de comentarios con una línea por defecto y resize visible, y mostrar toasts con `error.code` claro en fallo y confirmación en éxito, y verificar manualmente altura inicial, resize usable y ambos toasts.
- [ ] 9.2 Agregar test de textarea/toasts y verificar que pasa.

## 10. Brecha ponderada global y radar en self-view

- [ ] 10.1 Exponer peso global auto/rh desde backend (`ninebox_service.go`) y calcular brecha (`self vs expectedLevel` en self-view; promedio ponderado vs `expectedLevel` en manager/RH), ocultando dataset RH del radar en self-view, y verificar brecha self y ponderada contra fixtures y radar sin línea RH en self-view.
- [ ] 10.2 Agregar tests de cálculo de brecha por vista y verificar que pasan, cerrando con `rtk proxy openspec validate --all --strict --json` en verde.
