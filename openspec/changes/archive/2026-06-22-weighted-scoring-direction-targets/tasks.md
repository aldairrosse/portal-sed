## 1. Backend — Migración y modelo de datos

- [x] 1.1 Agregar campo `direction` (enum: `ascendente`, `descendente`, default `ascendente`) y `baseline_value` (float64, nullable) a la entidad `Goal` en Ent schema (`api/internal/migrate/schema.go` y schema files)
- [x] 1.2 Agregar campo `direction` (enum: `ascendente`, `descendente`, default `ascendente`) y `current_value` (float64, nullable) a la entidad `KPI` en Ent schema
- [x] 1.3 Generar migración Ent y verificar compilación (sin aplicar a BD)
- [x] 1.4 Actualizar seeds (`api/internal/seed/goals.go`) para incluir `direction` explícita: meta "Reducir ausentismo" → descendente con baselineValue 100; meta "Reducir quejas" → descendente con baselineValue 50; KPI "Ausentismo" → descendente; KPI "Rotación de inventario" → descendente

## 2. Backend — Paquete scoring

- [x] 2.1 Crear `api/internal/pkg/scoring/progress.go` con función `ProgressPercent(currentValue, targetValue, baselineValue float64, direction string) float64` — ascendente: `min(current/target*100, 100)`, descendente: `min((baseline-current)/(baseline-target)*100, 100)`, clamp 0–100
- [x] 2.2 Crear `api/internal/pkg/scoring/progress_test.go` con tests de tabla para ascendente parcial, completo, clamp; descendente parcial, completo, sin progreso, peor que baseline, baseline=target (división por cero)
- [x] 2.3 Crear `api/internal/pkg/scoring/weighted.go` con función `EmployeeScore(categories []CategoryScore) float64` — `Σ(cat_weight/100 × Σ(goal_weight/100 × progress%))`
- [x] 2.4 Crear `api/internal/pkg/scoring/weighted_test.go` con tests de scoring completo

## 3. Backend — GoalService con dirección

- [x] 3.1 Modificar `GoalService.CreateGoal` para aceptar `direction` y `baselineValue` en el request; validar que `baselineValue` es requerido cuando `direction === 'descendente'` y que `baselineValue > targetValue`
- [x] 3.2 Modificar `GoalService.UpdateGoal` para congelar `direction` y `baselineValue` en fase `avance`
- [x] 3.3 Modificar `ProgressService.UpdateGoalProgress` para usar `scoring.ProgressPercent()` con dirección al calcular internamente (si aplica cálculo server-side)
- [x] 3.4 Actualizar DTOs (`api/internal/dto/goal/goal_dto.go`): `CreateGoalRequest` y `UpdateGoalRequest` con campos `direction` y `baseline_value`; `GoalResponse` con `direction`, `baseline_value` y `progress_percent`
- [x] 3.5 Actualizar repositorio de goals para incluir nuevos campos en CREATE/UPDATE/SELECT

## 4. Backend — KPIService con currentValue y dirección

- [x] 4.1 Crear handler `PATCH /api/v1/kpis/{kpiId}/value` en `api/internal/handler/goal/goal_handler.go` — acepta `{ "current_value": float64 }`, valida `current_value >= 0`, retorna KPI actualizado
- [x] 4.2 Crear método `UpdateKPIValue` en `KPIService` o nuevo `KPIValueService`
- [x] 4.3 Crear query en repositorio para actualizar solo `current_value` del KPI
- [x] 4.4 Registrar ruta PATCH en `api/internal/handler/goal/routes.go`
- [x] 4.5 Modificar `KPIResponse` para incluir `direction`, `current_value` y `progress_percent`

## 5. Backend — ScoringService

- [x] 5.1 Crear `api/internal/service/goal/scoring_service.go` con método `GetEmployeeScore(ctx, empID) (float64, error)` que carga categorías → metas → calcula score usando `scoring.EmployeeScore()`
- [x] 5.2 Crear handler `GET /api/v1/employees/{empId}/score` que retorna `{ "score": float64 }`
- [x] 5.3 Registrar ruta en routes.go

## 6. Backend — Tests de integración

- [x] 6.1 Test de integración para `PATCH /kpis/{id}/value` — happy path, KPI no existe, valor negativo
- [x] 6.2 Test de integración para `GET /employees/{id}/score` — empleado con categorías y metas
- [x] 6.3 Test de integración para crear meta descendente con baselineValue requerido
- [x] 6.4 Test de integración para bloquear cambio de direction en fase avance

