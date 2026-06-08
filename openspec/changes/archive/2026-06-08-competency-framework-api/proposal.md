# C3: competency-framework-api

## Intent

C3 entrega el **REST API** para la administracion del **catalogo de competencias** — pilares, competencias, criterios de escala, definiciones de nivel, niveles de aceptacion por competencia x perfil, y perfiles de evaluacion. Este es el API de administracion de RH que gestiona el catalogo de competencias utilizado en todas las evaluaciones de desempeno del sistema SED.

Refleja las decisiones del spec `competency-framework` (B2):
- **Decision #1**: los pilares **NO tienen ponderacion**; no existe campo de peso en la entidad `Pillar`.
- **Decision #2**: catalogo **unico** de pilares y competencias para **todos los perfiles** de evaluacion. No hay pilares privados ni filtrados por rol.
- **Decision #5**: las **categorias de metas** (definidas por empleado) son **independientes** de los pilares de competencias (definidos por la empresa). No existe vinculo entre ambos dominios.

Este API es consumido por la UI de administracion de RH (pantalla A2) y alimenta los modulos de asignacion anual (A3) y evaluacion formal (C6).

## Scope

### In Scope

- `GET /api/v1/pillars` — listado de pilares con conteo de competencias.
- `POST /api/v1/pillars` — creacion de un nuevo pilar.
- `GET /api/v1/pillars/:id` — detalle de un pilar con sus competencias.
- `PUT /api/v1/pillars/:id` — actualizacion de nombre y descripcion de un pilar.
- `DELETE /api/v1/pillars/:id` — eliminacion de pilar con cascada a competencias, criterios y niveles de aceptacion.
- `GET /api/v1/pillars/:pillarId/competencies` — listado de competencias dentro de un pilar.
- `POST /api/v1/pillars/:pillarId/competencies` — creacion de competencia dentro de un pilar.
- `GET /api/v1/competencies/:id` — detalle de competencia con sus criterios de escala.
- `PUT /api/v1/competencies/:id` — actualizacion de competencia (nombre, descripcion, pilar).
- `DELETE /api/v1/competencies/:id` — eliminacion de competencia con cascada a criterios y niveles de aceptacion.
- `GET /api/v1/competencies/:id/scale-criteria` — criterios de escala de una competencia (por nivel 1-5).
- `POST /api/v1/competencies/:id/scale-criteria` — creacion o reemplazo masivo de criterios para una competencia (bulk write).
- `GET /api/v1/levels` — definiciones globales de nivel (5 registros fijos, catalogo de solo lectura para usuarios no-RH).
- `GET /api/v1/acceptance-levels` — todos los niveles de aceptacion, filtrables por `profile_id` y `competency_id`.
- `POST /api/v1/acceptance-levels` — establece o actualiza el nivel de aceptacion para una combinacion competencia x perfil.
- `GET /api/v1/profiles` — perfiles de evaluacion (8 registros fijos, catalogo de solo lectura).
- Validaciones de negocio:
  - `name` unico en `Pillar` y `EvaluationProfile`.
  - `level` en `ScaleCriterion` y `CompetencyAcceptanceLevel` restringido al rango 1-5.
  - Un pilar no puede eliminarse si aun contiene competencias (protegido por cascada explicita + validacion de aplicacion).
  - Un `CompetencyAcceptanceLevel` tiene clave unica `(competency_id, profile_id)`.
- Especificacion **OpenAPI 3.1** completa y validada.

### Out of Scope

- **Asignacion masiva de competencias a empleados** — es responsabilidad de A3 (asignacion anual) y su API correspondiente.
- **Importacion desde Excel** — carga masiva de pilares, competencias o criterios.
- **Gestion de objetivos del empleado** — metas, progreso, categorias de metas (C4).
- **Autenticacion y RBAC** — sesiones, roles y permisos son C7. Este cambio incluye **marcadores `TODO(auth:C7)`** en handlers donde debe integrarse validacion de roles (solo `rh` y `admin` pueden escribir en este API).
- **Evaluaciones formales** — autoevaluacion, 9x9, evaluacion RH (C6).
- **Notificaciones** por email ante cambios en el catalogo.
- **Audit log detallado** — quien cambio que criterio y cuando (scope futuro de C7).

