# Design: Activity Logs

## Technical Approach

Reemplazar el fixture `activity-logs.json` por registros reales persistidos en PostgreSQL. Se agrega una tabla `activity_logs` (append-only), un endpoint `GET /api/v1/employees/{employeeId}/activity-logs`, y una utility function `LogActivity()` como único punto de inserción. El frontend consume la API desde `/perfil` usando el `employeeId` de la sesión en lugar del `profileId` slug.

La implementación sigue los patrones existentes del codebase: schema Ent con `TimeMixin`, repository con raw SQL + `*sql.DB`, service con interface + DTOs, handler delgado con `writeJSON`/`writeError`, y store Svelte con `$state` triplete (data/loading/error).

## Architecture Decisions

| Decisión | Opción | Alternativa | Rationale |
|----------|--------|-------------|-----------|
| PK type | `UUID DEFAULT gen_random_uuid()` | `BIGSERIAL` | Consistencia con todas las tablas existentes (000001) |
| Mixin | `TimeMixin` (created_at only) | `AuditMixin` (created_by/updated_by) | Activity log es append-only; no tiene sentido actualizar. `updated_at` sería dead weight. Se usa `TimeMixin` pero solo se necesita `created_at` — el campo `updated_at` del mixin queda pero nunca se usa. |
| `action` field type | `TEXT` (no enum) | `ENUM` | Spec requiere que nuevas acciones no requieran migración. TEXT permite extensión sin ALTER TYPE. |
| `metadata` field | `JSONB NULL` | Columna separada por acción | Flexibilidad para contexto variable sin schema changes. Nullable para acciones que no necesitan metadata. |
| Index | `(employee_id, created_at DESC)` | `(created_at DESC)` solo | El query principal siempre filtra por empleado. El índice compuesto cubre el 100% de los queries del timeline. |
| Repository pattern | Raw SQL con `*sql.DB` | Ent client queries | Consistencia con `cycle_repo.go` y `org` repos. Raw SQL da control explícito del ORDER BY y LIMIT. |
| Authorization | Service layer compara `employeeID` con `auth.GetEmployeeID(ctx)` | Middleware dedicado | Un solo endpoint; no justifica un middleware. El service ya tiene acceso al contexto auth. |
| Límite de resultados | 50 hardcoded, max 100 via query param | Paginación por cursor | Spec dice "últimos 50, sin scroll infinito". Query param `limit` opcional para flexibilidad futura. |
| Frontend store | Module-level `$state` + getters (cycleStore pattern) | Svelte class | Consistencia con stores existentes. Simpler, same pattern. |

## Data Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        BACKEND                                   │
│                                                                  │
│  Domain Handler ──→ activity.LogActivity(ctx, empID, ...)       │
│        │                      │                                  │
│        │               ActivityService.Create()                  │
│        │                      │                                  │
│        │               ActivityRepo.Create()                     │
│        │                      │                                  │
│        │               INSERT INTO activity_logs                 │
│        │                                                         │
│  GET /employees/{id}/activity-logs                              │
│        │                                                         │
│  ActivityHandler.List() ──→ ActivityService.ListByEmployee()    │
│                                  │                               │
│                           ActivityRepo.ListByEmployee()          │
│                                  │                               │
│                           SELECT ... ORDER BY created_at DESC    │
│                                  LIMIT 50                        │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼ JSON response
┌─────────────────────────────────────────────────────────────────┐
│                        FRONTEND                                  │
│                                                                  │
│  +page.svelte (perfil)                                          │
│        │                                                         │
│  onMount → activityLogStore.loadActivityLogs(employeeId)        │
│        │                                                         │
│  client.GET('/employees/{id}/activity-logs')                    │
│        │                                                         │
│  $state logs[] → timeline render                                │
└─────────────────────────────────────────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `api/migrations/000006_add_activity_logs.up.sql` | Create | DDL: tabla `activity_logs` + índice compuesto |
| `api/migrations/000006_add_activity_logs.down.sql` | Create | DROP TABLE + DROP INDEX |
| `api/internal/schema/activitylog.go` | Create | Ent schema: fields, edges, indexes |
| `api/internal/repository/activity/activity_repo.go` | Create | Repository: `Create()`, `ListByEmployee()` con raw SQL |
| `api/internal/service/activity/activity_service.go` | Create | Service: `LogActivity()`, `ListByEmployee()`, auth check |
| `api/internal/handler/activity/activity_handler.go` | Create | Handler: `ListActivityLogs()` — GET endpoint |
| `api/internal/handler/activity/routes.go` | Create | `RegisterRoutes()` con RequireAuth + rate limit |
| `api/cmd/server/main.go` | Modify | Wire activity repo/service/handler; registrar en apiV1 |
| `api/openapi/activity-logs.yaml` | Create | OpenAPI 3.1 spec del endpoint |
| `web/src/lib/api/schemas/activity-logs.d.ts` | Create | Tipos TS generados desde OpenAPI |
| `web/src/lib/api/client.ts` | Modify | Agregar `ActivityLogsPaths` al type intersection |
| `web/src/lib/stores/activityLogStore.svelte.ts` | Create | Store: `$state` logs/loading/error + `loadActivityLogs()` |
| `web/src/routes/perfil/+page.svelte` | Modify | Reemplazar fixture por store; `profileId` → `employeeId` |

