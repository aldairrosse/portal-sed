# Proposal: org-hierarchy-api

## Intent

Implementar el **REST API** para la gestión de la jerarquía organizacional dual (corporativa y retail), empleados, nodos del árbol, cálculo de alcance de evaluadores y consultas de "mis evaluados". Este es el backbone que determina quién evalúa a quién en todo el sistema SED.

**Spec de referencia:** `org-hierarchy` (B4) — define el modelo de datos y reglas de negocio para la jerarquía organizacional.

**Decisiones reflejadas:** #2 (catálogo único de pilares/competencias — la jerarquía define quién evalúa, no el catálogo), #8 (jefes/directores/gerentes pueden ver y solicitar cambios en metas de subordinados, pero no crear/editar/eliminar metas ajenas).

## Scope

### In Scope

- **GET org trees** — listar árboles organizacionales (corporativo, retail, ambos)
- **GET org tree structure** — listado jerárquico/nested de un árbol completo
- **CRUD OrgNodes** — crear, actualizar, reasignar nodos del árbol
- **GET employees** — listado de empleados filtrable por árbol, nodo, perfil, activo
- **GET employee detail** — detalle de un empleado con su posición en jerarquía
- **GET "my evaluatees"** — subordinados directos de un evaluador dado
- **GET evaluator scope** — alcance de evaluación para un empleado + ciclo
- **Employee search** — búsqueda por nombre, email, número de empleado
- **GET chain of command** — cadena de mando desde empleado hasta raíz
- **OpenAPI 3.1** — contrato completo como fuente de verdad

### Out of Scope

- **Autenticación / SSO** — login, sesiones, tokens son C7
- **Gestión de metas (C4)** — fijación, edición, pesos de metas
- **Asignación de competencias (C3)** — perfiles, pilares, escalas
- **Importación masiva de empleados** — bulk upload desde Excel/CSV
- **Editor de árbol org UI** — la jerarquía es dato maestro de RH; no se edita desde UI
- **Multi-tenant** — una sola organización por ahora
- **Notificaciones** — envío de emails/push cuando cambia la jerarquía
- **Audit log** — historial de cambios en la jerarquía (C7/C5 futuro)

## Concurrency & High-Volume Design

La jerarquía organizacional es consultada por **cada API del sistema** — determina quién evalúa a quién, quién puede ver qué metas, y qué pantallas se muestran. Debe diseñarse para miles a millones de requests.

### Database Connection Management

- **pgx pool**: `MaxConns=30`, `MinConns=10` — alto volumen de lectura; cada API depende de jerarquía
- **Read replicas** para todas las operaciones de lectura (traversal de árbol es 95% read)
- **Ent client per-request** — inyectado en contexto de Chi; no client global
- **Connection timeout**: 2s para obtener conexión del pool; error 503 si se agota

### Concurrency Control

- **Optimistic locking** en updates de `OrgNode` — campo `version` (integer) con `WHERE version = N`; error `409 CONFLICT` si stale
- **Advisory locks** (PostgreSQL `pg_advisory_lock`) en operaciones de reestructura de árbol — prevenir nodos huérfanos durante movimiento de sub-árbol
- **Employee transfers** envueltos en transacción — actualizar `manager_id` + `org_node_id` atómicamente; rollback si falla cualquier paso
- **Serializable** solo para reestructura de árbol; **Read Committed** para todas las lecturas

### High-Volume Patterns

- **Materialized path** (`path` column `ltree` o `VARCHAR` con delimitadores) en `OrgNode` — evitar CTEs recursivas en árboles profundos (>4 niveles)
  - Ejemplo: `path = '1.3.7.12'` para nodo con ancestros [1, 3, 7, 12]
  - Índice GIST/GIN en `path` para búsquedas de ancestros/descendientes
- **Redis cache** para árbol completo — TTL 1h; invalidación solo en cambios estructurales (rare)
- **ETag** en respuestas de estructura de árbol (`ETag: "tree-{treeId}-{lastModified}"`) — 304 Not Modified si no cambió
- **Cursor pagination** para listados de empleados — `cursor` basado en `id` + `created_at`; no `OFFSET`
- **Batch employee lookup**: `POST /api/v1/employees/batch` — resolver múltiples IDs en una sola query (IN clause)
- **Pre-computed "my evaluatees"** — materialized view `mv_evaluator_scope` o consulta cacheada en Redis por `(evaluator_id, cycle_id)` con TTL 24h
- **Rate limiting** — 100 req/s escrituras, 2000 req/s lecturas (hierarchy es read everywhere)
- **Context timeouts** — 3s para lecturas simples, 10s para traversal de árbol, 5s para queries de evaluatees
- **Streaming** para exportación masiva de árbol org — `Transfer-Encoding: chunked` en `GET /api/v1/org-trees/:treeId/export`

### Transaction Design

- **Tree restructuring** en transacción con `DEFERRABLE INITIALLY DEFERRED` en FK checks — permite mover sub-árbol sin violar FKs intermedios
- **Employee transfer** en transacción: (1) update `org_node_id`, (2) update `manager_id`, (3) invalidate caches
- **Eval scope computation** es read-only — no requiere transacción; usa índice compuesto `(evaluator_id, cycle_id)`
- **Batch reads** sin transacción — cada query independiente; fallo parcial no afecta otras

