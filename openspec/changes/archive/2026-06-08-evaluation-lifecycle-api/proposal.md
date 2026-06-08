# C2: evaluation-lifecycle-api

## Intent

C2 entrega el **REST API** para la gestión del **ciclo de evaluación de desempeño** y las **transiciones de fase** que lo rigen. El ciclo anual tiene 3 fases — `asignación`, `avance`, `cierre` — y este cambio permite crear ciclos, consultar su estado, avanzar de fase y validar las reglas de transición.

Refleja las decisiones del spec `evaluation-lifecycle` (B1):
- **Decisión #3**: en fase `avance`, se permite editar metas y registrar progreso, pero **NO** eliminar metas ni crear nuevas. Esto afecta las reglas de `blockedActions` en `PhaseDefinition` y las validaciones de transición.
- **Decisión #4**: en fase `cierre`, existen **tres caminos paralelos** — autoevaluación, evaluación 9x9 del jefe, y evaluación formal de RH — que se modelan como acciones permitidas concurrentes en la misma fase.

El API es el pilar sobre el cual se construirán C3 (competencias), C4 (objetivos del empleado) y C6 (evaluaciones), ya que cada uno de esos módulos depende de conocer la fase activa del ciclo.

## Scope

### In Scope

- `GET /api/v1/cycles` — listado de ciclos con paginación por cursor, filtros por `organization_id`, `year` y `current_phase`.
- `POST /api/v1/cycles` — creación de un nuevo ciclo de evaluación.
- `GET /api/v1/cycles/:id` — detalle de un ciclo específico.
- `PUT /api/v1/cycles/:id/transition` — avance de fase del ciclo (transición lineal).
- `GET /api/v1/phases` — catálogo estático de definiciones de fase (`PhaseDefinition`).
- `GET /api/v1/cycles/:id/transitions` — transiciones disponibles desde la fase actual del ciclo.
- Validaciones de negocio:
  - **Un solo ciclo activo por año por organización** (`(organization_id, year)` unique).
  - **Transiciones lineales únicamente** (`asignación → avance → cierre`); no se permite retroceder ni saltar.
  - Verificación de condiciones de transición antes de aplicarla.
- Hooks de autorización basada en fase (`allowedActors`, `allowedActions`, `blockedActions`) — implementados como **middleware preparado** con `TODO(auth:C7)` para integración con el módulo de autenticación.
- Especificación OpenAPI 3.1 completa y validada.

### Out of Scope

- CRUD de objetivos del empleado (C4).
- Asignación de competencias (C3).
- Autoevaluación, evaluación 9x9 y evaluación formal de RH (C6).
- Autenticación y RBAC completa (C7) — se incluyen **marcadores `TODO(auth:C7)`** en middleware y handlers donde debe integrarse validación de roles y permisos.
- Notificaciones por email o push al cambiar de fase.

## Concurrency & High-Volume Design

Este API debe sostener **miles a millones de requests** por endpoint en ventanas de pico (cierre de ciclo, apertura de evaluación). El diseño asume PostgreSQL 15+ y entrega los siguientes mecanismos de concurrencia y rendimiento.

### Database Connection Management

- **pgx connection pool** configurado con:
  - `MaxConns = 25`
  - `MinConns = 5`
  - `MaxConnLifetime = 1h`
  - `MaxConnIdleTime = 30m`
- El `EntClient` se inyecta por request; las operaciones de escritura usan `Tx` con aislamiento explícito.
- **Réplicas de lectura** para endpoints GET no transaccionales (listado de ciclos, catálogo de fases). El enrutador de queries selecciona réplica si `dbRole = read`.
- **Métricas de pool** expuestas vía Prometheus: `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms`.

### Concurrency Control

- **Optimistic locking** en actualizaciones de `Cycle`: cada fila lleva `version` (integer) o se usa `updated_at` como token de concurrencia. El `PUT` de transición incluye `If-Match` con el `version` actual; si hay conflicto, se retorna `409 CONCURRENT_UPDATE`.
- **SELECT FOR UPDATE SKIP LOCKED** en transiciones de fase para evitar race conditions cuando múltiples actores (RH, jefe) intentan avanzar el mismo ciclo simultáneamente.
- **Idempotency keys** en `POST /cycles` y `PUT /cycles/:id/transition`: el cliente envía `Idempotency-Key: <uuid>` en header. El servidor guarda el key con TTL 24h en Redis; si se reenvía la misma petición, retorna el resultado previo sin reejecutar la transacción.
- **Advisory locks de PostgreSQL** (`pg_advisory_lock`) para la creación de ciclos, garantizando que solo un request a la vez pueda verificar la unicidad de `(organization_id, year)` y crear el registro.