## Database Design

### Migration: `000006_add_activity_logs.up.sql`

```sql
-- +goose Up
-- +goose StatementBegin

CREATE TABLE activity_logs (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    employee_id UUID        NOT NULL,
    action      TEXT        NOT NULL,
    description TEXT        NOT NULL,
    module      TEXT        NOT NULL,
    metadata    JSONB       NULL,

    CONSTRAINT fk_activity_logs_employee
        FOREIGN KEY (employee_id) REFERENCES employees(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_activity_logs_emp_created
    ON activity_logs (employee_id, created_at DESC);

-- +goose StatementEnd
```

### Migration: `000006_add_activity_logs.down.sql`

```sql
-- +goose Up
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_activity_logs_emp_created;
DROP TABLE IF EXISTS activity_logs;

-- +goose StatementEnd
```

### Ent Schema: `api/internal/schema/activitylog.go`

```go
package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/edge"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
    "github.com/google/uuid"
)

type ActivityLog struct {
    ent.Schema
}

func (ActivityLog) Mixin() []ent.Mixin {
    return []ent.Mixin{TimeMixin{}}
}

func (ActivityLog) Fields() []ent.Field {
    return []ent.Field{
        field.UUID("id", uuid.UUID{}).
            Default(uuid.New).
            StorageKey("id"),
        field.UUID("employee_id", uuid.UUID{}),
        field.String("action"),
        field.String("description"),
        field.String("module"),
        field.JSON("metadata", map[string]interface{}).
            Optional().
            Nillable(),
    }
}

func (ActivityLog) Edges() []ent.Edge {
    return []ent.Edge{
        edge.From("employee", Employee.Type).
            Ref("activity_logs").
            Unique().
            Required().
            Field("employee_id"),
    }
}

func (ActivityLog) Index() []ent.Index {
    return []ent.Index{
        index.Fields("employee_id", "created_at"),
    }
}
```

**Nota:** Se debe agregar `edge.To("activity_logs", ActivityLog.Type)` en `employee.go`.

## Backend Design

### Repository: `api/internal/repository/activity/activity_repo.go`

```go
package activity

// ActivityRow maps to the activity_logs table.
type ActivityRow struct {
    ID          uuid.UUID
    CreatedAt   time.Time
    EmployeeID  uuid.UUID
    Action      string
    Description string
    Module      string
    Metadata    []byte // raw JSONB, nil si NULL
}

type ActivityRepo struct {
    client *internal.Client
    db     *sql.DB
}

func NewActivityRepo(client *internal.Client, db *sql.DB) *ActivityRepo

// Create inserts a new activity log. Returns the created row.
func (r *ActivityRepo) Create(ctx context.Context, employeeID uuid.UUID,
    action, description, module string, metadata []byte) (*ActivityRow, error)

// ListByEmployee returns the latest logs for an employee, ordered by
// created_at DESC, limited to `limit` rows.
func (r *ActivityRepo) ListByEmployee(ctx context.Context,
    employeeID uuid.UUID, limit int) ([]*ActivityRow, error)
```