### Tree Traversal Optimization

- **ltree** extension de PostgreSQL para materialized path — eficiente para ancestros, descendientes, profundidad
- **Closure table** alternativa si `ltree` no está disponible — tabla `OrgNodeClosure` con `(ancestor_id, descendant_id, depth)`; índice compuesto en ambas columnas
- **Evitar recursive CTEs** para árboles > 1000 nodos o > 4 niveles de profundidad
- **Cache de parent chains** en Redis para hot paths — cadena de mando de empleados frecuentemente consultados (ejecutivos, RH)
- **Lazy loading** de sub-árbol — `GET /api/v1/org-nodes/:nodeId?depth=1` vs `depth=all`

## API Endpoints

### Organization Trees

| Method | Endpoint | Description | Auth (C7) |
|--------|----------|-------------|-----------|
| GET | `/api/v1/org-trees` | Listar árboles (corporativo, retail) | — |
| GET | `/api/v1/org-trees/:treeId` | Detalle de un árbol | — |
| GET | `/api/v1/org-trees/:treeId/nodes` | Todos los nodos del árbol (flat o nested) | — |
| GET | `/api/v1/org-trees/:treeId/export` | Exportar árbol completo (streaming JSON) | — |

### Org Nodes

| Method | Endpoint | Description | Auth (C7) |
|--------|----------|-------------|-----------|
| GET | `/api/v1/org-nodes/:nodeId` | Detalle de nodo con hijos directos | — |
| POST | `/api/v1/org-nodes` | Crear nodo (hoja o intermedio) | — |
| PUT | `/api/v1/org-nodes/:nodeId` | Actualizar nodo (reassign, title, profile) | — |
| DELETE | `/api/v1/org-nodes/:nodeId` | Eliminar nodo (solo si no tiene hijos) | — |
| POST | `/api/v1/org-nodes/:nodeId/move` | Mover nodo a nuevo padre | — |

### Employees

| Method | Endpoint | Description | Auth (C7) |
|--------|----------|-------------|-----------|
| GET | `/api/v1/employees` | Listar empleados (filter: tree, node, profile, active) | — |
| GET | `/api/v1/employees/:empId` | Detalle de empleado | — |
| GET | `/api/v1/employees/:empId/evaluatees` | Mis evaluados (subordinados directos) | — |
| GET | `/api/v1/employees/:empId/manager` | Mi manager (nodo padre directo) | — |
| GET | `/api/v1/employees/:empId/ancestors` | Cadena de mando hasta raíz | — |
| POST | `/api/v1/employees/batch` | Resolver múltiples IDs en batch | — |
| GET | `/api/v1/employees/search?q=...` | Búsqueda por nombre, email, número | — |

### Evaluator Scopes

| Method | Endpoint | Description | Auth (C7) |
|--------|----------|-------------|-----------|
| GET | `/api/v1/evaluator-scopes?evaluatorId=...&cycleId=...` | Alcance de evaluación para evaluador + ciclo | — |
| GET | `/api/v1/evaluator-scopes/:scopeId` | Detalle de un scope | — |

### Query Parameters (Employees)

| Parameter | Type | Description |
|-----------|------|-------------|
| `treeId` | UUID | Filtrar por árbol organizacional |
| `nodeId` | UUID | Filtrar por nodo específico |
| `profileId` | UUID | Filtrar por perfil de evaluación |
| `isActive` | boolean | Solo empleados activos |
| `q` | string | Búsqueda por nombre, email, employeeNumber |
| `cursor` | string | Paginación por cursor |
| `limit` | integer | Máximo resultados (default 50, max 200) |
| `depth` | integer | Profundidad de hijos en nested (1, 2, all) |

### Response Formats

- **Flat**: array de nodos con `parentId` explícito — eficiente para tablas
- **Nested**: árbol anidado con `children: []` — útil para renderizado de árbol UI
- **Select via header**: `Accept: application/vnd.sed.flat+json` vs `application/vnd.sed.nested+json`

## Error Handling

Formato estándar de error para todos los endpoints:

```json
{
  "code": "TREE_NOT_FOUND",
  "message": "Árbol organizacional no encontrado",
  "details": ["treeId: 550e8400-e29b-41d4-a716-446655440000"],
  "trace_id": "abc123-def456"
}
```

### Códigos de Error

| Code | HTTP | Cuándo ocurre |
|------|------|---------------|
| `TREE_NOT_FOUND` | 404 | Árbol no existe |
| `NODE_NOT_FOUND` | 404 | Nodo no existe en el árbol |
| `EMPLOYEE_NOT_FOUND` | 404 | Empleado no existe |
| `NODE_HAS_CHILDREN` | 409 | No se puede eliminar nodo con hijos |
| `INVALID_PARENT` | 400 | Padre crea ciclo en el árbol (A → B → A) |
| `UNAUTHORIZED_SCOPE` | 403 | Evaluador no tiene permiso para ese scope |
| `STALE_VERSION` | 409 | Optimistic locking falló (versión desactualizada) |
| `INVALID_TREE_TYPE` | 400 | Tipo de árbol no es `corporativa` ni `retail` |
| `EMPLOYEE_IN_MULTIPLE_TREES` | 409 | Empleado ya está en otro árbol |
| `RATE_LIMIT_EXCEEDED` | 429 | Demasiados requests |

