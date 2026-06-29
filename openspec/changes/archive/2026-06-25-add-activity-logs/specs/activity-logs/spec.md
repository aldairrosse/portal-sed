# Delta for Activity Logs

## Purpose

Registro, persistencia y consulta de eventos de actividad del empleado en base de datos. Sustituye el fixture `activity-logs.json` por logs reales con endpoint REST y visualización en `/perfil`.

## ADDED Requirements

### Requirement: Registro de actividad

El sistema SHALL persistir eventos de actividad cuando operaciones del dominio se completen exitosamente. Cada registro SHALL incluir: `employee_id` (UUID), `action` (string), `description` (string), `module` (string), `metadata` (JSONB nullable), `created_at` (timestamp). La utilidad `LogActivity(ctx, employeeID, action, description, module)` SHALL ser el único punto de inserción.

#### Scenario: Registro tras operación exitosa

- **WHEN** se completa una operación del dominio (e.g., `evaluation_started`, `goal_approved`, `kpi_updated`)
- **THEN** se invoca `LogActivity(ctx, employeeID, action, description, module)`
- **AND** el registro se persiste en la tabla `activity_logs`
- **AND** `metadata` puede contener contexto adicional como IDs de entidades

#### Scenario: Metadata opcional

- **WHEN** `LogActivity` se invoca sin metadata
- **THEN** el campo `metadata` queda `NULL`
- **AND** el registro se persiste sin error

### Requirement: Consulta de actividad por empleado

El sistema SHALL exponer `GET /api/v1/employees/{employeeId}/activity-logs`. El endpoint SHALL retornar los últimos 50 registros, ordenados por `created_at DESC`, usando el índice `(employee_id, created_at DESC)`.

#### Scenario: Consulta exitosa

- **WHEN** un empleado autenticado solicita `GET /api/v1/employees/{employeeId}/activity-logs`
- **THEN** el sistema retorna `200 OK` con array de los últimos 50 registros
- **AND** cada registro incluye `id`, `action`, `description`, `module`, `created_at`
- **AND** el array está ordenado por `created_at` descendente

#### Scenario: Empleado sin actividad

- **WHEN** el empleado no tiene registros en `activity_logs`
- **THEN** el sistema retorna `200 OK` con array vacío `[]`

#### Scenario: employeeId inválido

- **WHEN** el parámetro `employeeId` no es un UUID válido
- **THEN** el sistema retorna `400 Bad Request` con mensaje de validación

### Requirement: Autorización de consulta

El sistema SHALL restringir que un empleado solo pueda consultar sus propios logs de actividad. La autorización SHALL verificarse en la capa de servicio.

#### Scenario: Acceso a logs propios

- **WHEN** el `employeeId` de la URL coincide con la identidad del usuario autenticado
- **THEN** el sistema retorna `200 OK` con los registros

#### Scenario: Acceso a logs de otro empleado denegado

- **WHEN** el `employeeId` de la URL no coincide con la identidad del usuario autenticado
- **THEN** el sistema retorna `403 Forbidden`

### Requirement: Catálogo de acciones

El sistema SHALL documentar 16 tipos de acción como referencia de consistencia: `evaluation_started`, `evaluation_completed`, `goal_viewed`, `goal_approved`, `goal_progress`, `comment_added`, `login`, `profile_viewed`, `competency_viewed`, `hierarchy_viewed`, `evaluation_reviewed`, `cycle_configured`, `pillar_edited`, `scale_modified`, `report_exported`, `goal_overview`. El campo `action` es `TEXT` — no enum — por lo que nuevas acciones se aceptan sin migración de schema.

#### Scenario: Nueva acción sin migración

- **WHEN** se registra una acción no listada en el catálogo anterior
- **THEN** la acción se persiste sin alterar el schema
- **AND** el frontend muestra un ícono por defecto para acciones desconocidas

### Requirement: Visualización del timeline en /perfil

El frontend SHALL consumir `GET /api/v1/employees/{currentUserId}/activity-logs` en la página `/perfil` para mostrar el timeline de actividad. El fixture `activity-logs.json` SHALL ser reemplazado por la llamada a la API real.

#### Scenario: Carga del timeline desde API

- **WHEN** un usuario visita `/perfil`
- **THEN** la página llama `GET /api/v1/employees/{currentUserId}/activity-logs`
- **AND** muestra los registros en un timeline ordenado por `created_at DESC`

#### Scenario: Error de red

- **WHEN** el endpoint de activity-logs retorna error de red
- **THEN** el timeline muestra el mensaje "No se pudo cargar la actividad reciente"
- **AND** el resto de la página `/perfil` carga normalmente

## Acceptance Criteria

1. Tabla `activity_logs` con índice `(employee_id, created_at DESC)` creada por migración `000006`
2. `LogActivity()` persiste registros con los campos `employee_id`, `action`, `description`, `module`, `metadata`, `created_at`
3. `GET /api/v1/employees/{employeeId}/activity-logs` retorna máx. 50 registros, ordenados por `created_at DESC`
4. Consulta de logs de otro empleado retorna `403 Forbidden`
5. Catálogo documenta 16 acciones; nuevas acciones no requieren migración
6. `/perfil` consume API real y reemplaza el fixture
7. `pnpm run check` y `go test ./...` pasan
