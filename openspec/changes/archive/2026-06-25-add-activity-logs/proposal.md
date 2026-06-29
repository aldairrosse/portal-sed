## Why

La vista `/perfil` muestra un timeline de "Actividad reciente" usando datos fixture (`activity-logs.json`). El filtro actual usa `profileId` (un slug de texto como "colaborador" o "jefe"), no un identificador real de empleado. Esto impide que la actividad sea persistente, consultable por empleado específico, o auditable. Necesitamos reemplazar el fixture con registros reales persistidos en base de datos, un endpoint para consultarlos, y una utility function en el backend para registrar eventos de actividad desde cualquier capa del dominio.

## What Changes

- Agregar tabla `activity_logs` en PostgreSQL con migración `000006_add_activity_logs`
- Agregar schema Ent `activitylog.go` en `api/internal/schema/`
- Agregar utility function `LogActivity(ctx, employeeID, action, description, module, metadata)` en `api/internal/service/activity/`
- Agregar endpoint `GET /api/v1/employees/{employeeId}/activity-logs` (últimos 50 registros, ordenados por `created_at DESC`)
- Agregar endpoint al spec OpenAPI `activity-logs.yaml`
- Agregar store/request en frontend para consumir el endpoint desde `/perfil`
- Reemplazar el import del fixture en `web/src/routes/perfil/+page.svelte` por la llamada API real
- Cambiar el filtro de `profileId` (string slug) a `employeeId` (UUID)

## Capabilities

### New Capabilities

- `activity-logs`: Registro, persistencia y consulta de eventos de actividad del empleado. Incluye tabla BD, schema Ent, utility de registro, endpoint REST y consumo frontend.

### Modified Capabilities

- `ui-shell`: La vista `/perfil` deja de usar fixture y consume API real para el timeline de actividad.

## Impact

**Backend:**
- `api/migrations/000006_add_activity_logs.up.sql` / `.down.sql`: Nueva tabla `activity_logs` (UUID PK, FK a `employees`, índice compuesto `(employee_id, created_at DESC)`)
- `api/internal/schema/activitylog.go`: Schema Ent con campos `id`, `employee_id`, `action`, `description`, `module`, `metadata` (JSONB nullable), `created_at`
- `api/internal/service/activity/activity.go`: Utility `LogActivity()` — inserta registro; llamada desde handlers de dominio según convenga
- `api/internal/handler/activity/activity_handler.go`: Handler `GET /employees/{employeeId}/activity-logs`
- `api/internal/repository/activity/activity_repository.go`: Query con límite 50, orden `created_at DESC`, filtro por `employee_id`
- `api/openapi/activity-logs.yaml`: Spec OpenAPI del endpoint
- `cmd/server/main.go`: Registrar ruta y dependencias

**Frontend:**
- `web/src/lib/api/client.ts`: Agregar path para activity-logs
- `web/src/lib/api/schemas/activity-logs.d.ts`: Tipos generados desde OpenAPI
- `web/src/routes/perfil/+page.svelte`: Reemplazar fixture por llamada API; cambiar filtro `profileId` → `employeeId`

**Dependencias:**
- `openapi-typescript`: Ya instalado; generar tipos con `openapi-typescript api/openapi/activity-logs.yaml -o web/src/lib/api/schemas/activity-logs.d.ts`
- Ent: Ya instalado; `go generate ./api/internal/schema` para regenerar cliente

**Decisions técnicas:**
- `employee_id UUID` (no `profile_id string`): el fixture usa slugs de perfil, pero la consulta real es por empleado; `profileId` era un proxy de dev persona
- `metadata JSONB nullable`: permite extender acciones con contexto adicional (IDs de entidades relacionadas, valores anteriores) sin alterar el schema
- Índice `(employee_id, created_at DESC)`: alineado al query principal (últimos 50 de un empleado)
- Límite hardcodeado de 50: suficiente para el timeline de perfil; sin paginación en esta iteración
- Utility centralizada `LogActivity()`: un solo punto de inserción; los handlers de dominio la invocan después de operaciones exitosas

**Principles relevantes:**
- `principles/data-and-orm.md` — migraciones versionadas, índices explícitos
- `principles/contracts-api.md` — OpenAPI como fuente de verdad
- `principles/security.md` — RBAC en endpoint (empleado solo ve sus propios logs)

## Non-Goals

- Backfill retroactivo de actividad histórica (no hay datos previos que migrar)
- Paginación o cursor en activity-logs (el timeline muestra últimos 50, sin scroll infinito)
- Streaming o notificaciones en tiempo real de actividad
- Registro automático de TODAS las acciones del sistema (se registran solo las que el producto define como relevantes; el catálogo de acciones se define en la spec)
- Panel administrativo de actividad para RH o auditores (solo vista del empleado en `/perfil`)
- Exportación de logs (CSV/PDF)