Query strategy: `SELECT ... FROM activity_logs WHERE employee_id = $1 ORDER BY created_at DESC LIMIT $2` — cubierto por el índice `idx_activity_logs_emp_created`.

### Service: `api/internal/service/activity/activity_service.go`

```go
package activity

type Service interface {
    LogActivity(ctx context.Context, employeeID uuid.UUID,
        action, description, module string, metadata map[string]interface{}) error
    ListByEmployee(ctx context.Context, employeeID uuid.UUID,
        limit int) ([]*ActivityResponse, error)
}

type ActivityResponse struct {
    ID          string                 `json:"id"`
    Action      string                 `json:"action"`
    Description string                 `json:"description"`
    Module      string                 `json:"module"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   string                 `json:"created_at"`
}
```

`ListByEmployee` verifica autorización: compara `employeeID` con `auth.GetEmployeeID(ctx)`. Si no coincide, retorna `pkgerrors.ErrForbidden` (403).

### Handler: `api/internal/handler/activity/activity_handler.go`

```go
package activity

type ActivityHandler struct {
    svc activity.Service
}

// ListActivityLogs handles GET /api/v1/employees/{employeeId}/activity-logs
func (h *ActivityHandler) ListActivityLogs(w http.ResponseWriter, r *http.Request)
```

- Parsea `employeeId` de `chi.URLParam` — valida UUID (400 si inválido)
- Parsea `limit` query param (default 50, clamp 1-100)
- Delega a `svc.ListByEmployee` (que hace el auth check)
- Responde con `{ "data": [...], "total": N }`

### Routes: `api/internal/handler/activity/routes.go`

```go
func RegisterRoutes(r chi.Router, handler *ActivityHandler, authSvc *authsvc.AuthService) {
    readRateLimit := middleware.RateLimitConfig{
        Window: time.Minute, MaxCount: 2000,
        Store: middleware.NewInMemoryRateLimitStore(),
    }

    r.Group(func(r chi.Router) {
        r.Use(middleware.RequireAuth(authSvc))
        r.Group(func(r chi.Router) {
            r.Use(middleware.RateLimit(readRateLimit))
            r.Get("/employees/{employeeId}/activity-logs", handler.ListActivityLogs)
        })
    })
}
```

### Wiring: `cmd/server/main.go`

Agregar en la sección de DI:

```go
// Activity
activityRepo := repoactivity.NewActivityRepo(client, db)
activitySvc := activitysvc.NewService(activityRepo)
activityH := activityhandler.NewActivityHandler(activitySvc)
```

Registrar en `apiV1`:

```go
activityhandler.RegisterRoutes(apiV1, activityH, authSvc)
```

## OpenAPI Spec: `api/openapi/activity-logs.yaml`

```yaml
openapi: 3.1.0
info:
  title: SED Activity Logs API
  version: 1.0.0
servers:
  - url: /api/v1
paths:
  /employees/{employeeId}/activity-logs:
    get:
      operationId: listActivityLogs
      summary: List activity logs for an employee
      parameters:
        - name: employeeId
          in: path
          required: true
          schema: { type: string, format: uuid }
        - name: limit
          in: query
          required: false
          schema: { type: integer, default: 50, minimum: 1, maximum: 100 }
      responses:
        "200":
          description: Activity logs list
          content:
            application/json:
              schema:
                type: object
                properties:
                  data:
                    type: array
                    items:
                      $ref: "#/components/schemas/ActivityLog"
                  total:
                    type: integer
        "400": { description: Bad Request }
        "403": { description: Forbidden }
components:
  schemas:
    ActivityLog:
      type: object
      required: [id, action, description, module, created_at]
      properties:
        id: { type: string, format: uuid }
        action: { type: string }
        description: { type: string }
        module: { type: string }
        metadata: { type: object, nullable: true, additionalProperties: true }
        created_at: { type: string, format: date-time }
