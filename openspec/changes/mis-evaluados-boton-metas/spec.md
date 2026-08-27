# Delta Spec: Mis-Evaluados Status, Validación de Pesos y Navegación de Metas por Época

## Capability: goal-assignment-status-tracking

| Field | Detail |
|-------|--------|
| **Purpose** | Controlar el ciclo de vida de la formulación y envío de metas individuales mediante estados formales (`borrador`, `enviada`) y sellado de fecha de envío (`submitted_at`), validando la regla de ponderación Double 100%. |
| **Depends on** | `goals-and-weighting`, `evaluation-lifecycle` |

### REQ-MEM-001: Persistencia de status y submitted_at en GoalAssignment

La tabla `goal_assignments` y el esquema Ent `GoalAssignment` DEBEN almacenar los campos:
- `status`: enum (`borrador`, `enviada`), con valor por defecto `'borrador'`.
- `submitted_at`: timestamp con zona horaria (`TIMESTAMPTZ`), nulo inicialmente hasta el envío formal.

El endpoint `GET /api/v1/employees/{empId}/assignments` DEBE retornar `status` y `submitted_at` en `AssignmentResponse`.

#### Scenario: Creación inicial en borrador
- GIVEN un empleado que inicia la formulación de metas para el ciclo activo
- WHEN se crea la asignación antes de validar o enviar
- THEN `status` es `'borrador'`
- AND `submitted_at` es `null`

---

### REQ-MEM-002: Validación de Pesos Double 100% como compuerta para status 'enviada'

Al solicitar el guardado/envío formal de la asignación (`POST /api/v1/employees/{empId}/assignments`):
1. El backend DEBE ejecutar la validación de pesos Double 100% (`ValidateDoubleWeighting`):
   - La suma de pesos de las categorías DEBE ser 100% ($\pm \epsilon$).
   - La suma de pesos de las metas cuantitativas personales dentro de cada categoría DEBE ser 100% ($\pm \epsilon$).
2. Si la validación es exitosa:
   - La asignación transiciona a `status = 'enviada'`.
   - Se registra `submitted_at = now()`.
   - Se retorna `HTTP 201/200` con `AssignmentResponse` actualizado.
3. Si la validación no es exitosa (suma $<100\%$ o $>100\%$):
   - El estado NO cambia a `'enviada'`.
   - Permanece en `'borrador'`.
   - Se rechaza el envío formal con error `400` (`WeightInvalid`).

#### Scenario: Envío exitoso con ponderación al 100%
- GIVEN categorías que suman 100% y metas en cada categoría que suman 100%
- WHEN el colaborador o jefe envía la asignación
- THEN la validación retorna `Valid: true`
- AND `GoalAssignment.status` cambia a `'enviada'`
- AND `GoalAssignment.submitted_at` se actualiza con la fecha y hora actual

#### Scenario: Bloqueo de envío con ponderación incompleta o excedida
- GIVEN metas personales cuya suma de pesos es 80% o 110%
- WHEN se intenta enviar la asignación
- THEN el backend responde `400 Bad Request` indicando déficit o exceso de peso
- AND `GoalAssignment.status` permanece en `'borrador'`
- AND `GoalAssignment.submitted_at` permanece sin cambios (`null`)

---

### REQ-MEM-003: Eliminación de creación automática de asignaciones vacías en el frontend

En `goalsStore.svelte.ts`, la función `_doLoad()` NO DEBE realizar `POST /employees/{empId}/assignments` de forma automática ni silenciosa cuando no existe una asignación previa para un subordinado. La asignación solo debe crearse por acción deliberada del usuario (al guardar o formular metas).

#### Scenario: Consulta de subordinado sin asignación previa
- GIVEN un jefe que consulta las metas de un colaborador nuevo sin asignación en el ciclo
- WHEN el store ejecuta `loadForEmployee(empId)`
- THEN NO se dispara ninguna petición POST silenciosa en segundo plano
- AND el estado local refleja que el colaborador aún no cuenta con asignación (`borrador` o stub sin persistir)

---

## Capability: contextual-goals-navigation