### High-Volume Patterns

- **Paginación por cursor** en `GET /api/v1/cycles` (`?cursor=...&limit=...`). No se usa `OFFSET` para evitar degradación en listados grandes.
- **ETag / If-None-Match** en `GET /api/v1/phases` (datos estáticos). El cliente puede cachear el catálogo indefinidamente; el servidor retorna `304 Not Modified` si el hash no cambió.
- **Rate limiting**:
  - Escrituras (`POST`, `PUT`, `PATCH`): **100 req/s por organización**.
  - Lecturas (`GET`): **1000 req/s por organización**.
  - Implementado en capa de proxy (nginx/envoy) o middleware de Chi con Redis como store.
- **Timeouts de contexto**:
  - Lecturas: 5s
  - Escrituras: 10s
  - Transiciones de fase: 30s (incluye validaciones y logging)
- **Streaming JSON** para payloads grandes de detalle de ciclo (si se incluye lista de participantes en el futuro); para este cambio, respuesta JSON estándar con proyección ligera.
- **Redis cache** para `PhaseDefinition` (TTL 1h). Cache invalidado manualmente al modificar el catálogo (eventos raros). Clave: `phases:definitions:v1`.

### Transaction Design

- Las transiciones de fase se ejecutan dentro de una **transacción de base de datos** que:
  1. Bloquea el ciclo con `SELECT FOR UPDATE`.
  2. Valida la transición contra `PhaseTransition`.
  3. Actualiza `Cycle.current_phase` y `version`.
  4. Inserta registro en tabla de auditoría `cycle_phase_history` (quién, cuándo, desde/hasta qué fase).
  5. Confirma el commit.
- **Aislamiento**: `READ COMMITTED` (suficiente para este dominio; no requiere `SERIALIZABLE` dado el locking explícito).
- **Duración objetivo**: < 100ms por transacción de transición.
- **Defer constraint checks**: cuando sea posible, diferir validaciones de foreign keys al final de la transacción para reducir overhead.

## API Endpoints

