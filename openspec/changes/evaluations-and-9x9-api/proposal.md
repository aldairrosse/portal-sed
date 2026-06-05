# C6: evaluations-and-9x9-api

## Intent

C6 entrega el **REST API** para la gestión de **evaluaciones de desempeño** y la **matriz 9×9** en fase de cierre. Este cambio implementa las tres vías paralelas de evaluación fin de año: autoevaluación del empleado, evaluación formal de RH, y calificación 9×9 del jefe (desempeño vs potencial). Es el API de cierre anual — el período de mayor tráfico del sistema.

Refleja las decisiones de los specs `evaluation-lifecycle` (B1) y `manager-9x9` (B5):
- **Decisión #4**: en fase `cierre`, existen **tres caminos paralelos e independientes** — autoevaluación, evaluación 9×9 del jefe, y evaluación formal de RH. La 9×9 del jefe **NO** sustituye la evaluación RH.
- El cuadrante 9×9 se computa en escritura (denormalizado) para evitar recálculo en lectura.
- Cada evaluador ve **solo su propia matriz**; los colaboradores no tienen acceso al menú 9×9.

Este API consume el modelo de datos de C1 (`data-model-core`) y depende de C2 (`evaluation-lifecycle-api`) para validación de fase, C4 (`goals-api`) para estado de metas, y C5 (`org-hierarchy-api`) para resolución de evaluador/evaluado.

## Scope

### In Scope

- **Autoevaluación**: submit de calificaciones de competencias (escala 1–5) + comentarios de cierre de metas.
- **Evaluación formal de RH**: submit de calificaciones de competencias + cierre de la evaluación.
- **Matriz 9×9 del jefe**: creación/actualización de entradas en la matriz (desempeño 1–9 + potencial 1–9).
- **Seguimiento de estado de evaluación**: transiciones `pendiente-evaluacion-final` → `en-progreso` → `completada`.
- **Cálculo de cuadrante 9×9**: computado automáticamente al escribir scores de desempeño y potencial.
- **Endpoint de resumen/dashboard**: conteos de evaluaciones por estado para vista RH.
- **Listado de evaluaciones por ciclo**: vista RH con filtros y paginación.
- **OpenAPI 3.1** como contrato fuente de verdad; tipos TypeScript generados en `web/src/lib/api/`.

### Out of Scope

- CRUD de metas del empleado (C4).
- Gestión del catálogo de competencias (C3).
- Notificaciones/email al completar evaluación (C7).
- Autenticación y RBAC completo (C7) — se incluyen **marcadores `TODO(auth:C7)`** en middleware y handlers.
- Reportes agregados across evaluadores (consolidado de empresa).
- Comparación histórica entre ciclos.
- Exportación a PDF/Excel.

## Concurrency & High-Volume Design

Este API debe sostener **miles a millones de requests** en ventanas de pico (cierre de año: TODOS los empleados submiten simultáneamente). El diseño asume PostgreSQL 15+ y entrega los siguientes mecanismos de concurrencia y rendimiento.

### Database Connection Management

- **pgx connection pool** configurado con:
  - `MaxConns = 40`
  - `MinConns = 15`
  - `MaxConnLifetime = 1h`
  - `MaxConnIdleTime = 30m`
- El `EntClient` se inyecta por request; las operaciones de escritura usan `Tx` con aislamiento explícito.
- **Réplicas de lectura** para endpoints GET no transaccionales (listado de evaluaciones, vistas de matriz 9×9, definiciones de escala/cuadrante).
- **Métricas de pool** expuestas vía Prometheus: `pgx_pool_conns_busy`, `pgx_pool_conns_idle`, `pgx_pool_wait_duration_ms`.

### Concurrency Control