## 7. Frontend — Tipos y utilidades

- [x] 7.1 Actualizar tipo `Goal` en `web/src/lib/types/goal.ts` — agregar `direction: 'ascendente' | 'descendente'`, `baselineValue?: number`, `progressPercent?: number`
- [x] 7.2 Actualizar tipo `KPI` — agregar `direction: 'ascendente' | 'descendente'`, `currentValue?: number`, `progressPercent?: number`
- [x] 7.3 Crear `web/src/lib/utils/scoring.ts` con `progressPercent(current, target, baseline, direction)` reutilizando la lógica del backend
- [x] 7.4 Crear tests para `scoring.ts` en `web/src/lib/utils/__tests__/scoring.test.ts`

## 8. Frontend — Store y normalización

- [x] 8.1 Actualizar `normalizeApiData()` en `goalsStore.svelte.ts` para mapear `direction`, `baseline_value`, `current_value` de la API
- [x] 8.2 Actualizar fixtures `goals.json` y `kpis.json` con campos `direction` y `currentValue`
- [x] 8.3 Crear función `getWeightedScore()` en el store que calcule el score ponderado del empleado actual
- [x] 8.4 Actualizar `getCategoryProgressAverage()` para usar `progressPercent()` con dirección
- [x] 8.5 Actualizar `getGoalPermissions()` para congelar `direction` y `baselineValue` en fase `avance` (medio-anio)

## 9. Frontend — Formulario de meta con dirección

- [x] 9.1 Agregar radio group "Dirección" en formulario de meta (inline o modal) con opciones "Ascendente (↑)" y "Descendente (↓)", default ascendente
- [x] 9.2 Agregar input condicional "Valor inicial (baseline)" que solo aparece cuando direction === 'descendente', con helper text y validación `baselineValue > targetValue`
- [x] 9.3 Ocultar input de baseline cuando direction es ascendente
- [x] 9.4 Actualizar `validateGoal()` en `goalValidation.ts` para validar baselineValue requerido para descendente

## 10. Frontend — GoalRow con indicadores +/-

- [x] 10.1 Crear componente `DeltaIndicator` o extender `GoalRow` para mostrar badge +/- cuando currentValue supera/empeora target
- [x] 10.2 Lógica para ascendente: delta = currentValue - targetValue; badge `+N` verde si positivo, `-N` ámbar si negativo
- [x] 10.3 Lógica para descendente: delta calculado desde baseline; badge `+N%` verde si mejoró, `-N` rojo si empeoró
- [x] 10.4 Sin indicador cuando currentValue == targetValue

## 11. Frontend — KPIs con progreso

- [x] 11.1 Extender `KpiBadge` o crear `KpiProgressBar` para mostrar `currentValue` y cumplimiento del KPI
- [x] 11.2 Reutilizar `progressPercent()` para calcular cumplimiento del KPI según dirección
- [x] 11.3 Mostrar indicador +/- en KPI cuando supera/empeora target
- [x] 11.4 Mostrar "Sin datos" cuando KPI no tiene `currentValue`

## 12. Frontend — Score ponderado en UI

- [x] 12.1 Mostrar score ponderado del empleado en pantalla de asignación (`/objetivos/asignacion`) como número 0–100
- [x] 12.2 Agregar desglose por categoría (expandible) mostrando score individual por categoría
- [x] 12.3 Score se actualiza en tiempo real al editar pesos o progreso

## 13. Frontend — Tests

- [x] 13.1 Test para `getWeightedScore()` en el store
- [x] 13.2 Test para `getCategoryProgressAverage()` con metas ascendentes y descendentes
- [x] 13.3 Test para `getGoalPermissions()` con dirección congelada en avance
- [x] 13.4 Test para formulario de meta con dirección y baseline condicional
- [x] 13.5 Test para `DeltaIndicator` con escenarios ascendente/descendente

## 14. OpenAPI

- [x] 14.1 Actualizar `api/openapi/goals-api.yaml` — schema Goal con `direction` y `baseline_value`; schema KPI con `direction` y `current_value`
- [x] 14.2 Agregar endpoint `PATCH /kpis/{kpiId}/value` con request/response schema
- [x] 14.3 Agregar endpoint `GET /employees/{empId}/score` con response schema
- [ ] 14.4 Regenerar tipos TS con `openapi-typescript` (skipped per PR6 rules — already updated manually in PR3)