## Concurrency & High-Volume Design

Este API debe sostener **miles a millones de requests** por endpoint, especialmente en lecturas del catalogo de competencias (cargado por multiples usuarios concurrentes durante evaluaciones). El diseno asume PostgreSQL 15+ y entrega los siguientes mecanismos de concurrencia y rendimiento.

### Database Connection Management

- **pgx connection pool** configurado con:
  - `MaxConns = 25`
  - `MinConns = 5`
  - `MaxConnLifetime = 1h`
  - `MaxConnIdleTime = 30m`
- El `EntClient` se inyecta por request y su ciclo de vida esta atado al request HTTP. Las operaciones de escritura usan `Tx` con aislamiento explicito.
- **Replicas de lectura** para endpoints GET no transaccionales (listado de pilares, competencias, definiciones de nivel y perfiles). El enrutador de queries selecciona replica si `dbRole = read`.
- **Metricas de pool** expuestas via Prometheus: `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms`.

### Concurrency Control

- **Optimistic locking** en actualizaciones de `Pillar` y `Competency`: cada fila lleva `updated_at` como token de concurrencia. Los `PUT` incluyen `If-Match` con el timestamp actual; si hay conflicto, se retorna `409 CONCURRENT_UPDATE`.
- **Version field en `ScaleCriterion`**: campo `version` (integer) incrementado en cada escritura de criterios, permitiendo que multiples usuarios de RH editen criterios concurrentemente sin perdida de datos.
- **Advisory locks de PostgreSQL** (`pg_advisory_lock`) para eliminacion de pilares: evita que dos requests de RH intenten borrar el mismo pilar simultaneamente, validando previamente que no existan competencias huerfanas en ventana de race condition.
- **Idempotency keys** en `POST /api/v1/pillars`, `POST /api/v1/pillars/:pillarId/competencies` y `POST /api/v1/competencies/:id/scale-criteria`: el cliente envia `Idempotency-Key: <uuid>` en header. El servidor guarda el key con TTL 24h en Redis; si se reenvia la misma peticion, retorna el resultado previo sin reejecutar la transaccion.

### High-Volume Patterns

- **ETag / If-None-Match** en `GET /api/v1/levels` y `GET /api/v1/profiles` (datos estaticos, 5 y 8 registros respectivamente). El cliente puede cachear estos catalogos agresivamente; el servidor retorna `304 Not Modified` si el hash no cambio.
- **Redis cache** para el arbol completo de competencias (pilares -> competencias -> criterios). Clave: `competency:tree:v1`. Invalidado en cualquier escritura exitosa (`POST`, `PUT`, `DELETE` en pilares, competencias o criterios). TTL 1h como respaldo.
- **Paginacion por cursor** en `GET /api/v1/pillars` y `GET /api/v1/pillars/:pillarId/competencies` (`?cursor=...&limit=...`). No se usa `OFFSET` para evitar degradacion en catalogos grandes.
- **Bulk GET**: `GET /api/v1/pillars` con parametro `?include=competencies` permite obtener todos los pilares con sus competencias en una sola request, evitando N+1 en el frontend.
- **Rate limiting**:
  - Escrituras (`POST`, `PUT`, `DELETE`): **50 req/s** por usuario (RH admin).
  - Lecturas (`GET`): **500 req/s** por IP/organizacion.
  - Implementado en capa de proxy (nginx/envoy) o middleware de Chi con Redis como store.
- **Timeouts de contexto**:
  - Lecturas: 3s
  - Escrituras: 8s
  - Operaciones de cascada (delete de pilar): 15s
- **Streaming JSON** para exportaciones grandes del catalogo (endpoint futuro; para este cambio, respuesta JSON estandar con proyeccion ligera en listados).

### Transaction Design

- Eliminacion de pilar con cascada envuelta en una **transaccion de base de datos** que:
  1. Bloquea el pilar con `SELECT FOR UPDATE`.
  2. Valida que no existan competencias activas asociadas (proteccion de aplicacion + foreign key).
  3. Elimina en cascada `ScaleCriterion`, `CompetencyAcceptanceLevel` y `Competency`.
  4. Elimina el pilar.
  5. Invalida la cache Redis del arbol de competencias.
  6. Confirma el commit.