- **Optimistic locking** en actualizaciones de `Evaluation`: cada fila lleva `version` (integer). El `PUT` de evaluación incluye `If-Match` con el `version` actual; si hay conflicto, se retorna `409 CONCURRENT_UPDATE`. Esto protege contra autoevaluación concurrente con evaluación RH.
- **SELECT FOR UPDATE** en `NineBoxEntry` al actualizar scores — previene cálculo de cuadrante stale cuando múltiples pestañas del jefe intentan guardar simultáneamente.
- **Advisory locks de PostgreSQL** (`pg_advisory_lock`) en la finalización de evaluación (`POST /evaluations/:id/finalize`) — previene doble-cierre si RH envía la petición dos veces.
- **Idempotency keys** en `POST /evaluations/:id/self-evaluation` y `POST /evaluations/:id/rh-evaluation`: el cliente envía `Idempotency-Key: <uuid>` en header. El servidor guarda el key con TTL 24h en Redis; reenvíos retornan el resultado previo sin reejecutar la transacción.
- **Máquina de estados suave**: las transiciones de estado de `Evaluation` se actualizan condicionalmente (`WHERE state = 'pendiente-evaluacion-final'`), garantizando atomicidad sin necesidad de locking adicional en la mayoría de casos.

### High-Volume Patterns

- **Paginación por cursor** en `GET /api/v1/evaluations` (`?cursor=...&limit=...`). No se usa `OFFSET` para evitar degradación en listados de 10K+ empleados.
- **Batch endpoint**: `POST /api/v1/nine-box/batch` — permite submitir múltiples entradas 9×9 en una única transacción atómica, reduciendo round trips a la BD.
- **Vista materializada pre-computada** `evaluation_summary` — conteos por estado, refrescada en trigger al cambiar `evaluation.state`. El endpoint `GET /api/v1/evaluations/summary` lee esta vista en vez de agregar en tiempo real.
- **Redis cache** para `NineBoxScale` y `NineBoxQuadrant` (TTL 1h). Son catálogos estáticos; claves: `ninebox:scales:v1`, `ninebox:quadrants:v1`.
- **ETag / If-None-Match** en `GET /api/v1/nine-box/scales` y `GET /api/v1/nine-box/quadrants`. El servidor retorna `304 Not Modified` si el hash no cambió.
- **Rate limiting**:
  - Escrituras de evaluación (`POST`, `PUT`): **100 req/s por organización**.
  - Lecturas (`GET`): **2000 req/s por organización**.
  - Implementado en capa de proxy o middleware de Chi con Redis como store.
- **Timeouts de contexto**:
  - Lecturas: 5s
  - Escrituras de evaluación: 20s (incluyen múltiples competency ratings + goal comments + state)
  - Batch 9×9: 30s
  - Finalización: 30s
- **Streaming JSON** para exportaciones grandes de evaluación (reservado para futuro; en este change, respuestas JSON estándar con proyección ligera).
- **Circuit breaker** en escrituras a BD durante peak de cierre de año: si el pool de conexiones está saturado (> 90% ocupado por > 10s), se rechazan nuevas escrituras con `503 SERVICE_UNAVAILABLE` y se redirige al cliente a reintentar con backoff exponencial.
- **Queue-based write smoothing**: si > 500 escrituras concurrentes de evaluación, se encolan en un buffer Redis-backed con workers que drenan a la BD en lotes. Prioridad: submissions cerca del deadline.
- **Batch aggregation de competency ratings**: el endpoint de autoevaluación y RH evaluación recopilan todos los ratings en un solo round trip a la BD (INSERT/UPDATE múltiples filas de `EvaluationCompetency` en una query).

### Transaction Design

- **Autoevaluación** (`POST /evaluations/:id/self-evaluation`): transacción que:
  1. Bloquea la `Evaluation` con `SELECT FOR UPDATE`.
  2. Valida que la evaluación no esté ya finalizada y que el ciclo esté en fase `cierre`.
  3. Actualiza todas las filas de `EvaluationCompetency` (ratings 1–5).
  4. Actualiza `EvaluationGoal` (comentarios de cierre).
  5. Actualiza `Evaluation.state` → `completada` (si es submit final) o `en-progreso`.
  6. Commit.
