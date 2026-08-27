# Proposal: Mis-Evaluados Status, Validación de Pesos y Navegación de Metas por Época

## Intent

En el flujo de fijación y evaluación de metas del portal SED, se identificaron brechas clave de experiencia de usuario y persistencia del estado de asignación:

1. **Falta de estado y auditoría en `GoalAssignment`**: La entidad `GoalAssignment` únicamente persiste `id`, `employee_id` y `cycle_id`. No existe registro de si la asignación de metas está en elaboración (`borrador`) o si ya fue enviada formalmente (`enviada`), ni la marca temporal de envío (`submitted_at`).
2. **Auto-POST prematuro en frontend**: El cliente web realizaba una creación silenciosa de asignaciones vacías en `_doLoad()` al consultar subordinados, generando registros en BD antes de que el usuario o jefe hubiera formulado metas válidas.
3. **Validación de pesos desconectada del estado**: La regla Double 100% (`ValidateDoubleWeighting`) evalúa que la suma de categorías sea 100% y que las metas dentro de cada categoría sumen 100%. Sin embargo, esta validación no condicionaba el cambio formal de estado de la asignación ni bloqueaba envíos inválidos en backend.
4. **Navegación ambigua para el evaluador (Jefe) en `mis-evaluados`**:
   - En la tabla de `mis-evaluados`, no se visualiza el estado de asignación de metas del colaborador.
   - No existe un botón de acción directo "Metas" adaptado a la fase del ciclo anual.
   - En **inicio de año** (o con metas en `borrador`), el jefe necesita acceder a la vista de formulación/edición (`/objetivos/asignacion?empId={id}`).
   - En **medio de año** y **fin de año** (o con metas `enviada`), el jefe debe acceder a la vista de seguimiento y evaluación de cumplimiento (`/mis-evaluados/{id}/metas`).

## Scope

### In Scope
- **Modelo de datos y BD**:
  - Migración SQL `000039` agregando enum `goal_assignment_status` (`borrador`, `enviada`) y columna `submitted_at` (`TIMESTAMPTZ`, nullable) a la tabla `goal_assignments`.
  - Actualización del esquema Ent `GoalAssignment` (`status`, `submitted_at`) y regeneración de entidades.
- **Backend API & Lógica**:
  - Actualización de OpenAPI `api/openapi/goals-api.yaml` (`AssignmentResponse` con `status` y `submitted_at`).
  - Mapeo de `AssignmentRow` y `dtogoal.AssignmentResponse`.
  - Transición de estado en `GoalHandler.CreateAssignment` / endpoint de envío: validar que Double 100% (`ValidateDoubleWeighting`) sea válido para marcar `status = 'enviada'` y `submitted_at = now()`. Si es inválido, permanece en `borrador` y rechaza el envío con error de validación.
- **Frontend Stores & Tipos**:
  - Actualización de `EmployeeAssignment` en `web/src/lib/types/goal.ts` con `status` y `submittedAt`.
  - Mapeo de `status` y `submitted_at` en `normalizeApiData()` de `goalsStore.svelte.ts`.
  - Eliminación del auto-POST silencioso de creación de asignaciones en `_doLoad()`.
- **Frontend UI & Rutas**:
  - `/objetivos/asignacion/+page.svelte`: soporte para query param `?empId=` inicializando `selectedEmployeeId` y cargando los datos del colaborador vía `loadForEmployee(empId)`.
  - `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte`: corrección de columnas, badge de estado de metas y botón **Metas** con navegación contextual por fase/estado.
  - Creación de la ruta `/mis-evaluados/[id]/metas/+page.svelte` para la consulta/revisión de metas en medio y fin de año.

### Out of Scope
- Configuración de ponderación global 30/70 y tabla `cycle_config` (Track `b` en change separado `global-goal-peso-30-70`).
- Sincronización batch nocturna de `is_active` desde Mobonet (Track `c` en change separado `mobonet-sync-batch`).
- Paginación del servidor en la pantalla de asignación de objetivos.

## Capabilities

### New Capabilities
- `goal-assignment-status-tracking`: Persistencia del ciclo de vida de la asignación (`borrador` vs `enviada`) con marca temporal `submitted_at`.
- `manager-evaluatee-goals-view`: Ruta dedicada `/mis-evaluados/[id]/metas` para que los evaluadores consulten y evalúen metas de su equipo en medio y fin de año.
- `contextual-goals-navigation`: Enrutamiento inteligente desde `mis-evaluados` hacia asignación (`/objetivos/asignacion?empId=`) o evaluación (`/mis-evaluados/[id]/metas`) según la época del ciclo.

### Modified Capabilities
- `goals-and-weighting`: La validación de pesos Double 100% ahora condiciona directamente el cambio de estado a `enviada` y el sellado de `submitted_at`.
- `mis-evaluados-ui`: La tabla de evaluados incorpora estado de metas y botón de acceso a metas.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/cmd/server/migrations/` | Modified | Nueva migración `000039_add_goal_assignment_status_and_submitted_at.up.sql` / `.down.sql` |
| `api/internal/schema/goalassignment.go` | Modified | Nuevos campos `status` y `submitted_at` en Ent schema |
| `api/internal/repository/goal/assignment_repo.go` | Modified | Soporte de `status` y `submitted_at` en `AssignmentRow`, queries y creación |
| `api/internal/service/goal/` | Modified | Validación de pesos vinculada al guardado/envío de asignación |
| `api/internal/handler/goal/goal_handler.go` | Modified | DTO response y validación en `CreateAssignment` |
| `api/openapi/goals-api.yaml` | Modified | Contrato `AssignmentResponse` con `status` y `submitted_at` |
| `web/src/lib/types/goal.ts` | Modified | Campos `status` y `submittedAt` en `EmployeeAssignment` |
| `web/src/lib/stores/goalsStore.svelte.ts` | Modified | Normalización de status, eliminación de POST silencioso |
| `web/src/routes/objetivos/asignacion/+page.svelte` | Modified | Lectura de query param `?empId=` para selección de colaborador |
| `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte` | Modified | Corrección visual de columnas, badge de status y botón Metas |
| `web/src/routes/mis-evaluados/[id]/metas/+page.svelte` | Added | Nueva ruta para visualización de metas de evaluados en medio/fin de año |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Asignaciones existentes sin valor en `status` | Low | Migración define `DEFAULT 'borrador'` para `status`. |
| Evaluador accede a metas de empleado fuera de su jerarquía | Medium | Backend valida permisos y jerarquía del caller en `parseEmpID` / middleware; frontend solo navega a evaluados directos/indirectos. |
| Incompatibilidad de tipos generados en TypeScript | Low | Mantener contratos OpenAPI sincronizados y campos opcionales/compatibles. |