| Field | Detail |
|-------|--------|
| **Purpose** | Proveer acceso contextual desde la lista de evaluados del jefe hacia la formulación o evaluación de metas según la época del ciclo y el estado de la asignación. |
| **Depends on** | `mis-evaluados-ui`, `goal-assignment-status-tracking` |

### REQ-MEM-004: Selección de subordinado mediante query param ?empId= en Asignación de Objetivos

La ruta `/objetivos/asignacion` DEBE leer el parámetro de consulta `?empId={UUID}` de la URL:
- Si el usuario autenticado tiene rol de evaluador (Jefe/RH) y el `empId` corresponde a un evaluado de su equipo:
  - `selectedEmployeeId` DEBE inicializarse con dicho `empId`.
  - Se DEBE cargar el árbol de categorías y metas del colaborador vía `loadForEmployee(empId)`.

#### Scenario: Jefe navega con parámetro de colaborador
- GIVEN un jefe en la ruta `/objetivos/asignacion?empId=123e4567-e89b-12d3-a456-426614174000`
- WHEN la página se monta
- THEN el selector de empleado se establece en el colaborador con ID `123e4567-e89b-12d3-a456-426614174000`
- AND se visualizan las metas y categorías de dicho colaborador

---

### REQ-MEM-005: Columna de estado y botón Metas en tabla de Mis Evaluados

En `web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte`:
1. Se DEBE mostrar una columna indicadora de estado de metas:
   - Badge `'Enviada'` (verde) si la asignación está enviada.
   - Badge `'Borrador'` (amarillo/gris) si está en borrador o pendiente de formulación.
2. En la columna de **Acciones**, se DEBE incluir el botón **"Metas"**:
   - Si la fase activa es `'inicio-anio'` (o la meta está en `'borrador'`):
     - El enlace DEBE dirigir a `/objetivos/asignacion?empId={employee.id}`.
   - Si la fase activa es `'medio-anio'` o `'fin-anio'` (o la meta está en `'enviada'`):
     - El enlace DEBE dirigir a `/mis-evaluados/{employee.id}/metas`.

#### Scenario: Clic en botón Metas en inicio de año
- GIVEN la fase del ciclo es `inicio-anio` y el colaborador tiene metas en `borrador`
- WHEN el jefe hace clic en el botón "Metas" en la fila del colaborador
- THEN la aplicación navega a `/objetivos/asignacion?empId={colaborador.id}`

#### Scenario: Clic en botón Metas en medio año con metas enviadas
- GIVEN la fase del ciclo es `medio-anio` y el colaborador tiene metas en estado `enviada`
- WHEN el jefe hace clic en el botón "Metas" en la fila del colaborador
- THEN la aplicación navega a `/mis-evaluados/{colaborador.id}/metas`

---

## Capability: manager-evaluatee-goals-view

| Field | Detail |
|-------|--------|
| **Purpose** | Permitir al evaluador consultar el detalle y avance de las metas enviadas por su colaborador en periodos de seguimiento y evaluación final. |
| **Depends on** | `goals-api`, `contextual-goals-navigation` |

### REQ-MEM-006: Ruta dedicada /mis-evaluados/[id]/metas

Se DEBE implementar la ruta `web/src/routes/mis-evaluados/[id]/metas/+page.svelte`:
- DEBE obtener el `id` del colaborador desde `$page.params.id`.
- DEBE consultar la información del empleado (`/employees/{id}`) y cargar sus metas (`loadForEmployee(id)`).
- DEBE renderizar las categorías, metas individuales, metas globales/compartidas asociadas, unidades, valores objetivo, líneas base y progresos actuales.
- DEBE incluir un botón o breadcrumb de navegación para regresar a `/mis-evaluados`.

#### Scenario: Consulta de metas enviadas por el colaborador
- GIVEN un jefe navegando a `/mis-evaluados/{colaboradorId}/metas`
- WHEN la página carga
- THEN se muestra el nombre y puesto del colaborador
- AND se despliega la lista de categorías con sus respectivas metas y KPIs
- AND se muestra el botón de retorno hacia `/mis-evaluados`