- **Evaluación RH** (`POST /evaluations/:id/rh-evaluation`): transacción que:
  1. Bloquea la `Evaluation` con `SELECT FOR UPDATE`.
  2. Actualiza `EvaluationCompetency`.
  3. Actualiza `Evaluation.state` → `completada`.
  4. Commit.
- **Finalización RH** (`POST /evaluations/:id/finalize`): transacción que:
  1. Adquiere advisory lock sobre `evaluation_id`.
  2. Verifica que la evaluación esté en estado permitido.
  3. Actualiza `Evaluation.state` → `completada` + timestamps.
  4. Invalida vista materializada `evaluation_summary`.
  5. Commit.
- **9×9 entry** (`POST/PUT /nine-box/entries/:entryId`): transacción que:
  1. Bloquea `NineBoxEntry` con `SELECT FOR UPDATE`.
  2. Actualiza `performanceScore` y `potentialScore`.
  3. Recomputa `quadrant` basado en la matriz de rangos (1–9).
  4. Persiste `quadrant` denormalizado.
  5. Commit.
- **Aislamiento**:
  - Lecturas: `READ COMMITTED`.
  - Transiciones de estado y finalización: `REPEATABLE READ`.
  - Duración objetivo por transacción: < 100ms (single entry), < 300ms (batch 20 entries).
- **Year-End Burst Pattern**:
  - Deadline-aware priority: submissions dentro de 24h del deadline reciben prioridad en la cola de escritura.
  - Pre-warming de conexiones (`MinConns = 15`) antes de la ventana de pico.

## API Endpoints

### Evaluaciones