## Non-Goals

- No implementar autenticación ni autorización (C7)
- No gestionar metas, competencias, ni ciclos de evaluación (C3, C4, C2)
- No importar empleados masivamente desde Excel/CSV
- No proveer UI de editor de árbol org (la jerarquía es dato maestro)
- No soportar multi-tenant (una sola org por ahora)
- No calcular cascada de evaluación (solo evaluación directa, no indirecta)
- No registrar audit log de cambios en jerarquía
- No soportar soft delete en nodos (hard delete con restricción de hijos)

## Dependencies

- **data-model-core (C1)** — esquema Ent con `Organization`, `OrgNode`, `Employee`, `EvaluatorScope`, índices, y relaciones
- **org-hierarchy spec (B4)** — fuente de verdad de reglas de negocio, perfiles, árbol dual
- **principles/architecture.md** — estructura de paquetes `internal/handler`, `internal/service`, `internal/repository`
- **principles/api-design.md** — convenciones REST, paginación, errores, ETag
- **Go 1.22+**, `chi`, `ent`, `pgx`, `redis` (opcional para cache)

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/handler/orgtree.go` | New | Handlers para árboles org |
| `api/internal/handler/orgnode.go` | New | Handlers para nodos org |
| `api/internal/handler/employee.go` | New | Handlers para empleados y evaluatees |
| `api/internal/handler/evaluator.go` | New | Handlers para scopes de evaluador |
| `api/internal/service/orgtree.go` | New | Lógica de negocio de árboles |
| `api/internal/service/orgnode.go` | New | Lógica de negocio de nodos |
| `api/internal/service/employee.go` | New | Lógica de empleados y búsqueda |
| `api/internal/service/evaluator.go` | New | Cálculo de scope de evaluador |
| `api/internal/repository/orgtree.go` | New | Queries de árbol con materialized path |
| `api/internal/repository/employee.go` | New | Queries de empleados con filtros |
| `api/internal/repository/evaluator.go` | New | Queries de scope de evaluación |
| `api/internal/cache/` | New | Capa de cache Redis para árboles y evaluatees |
| `api/internal/middleware/` | Modified | Rate limiting, ETag, context timeout |
| `api/openapi/` | New | OpenAPI 3.1 spec para todos los endpoints |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Materialized path se desincroniza con relaciones | Medium | Trigger en DB o lógica en service para mantener `path` actualizado; validación en CI |
| CTEs recursivas lentas en árboles grandes | Medium | Usar `ltree` o closure table; benchmark con 10k nodos antes de deploy |
| Cache Redis stale tras movimiento de nodo | Medium | Invalidación explícita en service de `OrgNode.Move`; transacción DB + cache con retry |
| Race condition en reasignación de empleado | Low | Transacción serializable + advisory lock; optimistic locking en `OrgNode` |
| Índice `(manager_id, is_active)` insuficiente para "my evaluatees" | Low | Materialized view `mv_evaluator_scope`; benchmark query antes de release |
| Rate limiting afecta lecturas legítimas | Low | Límites generosos en reads (2000/s); 429 con `Retry-After` header |

## Rollback Plan

1. **Revertir código** — rollback de handlers, services, repos a versión anterior
2. **Migración de DB** — si se agregó `ltree` o closure table, no es reversible sin pérdida; planificar migración en mantenimiento
3. **Cache** — flush de Redis keys `org:*` y `evaluator:*`
4. **Si en producción**: PITR de PostgreSQL si hay corrupción de datos en árbol

## Success Criteria

- [ ] `GET /api/v1/org-trees/:treeId/nodes` retorna en < 50ms para árboles de hasta 10,000 nodos
- [ ] `GET /api/v1/employees/:empId/evaluatees` retorna en < 30ms con cache, < 100ms sin cache
- [ ] Load test: 2000 req/s sostenidos en `GET /api/v1/employees` sin degradación
- [ ] ETag + 304 Not Modified reduce carga de DB en 80% para lecturas de árbol
- [ ] Creación de nodo con padre inválido (ciclo) retorna `INVALID_PARENT` 400
- [ ] Eliminación de nodo con hijos retorna `NODE_HAS_CHILDREN` 409
- [ ] Optimistic locking funciona: dos updates concurrentes al mismo nodo, uno gana, otro recibe 409
- [ ] Búsqueda de empleados por nombre/email retorna resultados relevantes en < 50ms
- [ ] Batch lookup de 100 empleados resuelve en < 50ms
- [ ] OpenAPI 3.1 valida sin errores con `openapi-generator` o `swagger-cli`
- [ ] `openspec validate --all` pasa