| Método | Ruta | Descripción | Auth (TODO) |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/cycles` | Listado de ciclos. Query params: `organization_id` (req), `year`, `current_phase`, `cursor`, `limit` (default 20, max 100). | `TODO(auth:C7)` — rol `rh` o `admin` de la org. |
| `POST` | `/api/v1/cycles` | Crea un nuevo ciclo. Body: `year`, `organization_id`. Retorna `201` con ciclo creado en fase `asignación`. | `TODO(auth:C7)` — rol `rh` o `admin` de la org. |
| `GET` | `/api/v1/cycles/:id` | Detalle de ciclo incluyendo fase actual, año, organización y metadata. | `TODO(auth:C7)` — miembro de la org. |
| `PUT` | `/api/v1/cycles/:id/transition` | Avanza la fase del ciclo. Body opcional: `trigger` (`auto` o `manual-rh`), `reason`. Requiere `If-Match` con `version` actual. | `TODO(auth:C7)` — actor permitido según `allowedActors` de la fase destino. |
| `GET` | `/api/v1/phases` | Catálogo estático de fases. Incluye `phase`, `label`, `order`, `allowedActors`, `allowedActions`, `blockedActions`. Cacheable con ETag. | Público (o `TODO(auth:C7)` — cualquier usuario autenticado). |
| `GET` | `/api/v1/cycles/:id/transitions` | Transiciones disponibles desde la fase actual del ciclo `:id`. Incluye `from_phase`, `to_phase`, `trigger`, `conditions`. | `TODO(auth:C7)` — miembro de la org. |

### Proyecciones de respuesta

- **Listado (`GET /cycles`)**: proyección ligera — `id`, `year`, `current_phase`, `organization_id`, `created_at`, `updated_at`. Sin joins pesados.
- **Detalle (`GET /cycles/:id`)**: proyección completa del ciclo. En futuras iteraciones puede incluir conteos de participantes (proyección separada o endpoint adicional).
- **Transiciones (`GET /cycles/:id/transitions`)**: solo lectura de `PhaseTransition` filtrada por `from_phase = cycle.current_phase`.

## Error Handling

Todos los errores siguen el formato estándar:

```json
{
  "error": {
    "code": "INVALID_TRANSITION",
    "message": "La transición solicitada no es válida desde la fase actual.",
    "details": [
      "from_phase: avance",
      "requested_phase: asignacion",
      "reason: las transiciones son lineales y no se puede retroceder"
    ],
    "trace_id": "abc123-def456"
  }
}
```

### Códigos de error específicos

| Código | HTTP | Descripción |
|--------|------|-------------|
| `CYCLE_NOT_FOUND` | `404` | El ciclo solicitado no existe. |
| `INVALID_TRANSITION` | `409` | La transición no está definida en `PhaseTransition` o viola dirección lineal. |
| `CYCLE_ALREADY_ACTIVE` | `409` | Ya existe un ciclo activo para el mismo `(organization_id, year)`. |
| `PHASE_NOT_ADVANCEABLE` | `409` | La fase actual no permite avanzar (condiciones de `PhaseTransition` no satisfechas). |
| `CONCURRENT_UPDATE` | `409` | Conflicto de optimistic locking (`version` o `updated_at` no coincide). |
| `IDEMPOTENCY_KEY_CONFLICT` | `409` | El `Idempotency-Key` ya fue usado con un payload diferente. |
| `RATE_LIMIT_EXCEEDED` | `429` | Límite de requests por organización superado. |
| `REQUEST_TIMEOUT` | `408` | El contexto expiró antes de completar la operación. |

## Non-Goals (explicit)

- **No** gestión de objetivos del empleado (CRUD de metas, progreso, etc.).
- **No** asignación de competencias al puesto o al empleado.
- **No** evaluaciones finales: autoevaluación, 9x9, evaluación formal de RH.
- **No** autenticación, autorización ni RBAC completo — solo marcadores `TODO(auth:C7)` para integración futura.
- **No** envío de notificaciones por email o push al cambiar de fase.
- **No** soporte para múltiples ciclos simultáneos en la misma organización y año (la regla de negocio es uno activo).
- **No** rollback de fase (transición hacia atrás). Si se requiere corrección, se hará por vía administrativa fuera de este API.

## Dependencies

| Cambio | Descripción | Estado requerido |
|--------|-------------|-----------------|
| **C1** | `data-model-core` — creación de tablas `Cycle`, `PhaseDefinition`, `PhaseTransition` y sus índices. | **Aplicado primero** |
| **B1** | `evaluation-lifecycle` spec — definición del modelo de dominio, reglas de transición y fases. | **Finalizado** |
| **C7** | `auth-rbac` — middleware de autenticación y roles (este cambio lo referencia con `TODO(auth:C7)`). | Referencia futura |

## Success Criteria

- [ ] Todos los endpoints documentados en esta propuesta están implementados en Go con Chi.
- [ ] **OpenAPI 3.1** generado y validado (usando `oapi-codegen` o similar); tipos TypeScript generados en `web/src/lib/api/`.
- [ ] **Tests unitarios** para cada handler: casos felices, errores de validación, transiciones inválidas, ciclo duplicado.
- [ ] **Tests de concurrencia**: al menos 50 goroutines concurrentes intentando transicionar el mismo ciclo; ninguna race condition, solo una transición exitosa, el resto recibe `CONCURRENT_UPDATE` o `PHASE_NOT_ADVANCEABLE`.
- [ ] **Tests de carga**: `GET /api/v1/cycles` sostiene **1000 req/s** durante 60s sin degradación de latencia > p95 200ms.
- [ ] **Tests de idempotencia**: reenvío de `POST /cycles` con mismo `Idempotency-Key` retorna `201` idéntico sin crear duplicado.
- [ ] **Tests de pool**: métricas Prometheus `pgx_pool_*` expuestas y verificables.
- [ ] **Documentación**: `README.md` en `internal/service/cycle/` y `internal/handler/cycle/` explicando el flujo de transición y los mecanismos de locking.

---

*Generado por OpenSpec Proposal Generator para el proyecto SED Evaluación de Desempeño.*