| Método | Ruta | Descripción | Auth (TODO) |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/evaluations` | Listado de evaluaciones. Query params: `cycle_id` (req), `state`, `employee_id`, `cursor`, `limit` (default 20, max 100). Proyección ligera. | `TODO(auth:C7)` — rol `rh` o `admin` de la org. |
| `GET` | `/api/v1/evaluations/:id` | Detalle completo de evaluación incluyendo competencias (`EvaluationCompetency`) y metas (`EvaluationGoal`). | `TODO(auth:C7)` — dueño de la evaluación, su jefe, o `rh`. |
| `POST` | `/api/v1/evaluations/:id/self-evaluation` | Empleado submita autoevaluación. Body: `competencies[{competencyId, rating(1-5), comments?}]`, `goalComments[{goalId, comment?}]`. Requiere `Idempotency-Key`. | `TODO(auth:C7)` — empleado dueño de la evaluación. |
| `PUT` | `/api/v1/evaluations/:id/self-evaluation` | Actualiza autoevaluación si aún no está finalizada. Body: mismo schema que POST. Requiere `If-Match`. | `TODO(auth:C7)` — empleado dueño. |
| `POST` | `/api/v1/evaluations/:id/rh-evaluation` | RH submita evaluación formal. Body: `competencies[{competencyId, rating(1-5), comments?}]`, `finalComments?`. Requiere `Idempotency-Key`. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `PUT` | `/api/v1/evaluations/:id/rh-evaluation` | Actualiza evaluación RH si no está finalizada. Requiere `If-Match`. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `POST` | `/api/v1/evaluations/:id/finalize` | Finaliza la evaluación (cierre definitivo). Body opcional: `reason`. Requiere advisory lock interno. | `TODO(auth:C7)` — rol `rh` o `admin`. |
| `GET` | `/api/v1/evaluations/summary` | Resumen/dashboard: conteos por `state` para el ciclo activo. Lee vista materializada. | `TODO(auth:C7)` — rol `rh` o `admin`. |

### Matriz 9×9

| Método | Ruta | Descripción | Auth (TODO) |
|--------|------|-------------|-------------|
| `GET` | `/api/v1/nine-box/matrices` | Listado de matrices. Query params: `cycle_id`, `evaluator_id`. | `TODO(auth:C7)` — rol `rh` o evaluador dueño. |
| `GET` | `/api/v1/nine-box/matrices/:matrixId` | Detalle de matriz con todas sus entradas (`NineBoxEntry`). | `TODO(auth:C7)` — evaluador dueño de la matriz, o `rh`. |
| `POST` | `/api/v1/nine-box/matrices` | Crea una matriz para un evaluador en un ciclo. Body: `cycleId`, `evaluatorId`. | `TODO(auth:C7)` — evaluador mismo, o `rh`. |
| `GET` | `/api/v1/nine-box/matrices/:matrixId/entries` | Todas las entradas de una matriz. | `TODO(auth:C7)` — evaluador dueño, o `rh`. |
| `POST` | `/api/v1/nine-box/matrices/:matrixId/entries` | Crea o actualiza una entrada en la matriz (upsert por `evaluateeId`). Body: `evaluateeId`, `performanceScore(1-9)`, `potentialScore(1-9)`, `comments?`. | `TODO(auth:C7)` — evaluador dueño. |
| `PUT` | `/api/v1/nine-box/entries/:entryId` | Actualiza una entrada existente (scores + comentarios). Requiere `If-Match`. | `TODO(auth:C7)` — evaluador dueño. |
| `POST` | `/api/v1/nine-box/batch` | Submit batch de entradas 9×9. Body: `entries[{evaluateeId, performanceScore, potentialScore, comments?}]`. Atómico: todo o nada. | `TODO(auth:C7)` — evaluador dueño. |
| `GET` | `/api/v1/nine-box/scales` | Definiciones de escala: 9 niveles × 2 ejes (performance, potential). Cacheable con ETag. | `TODO(auth:C7)` — cualquier usuario autenticado. |
| `GET` | `/api/v1/nine-box/quadrants` | Definiciones de cuadrantes 1–9 (label, descripción, color, acción recomendada). Cacheable con ETag. | `TODO(auth:C7)` — cualquier usuario autenticado. |

### Proyecciones de respuesta

- **Listado (`GET /evaluations`)**: proyección ligera — `id`, `employeeId`, `cycleId`, `state`, `createdAt`, `updatedAt`. Sin joins pesados.
- **Detalle (`GET /evaluations/:id`)**: proyección completa con arrays anidados de `competencies` y `goals`.
- **Matriz (`GET /nine-box/matrices/:matrixId`)**: incluye `entries[]` con `quadrant`, `quadrantLabel`, `quadrantColor` (denormalizados desde `NineBoxQuadrant`).

## Error Handling

Todos los errores siguen el formato estándar:

```json
{
  "error": {
    "code": "EVALUATION_NOT_FOUND",
    "message": "La evaluación solicitada no existe.",
    "details": [
      "evaluation_id: 550e8400-e29b-41d4-a716-446655440000"
    ],
    "trace_id": "abc123-def456"
  }
}
```

### Códigos de error específicos

| Código | HTTP | Descripción |
|--------|------|-------------|
| `EVALUATION_NOT_FOUND` | `404` | La evaluación solicitada no existe. |
| `MATRIX_NOT_FOUND` | `404` | La matriz 9×9 solicitada no existe. |
| `ENTRY_NOT_FOUND` | `404` | La entrada de matriz solicitada no existe. |
| `EVALUATION_ALREADY_FINALIZED` | `409` | La evaluación ya fue cerrada; no se permiten más cambios. |
| `SELF_EVAL_DEADLINE_PASSED` | `409` | El período de autoevaluación ha finalizado (configurable por ciclo). |
| `INVALID_PHASE` | `409` | La acción requiere que el ciclo esté en fase `cierre`; el ciclo actual está en otra fase. |
| `QUADRANT_OUT_OF_RANGE` | `400` | Los scores de desempeño o potencial están fuera del rango 1–9. |
| `UNAUTHORIZED_EVALUATOR` | `403` | El usuario autenticado no es el evaluador dueño de esta matriz/entrada. |
| `CONCURRENT_UPDATE` | `409` | Conflicto de optimistic locking (`version` no coincide con `If-Match`). |
| `IDEMPOTENCY_KEY_CONFLICT` | `409` | El `Idempotency-Key` ya fue usado con un payload diferente. |
| `RATE_LIMIT_EXCEEDED` | `429` | Límite de requests por organización superado. |
| `REQUEST_TIMEOUT` | `408` | El contexto expiró antes de completar la operación. |
| `SERVICE_UNAVAILABLE` | `503` | Circuit breaker abierto por saturación de BD; reintentar con backoff. |

## Non-Goals (explicit)

- **No** gestión de metas del empleado (CRUD de metas, progreso, etc.) — eso es C4.
- **No** asignación de competencias al puesto o empleado — eso es C3.
- **No** notificaciones por email o push al completar evaluación — eso es C7.
- **No** autenticación, autorización ni RBAC completo — solo marcadores `TODO(auth:C7)` para integración futura.
- **No** reportes agregados across evaluadores (vista consolidada de empresa 9×9).
- **No** comparación histórica entre ciclos.
- **No** exportación a PDF, Excel u otro formato.
- **No** soporte para calificación 9×9 por parte de RH (la matriz es exclusiva del jefe/director/gerente).
- **No** rollback de evaluación finalizada. Si se requiere corrección, se hará por vía administrativa fuera de este API.

## Dependencies

| Cambio | Descripción | Estado requerido |
|--------|-------------|-----------------|
| **C1** | `data-model-core` — creación de tablas `Evaluation`, `EvaluationCompetency`, `EvaluationGoal`, `NineBoxMatrix`, `NineBoxEntry`, `NineBoxQuadrant`, `NineBoxScale` y sus índices. | **Aplicado primero** |
| **B1** | `evaluation-lifecycle` spec — definición de fases, estados de evaluación, reglas de transición. | **Finalizado** |
| **B5** | `manager-9x9` spec — definición de matriz, cuadrantes, escalas, reglas de cálculo. | **Finalizado** |
| **C2** | `evaluation-lifecycle-api` — validación de fase activa del ciclo. | **Aplicado primero** |
| **C4** | `goals-api` — verificación de estado de metas al cerrar evaluación. | **Aplicado primero** |
| **C5** | `org-hierarchy-api` — resolución de evaluador/evaluado, jerarquía orgánica. | **Aplicado primero** |
| **C7** | `auth-rbac` — middleware de autenticación y roles (este cambio lo referencia con `TODO(auth:C7)`). | Referencia futura |

## Success Criteria

- [ ] Todos los endpoints de evaluación y 9×9 documentados en esta propuesta están implementados en Go con Chi.
- [ ] **OpenAPI 3.1** generado y validado; tipos TypeScript generados en `web/src/lib/api/`.
- [ ] **Tests unitarios** para cada handler: casos felices, errores de validación, evaluación ya finalizada, fase inválida.
- [ ] **Tests de cuadrante**: todos los 9 cuadrantes computados correctamente para combinaciones de performance (1–9) × potential (1–9).
- [ ] **Tests de concurrencia**: al menos 50 goroutines concurrentes intentando submitir autoevaluación y RH evaluación sobre la misma evaluación; ninguna race condition, solo una transición exitosa, el resto recibe `CONCURRENT_UPDATE`.
- [ ] **Tests de carga**: `POST /api/v1/evaluations/:id/self-evaluation` sostiene **1000 concurrent submissions en < 30s** sin degradación de latencia > p95 500ms.
- [ ] **Tests de batch**: `POST /api/v1/nine-box/batch` procesa **20 entradas en < 300ms** p95.
- [ ] **Tests de idempotencia**: reenvío de evaluación con mismo `Idempotency-Key` retorna resultado idéntico sin crear duplicado.
- [ ] **Validación de fase**: todos los endpoints de evaluación y 9×9 rechazan operaciones con `INVALID_PHASE` si el ciclo no está en fase `cierre`.
- [ ] **Documentación**: `README.md` en `internal/service/evaluation/`, `internal/service/ninebox/`, `internal/handler/evaluation/` y `internal/handler/ninebox/` explicando el flujo de cierre, locking y cálculo de cuadrante.

---

*Generado por OpenSpec Proposal Generator para el proyecto SED Evaluación de Desempeño.*
