# Tasks: Mis-Evaluados Status, Validación de Pesos y Navegación de Metas por Época

## Task 1: Migración de Base de Datos y Esquema Ent

- [x] Crear migración SQL `api/cmd/server/migrations/000039_add_goal_assignment_status_and_submitted_at.up.sql`:
  - [x] Crear tipo enum `goal_assignment_status` con valores `'borrador'`, `'enviada'`.
  - [x] Agregar columnas `status` (`goal_assignment_status DEFAULT 'borrador' NOT NULL`) y `submitted_at` (`TIMESTAMPTZ NULL`) a `goal_assignments`.
- [x] Crear migración SQL `api/cmd/server/migrations/000039_add_goal_assignment_status_and_submitted_at.down.sql`.
- [x] Actualizar esquema Ent `api/internal/schema/goalassignment.go` con los campos `status` y `submitted_at`.

---

## Task 2: Backend Repository, DTO y Validación de Pesos

- [x] Actualizar `AssignmentRow` en `api/internal/repository/goal/assignment_repo.go` para incluir `Status` y `SubmittedAt`.
- [x] Actualizar `CreateAssignment` o agregar método `CreateOrSubmitAssignment` en `assignment_repo.go` para persistir `status` y `submitted_at`.
- [x] Actualizar `AssignmentResponse` en `api/internal/dto/goal/goal_dto.go` con `Status` y `SubmittedAt`.
- [x] Modificar `GoalHandler.CreateAssignment` en `api/internal/handler/goal/goal_handler.go`:
  - [x] Validar Double 100% de pesos con `h.weightSvc.ValidateDoubleWeighting(ctx, empID)`.
  - [x] Si es válido, asignar `status = "enviada"` y `submitted_at = time.Now()`.
  - [x] Si es inválido, mantener `status = "borrador"` y `submitted_at = nil` (o responder con error si es intento explícito de envío).

---

## Task 3: Contrato OpenAPI 3.1

- [x] Actualizar `api/openapi/goals-api.yaml`:
  - [x] Agregar `status` (`enum: [borrador, enviada]`) y `submitted_at` (`type: string, format: date-time, nullable: true`) en el schema `AssignmentResponse`.

---

## Task 4: Frontend Types y Goals Store

- [x] Actualizar interfaz `EmployeeAssignment` en `web/src/lib/types/goal.ts` con `status?: 'borrador' | 'enviada'` y `submittedAt?: string | null`.
- [x] Actualizar `normalizeApiData()` en `web/src/lib/stores/goalsStore.svelte.ts` para mapear `status` y `submitted_at`.
- [x] Eliminar la creación automática mediante POST silencioso en `_doLoad()` de `goalsStore.svelte.ts`.

---

## Task 5: Frontend UI, Rutas y Navegación Contextual

- [x] Actualizar `web/src/routes/objetivos/asignacion/+page.svelte`:
  - [x] Leer el query param `?empId=` desde `$page.url.searchParams`.
  - [x] Inicializar `selectedEmployeeId` con dicho parámetro y disparar `loadForEmployee(empId)`.
- [x] Actualizar `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte`:
  - [x] Corregir la alineación de columnas `<thead>` y `<tbody>`.
  - [x] Agregar badge de estado de metas (`Enviada` / `Borrador`).
  - [x] Agregar botón **"Metas"** en la columna de Acciones:
    - [x] `inicio-anio` (o `borrador`): enlace a `/objetivos/asignacion?empId=${row.id}`.
    - [x] `medio-anio` / `fin-anio` (o `enviada`): enlace a `/mis-evaluados/${row.id}/metas`.
- [x] Crear la ruta `web/src/routes/mis-evaluados/[id]/metas/+page.svelte`:
  - [x] Cargar datos del evaluado (`/employees/{id}`) y sus objetivos (`loadForEmployee(id)`).
  - [x] Renderizar las metas, categorías, avances y KPIs del colaborador.
  - [x] Incluir breadcrumb / botón de retorno hacia `/mis-evaluados`.

---

## Task 6: Verificación y Pruebas

- [x] Ejecutar tests de backend en `api/` (`go test ./...` o `go test ./internal/handler/goal/...`).
- [x] Ejecutar validaciones de frontend en `web/` (`pnpm check`).