```

## Frontend Design

### API Client Update: `web/src/lib/api/client.ts`

Agregar import y type intersection:

```ts
import type { paths as ActivityLogsPaths } from './schemas/activity-logs.d.ts';
type AppPaths = AuthPaths & CyclePaths & GoalsPaths & CompetencyPaths
    & OrgHierarchyPaths & ActivityLogsPaths;
```

### Store: `web/src/lib/stores/activityLogStore.svelte.ts`

```ts
import { client } from '$lib/api/client';

interface ActivityLog {
    id: string;
    action: string;
    description: string;
    module: string;
    metadata?: Record<string, unknown> | null;
    created_at: string;
}

let logs = $state<ActivityLog[]>([]);
let loading = $state(false);
let error = $state<string | null>(null);

export function getActivityLogs(): ActivityLog[] { return logs; }
export function isLoading(): boolean { return loading; }
export function getError(): string | null { return error; }

export async function loadActivityLogs(
    employeeId: string, limit = 50
): Promise<void> {
    loading = true;
    error = null;
    try {
        const { data, error: apiError } = await client.GET(
            '/employees/{employeeId}/activity-logs',
            { params: { path: { employeeId }, query: { limit } } }
        );
        if (apiError) throw new Error('Error al cargar actividad');
        const raw = data as { data?: ActivityLog[] };
        logs = raw?.data ?? [];
    } catch (e) {
        error = e instanceof Error ? e.message : 'Error al cargar actividad';
        logs = [];
    } finally {
        loading = false;
    }
}

export function clearLogs(): void { logs = []; }
```

### Component Update: `web/src/routes/perfil/+page.svelte`

Cambios:
1. **Remover:** `import activityLogs from "$lib/fixtures/activity/activity-logs.json";`
2. **Agregar:** `import { loadActivityLogs, getActivityLogs, isLoading as logsLoading } from "$lib/stores/activityLogStore.svelte";`
3. **Agregar:** `import { onMount } from "svelte";`
4. **Reemplazar** `filteredLogs` derived con:
   ```ts
   const logs = $derived(getActivityLogs());
   onMount(() => { if (user?.employeeId) loadActivityLogs(user.employeeId); });
   ```
5. **En el template:** cambiar `filteredLogs` → `logs`, `log.timestamp` → `log.created_at`

## Security & Authorization

| Layer | Mechanism | Detail |
|-------|-----------|--------|
| Transport | Session cookie / Bearer token | `RequireAuth` middleware valida sesión |
| Handler | UUID validation | `employeeId` path param debe ser UUID válido (400 si no) |
| Service | Identity comparison | `auth.GetEmployeeID(ctx)` vs `employeeId` param → 403 si no coincide |
| Database | FK constraint | `employee_id REFERENCES employees(id) ON DELETE CASCADE` |

No hay RBAC por rol — la autorización es ownership: cada empleado solo ve sus propios logs.

## Testing Strategy

| Layer | What | Approach |
|-------|------|----------|
| Unit (Go) | `ActivityRepo.ListByEmployee` returns correct rows, respects limit | Table test con DB de test |
| Unit (Go) | `ActivityService.ListByEmployee` rejects mismatched employeeID | Assert 403 error |
| Integration | `GET /employees/{id}/activity-logs` end-to-end | HTTP test con test server |
| Frontend | Store loads logs, handles error state | Vitest + mock client |
| E2E | `/perfil` shows activity timeline from API | Playwright (si existe infra) |

## Migration / Rollout

1. Migración `000006` crea tabla e índice — zero-downtime (CREATE TABLE + CREATE INDEX son non-blocking en PostgreSQL con `CONCURRENTLY` si se requiere en producción).
2. Ent auto-migration en dev crea la tabla automáticamente al arrancar.
3. No hay backfill necesario — la tabla empieza vacía.
4. El frontend funciona con array vacío mientras no haya logs registrados.
5. `LogActivity()` se invoca desde handlers de dominio en cambios futuros (no en este change).

## Open Questions

- None.