- Actualizacion de competencia con criterios en una **transaccion unica**:
  1. Actualiza campos de la competencia.
  2. Reemplaza criterios de escala (bulk delete + insert si es necesario).
  3. Incrementa `version` en criterios afectados.
  4. Invalida cache.
- **Aislamiento**: `READ COMMITTED` (suficiente para catalogo; locking explicito en operaciones criticas).
- **Duracion objetivo**: < 50ms por operacion de catalogo simple; < 200ms para transacciones con cascada.

## API Endpoints

| Metodo | Ruta | Descripcion | Auth (TODO) |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/pillars` | Listado de pilares. Query: `?include=competencies`, `cursor`, `limit` (default 20, max 100). Proyeccion ligera: `id`, `name`, `description`, `competency_count`. | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `POST` | `/api/v1/pillars` | Crea un nuevo pilar. Body: `name`, `description`. Retorna `201` con pilar creado. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/pillars/:id` | Detalle de pilar incluyendo competencias anidadas (si `?include=competencies`). | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `PUT` | `/api/v1/pillars/:id` | Actualiza pilar. Body: `name`, `description`. Requiere `If-Match` con `updated_at` actual. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `DELETE` | `/api/v1/pillars/:id` | Elimina pilar y todo su contenido en cascada. Requiere confirmacion de aplicacion (no vacio). | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/pillars/:pillarId/competencies` | Listado de competencias del pilar. Paginacion por cursor. | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `POST` | `/api/v1/pillars/:pillarId/competencies` | Crea competencia en el pilar. Body: `name`, `description`. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/competencies/:id` | Detalle de competencia con criterios de escala agrupados por nivel. | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `PUT` | `/api/v1/competencies/:id` | Actualiza competencia (nombre, descripcion, puede mover de pilar). Requiere `If-Match`. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `DELETE` | `/api/v1/competencies/:id` | Elimina competencia con cascada a criterios y niveles de aceptacion. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/competencies/:id/scale-criteria` | Criterios de escala de la competencia. Respuesta agrupada por `level` (1-5). | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `POST` | `/api/v1/competencies/:id/scale-criteria` | Crea o reemplaza criterios para la competencia (bulk). Body: array de `{level, description}`. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/levels` | Definiciones globales de nivel (5 registros). Cacheable con ETag. | `TODO(auth:C7)` — cualquier usuario autenticado. |
| `GET` | `/api/v1/acceptance-levels` | Niveles de aceptacion. Query: `?profile_id=...&competency_id=...`. | `TODO(auth:C7)` — cualquier usuario autenticado de la org. |
| `POST` | `/api/v1/acceptance-levels` | Establece nivel de aceptacion. Body: `competency_id`, `profile_id`, `level` (1-5). Upsert. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/profiles` | Perfiles de evaluacion (8 registros). Cacheable con ETag. | `TODO(auth:C7)` — cualquier usuario autenticado. |

### Proyecciones de respuesta

- **Listado de pilares (`GET /pillars`)**: proyeccion ligera — `id`, `name`, `description`, `competency_count`. Sin joins pesados.
- **Detalle de pilar (`GET /pillars/:id`)**: proyeccion completa del pilar con competencias anidadas si se solicita.
- **Detalle de competencia (`GET /competencies/:id`)**: incluye criterios de escala agrupados por nivel, evitando N+1.
- **Catalogos estaticos (`GET /levels`, `GET /profiles`)**: respuesta plana completa; habilitados para cache agresivo.

## Error Handling

Todos los errores siguen el formato estandar:

```json
{
  "error": {
    "code": "PILLAR_HAS_COMPETENCIES",
    "message": "No se puede eliminar el pilar porque aun contiene competencias.",
    "details": [
      "pillar_id: 550e8400-e29b-41d4-a716-446655440000",
      "competencies_count: 3",
      "action_required: eliminar competencias primero o usar force=false"
    ],
    "trace_id": "abc123-def456"
  }
}
```

### Codigos de error especificos

| Codigo | HTTP | Descripcion |
|--------|------|-------------|
| `PILLAR_NOT_FOUND` | `404` | El pilar solicitado no existe. |
| `COMPETENCY_NOT_FOUND` | `404` | La competencia solicitada no existe. |
| `PILLAR_HAS_COMPETENCIES` | `409` | No se puede eliminar un pilar que aun contiene competencias (validacion de aplicacion). |
| `COMPETENCY_HAS_CRITERIA` | `409` | No se puede eliminar una competencia que tiene criterios de escala asociados (proteccion de cascada). |
| `DUPLICATE_NAME` | `409` | El nombre ya existe en `Pillar` o `EvaluationProfile`. |
| `INVALID_LEVEL` | `400` | El nivel debe estar entre 1 y 5. |
| `CONCURRENT_UPDATE` | `409` | Conflicto de optimistic locking (`updated_at` no coincide). |
| `IDEMPOTENCY_KEY_CONFLICT` | `409` | El `Idempotency-Key` ya fue usado con un payload diferente. |
| `RATE_LIMIT_EXCEEDED` | `429` | Limite de requests superado. |
| `REQUEST_TIMEOUT` | `408` | El contexto expiro antes de completar la operacion. |

## Non-Goals (explicit)

- **No** asignacion masiva de competencias a empleados (es A3 / C3-asignacion).
- **No** importacion desde Excel de pilares, competencias o criterios.
- **No** gestion de objetivos del empleado (metas, progreso, categorias) — es C4.
- **No** autoevaluacion, evaluacion 9x9 ni evaluacion formal de RH — es C6.
- **No** autenticacion, autorizacion ni RBAC completo — solo marcadores `TODO(auth:C7)` para integracion futura.
- **No** envio de notificaciones por email ante cambios en el catalogo.
- **No** audit log detallado de quien modifico cada criterio.
- **No** ponderacion de pilares (decision #1 ya cubierta en modelo).
- **No** competencias transversales entre multiples pilares.

## Dependencies

| Cambio / Spec | Descripcion | Estado requerido |
|--------|-------------|-----------------|
| **C1** | `data-model-core` — creacion de tablas `Pillar`, `Competency`, `ScaleCriterion`, `LevelDefinition`, `CompetencyAcceptanceLevel`, `EvaluationProfile` y sus indices. | **Aplicado primero** |
| **B2** | `competency-framework` spec — definicion del modelo de dominio, reglas de catalogo unico y escala 1-5. | **Finalizado** |
| **C7** | `auth-rbac` — middleware de autenticacion y roles (este cambio lo referencia con `TODO(auth:C7)`). | Referencia futura |

## Success Criteria

- [ ] Todos los endpoints documentados en esta propuesta estan implementados en Go con Chi.
- [ ] **OpenAPI 3.1** generado y validado; tipos TypeScript generados en `web/src/lib/api/`.
- [ ] **Tests unitarios** para cada handler: casos felices, errores de validacion, duplicados, niveles invalidos, cascade delete.
- [ ] **Tests de concurrencia**: al menos 50 goroutines concurrentes intentando actualizar criterios de la misma competencia; ninguna perdida de datos, versiones consistentes.
- [ ] **Tests de carga**: `GET /api/v1/pillars` sostiene **500 req/s** durante 60s sin degradacion de latencia > p95 200ms.
- [ ] **Tests de idempotencia**: reenvio de `POST /pillars` con mismo `Idempotency-Key` retorna `201` identico sin crear duplicado.
- [ ] **Tests de cache ETag**: `GET /levels` y `GET /profiles` retornan `304` en segundo request con `If-None-Match`.
- [ ] **Tests de pool**: metricas Prometheus `pgx_pool_*` expuestas y verificables.
- [ ] **Sin N+1**: el listado de pilares con `?include=competencies` resuelve el arbol completo en una sola query o con eager loading de Ent.
- [ ] **Documentacion**: `README.md` en `internal/service/competency/` y `internal/handler/competency/` explicando el flujo de cascada y los mecanismos de locking.

---

*Generado por OpenSpec Proposal Generator para el proyecto SED Evaluacion de Desempeno.*
