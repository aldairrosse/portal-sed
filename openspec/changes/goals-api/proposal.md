# Propuesta: C4 — goals-api

## 1. Intent

Implementar la REST API para la gestión de objetivos (goals) del portal SED, incluyendo CRUD de categorías, objetivos, KPIs, validación de ponderación (doble 100%) y restricciones de edición basadas en fases del ciclo de evaluación. Este es el núcleo de la API orientada a empleados: todo colaborador interactúa con ella para definir, revisar y actualizar sus objetivos durante el ciclo.

**Especificación base:** `specs/goals-and-weighting.md`

**Decisiones de arquitectura referenciadas:**
- **#1 (Double 100%):** Las categorías de objetivos de un empleado deben sumar 100% de peso, y los objetivos dentro de cada categoría deben sumar 100% de peso.
- **#5 (Categorías personalizadas):** Cada empleado define sus propias categorías; no hay catálogo global de categorías.
- **#6 (KPIs reutilizables):** Los KPIs son indicadores compartidos que se vinculan a uno o más objetivos (relación N:M).
- **#7 (RH también tiene objetivos):** Los empleados de Recursos Humanos también tienen objetivos propios, igual que cualquier otro empleado.
- **#8 (Reglas de edición por jerarquía):** Los jefes pueden visualizar y solicitar cambios, pero NO crear, editar ni eliminar objetivos de sus subordinados.

## 2. Scope

### In Scope

- **CRUD de Categorías de Objetivos** (`GoalCategory`) en el contexto de un empleado.
- **CRUD de Objetivos** (`Goal`) dentro de categorías, con validación de peso.
- **CRUD de KPIs** (`KPI`) como catálogo compartido reutilizable.
- **Vinculación/desvinculación** de KPIs a objetivos (N:M via `GoalKpiLink`).
- **Endpoint de validación de ponderación** (`validate-weights`) que verifica doble 100% antes de guardar.
- **Seguimiento de progreso** de objetivos: actualización de `currentValue` durante la fase de avance.
- **Asignación de objetivos** por empleado y ciclo (`GoalAssignment`).
- **Restricciones de fase:** operaciones bloqueadas según la fase actual del ciclo de evaluación.
- **Especificación OpenAPI 3.1** como contrato de verdad para todos los endpoints.

### Out of Scope

- **Evaluación final y scoring** (cubierto por C6: evaluation-scoring-api).
- **Gestión de competencias** (cubierto por C3: competencies-api).
- **Autenticación y autorización** (cubierto por C7: auth-api). Se asume que los middlewares de auth inyectan el `employeeId` y `roles` en el contexto.
- **Notificaciones** por correo o push (cubierto por C8: notifications-api).
- **Importación/exportación** Excel de objetivos.

## 3. Concurrency & High-Volume Design

La API de objetivos es la de mayor tráfico del sistema: todo empleado la consulta al menos mensualmente durante el ciclo. El diseño debe soportar miles a millones de peticiones.

### Database Connection Management

- **Pool de conexiones `pgx`:** `MaxConns=30`, `MinConns=8`, priorizando este servicio por ser el más transitado.
- **Réplicas de lectura:** Para listados de objetivos y catálogo de KPIs. Las escrituras siempre van al nodo primario.
- **Cliente Ent por request:** No se comparte estado entre requests; el cliente se inyecta desde el pool.

### Concurrency Control

- **Optimistic locking en Goal:** Campo `version` (integer) para prevenir actualizaciones concurrentes del mismo empleado. Se incrementa en cada escritura.
- **Validación de peso en transacción:** Las operaciones que modifican peso usan `SELECT SUM(weight) ...` dentro de la transacción para evitar lecturas obsoletas (race conditions entre dos requests simultáneas del mismo empleado).
- **Pessimistic locking en creación de objetivos:** `SELECT FOR UPDATE` sobre la categoría al crear un objetivo, para prevenir que la suma de pesos exceda 100% por concurrencia.
- **Advisory locks en GoalAssignment:** `pg_advisory_lock` usando `hashtext(employee_id || cycle_id)` para evitar creación duplicada de asignación por race condition.
- **Idempotency keys:** En creación y actualización de objetivos (`Idempotency-Key` header). Los keys se almacenan en Redis con TTL 24h para deduplicación.

### High-Volume Patterns

- **Cursor pagination:** Todos los listados usan cursor-based pagination (`cursor`, `limit`) en lugar de offset, para evitar degradación en listados grandes.
- **Batch endpoint:** `POST /api/v1/goals/batch` permite crear/actualizar múltiples objetivos de forma atómica en una sola transacción, reduciendo N round trips a 1.
- **Bulk validation:** El endpoint `validate-weights` ejecuta una única query agregada para validar todas las categorías de un empleado, en lugar de N queries.
- **Redis cache para catálogo de KPIs:** El listado de KPIs (compartido, raramente modificado) se cachea en Redis con TTL 5 minutos. Invalidación por write-through al crear/actualizar un KPI.
- **ETag en listados de objetivos:** El endpoint `GET /api/v1/employees/:empId/categories` incluye `ETag` basado en el `updated_at` más reciente de cualquier objetivo del empleado. Permite al cliente detectar ediciones concurrentes y evitar re-fetch innecesario (`304 Not Modified`).
- **Rate limiting:** 200 req/s lecturas, 50 req/s escrituras por empleado (implementado vía middleware compartido, pero este servicio define los límites más permisivos por ser el de mayor uso).
- **Timeouts:** `5s` para lecturas, `15s` para escrituras (la validación de ponderación es computacionalmente más compleja que un simple INSERT).
- **Streaming para exportaciones grandes:** Si en el futuro se agrega exportación, el endpoint de asignaciones debe soportar streaming de JSON para evitar materializar todo en memoria.

### Transaction Design

| Operación | Nivel de aislamiento | Notas |
|---|---|---|
| Crear/actualizar objetivo | `Read Committed` + `SELECT FOR UPDATE` sobre categoría | Previene desbordamiento de peso por concurrencia |
| Guardar categoría + objetivos | `Read Committed` | Atomicidad para doble 100% |
| Validación de peso | `Repeatable Read` | Garantiza que la suma no cambia durante la validación |
| Actualizar progreso | `Read Committed` | Transacción ligera: solo `currentValue` + verificación de estado |
| Batch de objetivos | `Read Committed` | Una transacción atómica para todo el batch |

## 4. API Endpoints

### Categorías de Objetivos

| Método | Ruta | Descripción | Fases permitidas |
|---|---|---|---|
| `GET` | `/api/v1/employees/{empId}/categories` | Lista categorías con objetivos anidados | Todas |
| `POST` | `/api/v1/employees/{empId}/categories` | Crea una categoría | `asignacion` |
| `PUT` | `/api/v1/employees/{empId}/categories/{catId}` | Actualiza nombre y peso de categoría | `asignacion`, `avance` (solo peso si no cambia objetivos) |
| `DELETE` | `/api/v1/employees/{empId}/categories/{catId}` | Elimina categoría (solo si está vacía o se permite cascada según fase) | `asignacion` |

### Objetivos

| Método | Ruta | Descripción | Fases permitidas |
|---|---|---|---|
| `POST` | `/api/v1/employees/{empId}/categories/{catId}/goals` | Crea un objetivo en una categoría | `asignacion` |
| `PUT` | `/api/v1/goals/{goalId}` | Actualiza un objetivo (nombre, descripción, unidad, peso, targetValue) | `asignacion` (todos los campos), `avance` (solo campos permitidos) |
| `DELETE` | `/api/v1/goals/{goalId}` | Elimina un objetivo | `asignacion` |
| `PATCH` | `/api/v1/goals/{goalId}/progress` | Actualiza `currentValue` (progreso del objetivo) | `avance` |
| `POST` | `/api/v1/goals/batch` | Crea/actualiza múltiples objetivos atómicamente | `asignacion` |

### Validación de Ponderación

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/v1/employees/{empId}/validate-weights` | Valida doble 100% para las categorías y objetivos del empleado |

### KPIs (Catálogo compartido)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/kpis` | Lista KPIs (paginado, con cache) |
| `POST` | `/api/v1/kpis` | Crea un KPI (admin/RH) |
| `PUT` | `/api/v1/kpis/{kpiId}` | Actualiza un KPI |
| `DELETE` | `/api/v1/kpis/{kpiId}` | Elimina un KPI (solo si no está vinculado) |

### Vinculación KPI ↔ Objetivo

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/v1/goals/{goalId}/kpis` | Vincula un KPI a un objetivo |
| `DELETE` | `/api/v1/goals/{goalId}/kpis/{kpiId}` | Desvincula un KPI de un objetivo |

### Asignación de Objetivos

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/employees/{empId}/assignments` | Obtiene la asignación del empleado para el ciclo activo |
| `POST` | `/api/v1/employees/{empId}/assignments` | Crea o finaliza la asignación de objetivos para el ciclo |

### Notas de diseño de endpoints

- **Cursores de paginación:** Todos los `GET` que devuelven listas usan `?cursor=...&limit=...` (default `limit=20`, max `limit=100`).
- **Nesting de recursos:** Las categorías se listan con objetivos anidados para reducir round trips, pero la creación de objetivos es un endpoint separado para mantener REST claro.
- **`empId` en path:** Aunque el empleado autenticado normalmente accede a sus propios objetivos, se usa `empId` para permitir que jefes visualicen objetivos de subordinados (con restricciones de escritura según decisión #8).
- **Batch endpoint:** El body acepta un array de objetos con `operation: create | update` y los datos del objetivo. La transacción es atómica: si falla la validación de peso de un elemento, falla todo el batch.

## 5. Error Handling

Todas las respuestas de error siguen el esquema estándar del proyecto:

```json
{
  "error": {
    "code": "GOAL_NOT_FOUND",
    "message": "El objetivo solicitado no existe",
    "details": [
      {
        "field": "goalId",
        "value": "123e4567-e89b-12d3-a456-426614174000",
        "issue": "not_found"
      }
    ],
    "trace_id": "abc123"
  }
}
```

### Códigos de error específicos de este dominio

| Código | HTTP Status | Descripción |
|---|---|---|
| `CATEGORY_NOT_FOUND` | `404` | La categoría no existe para el empleado indicado |
| `GOAL_NOT_FOUND` | `404` | El objetivo no existe |
| `KPI_NOT_FOUND` | `404` | El KPI no existe en el catálogo |
| `WEIGHT_SUM_INVALID` | `422` | La suma de pesos no alcanza 100% ± 0.01. Incluye `details` con el valor actual y esperado |
| `PHASE_RESTRICTED` | `403` | La operación no está permitida en la fase actual del ciclo |
| `DUPLICATE_CATEGORY_NAME` | `409` | Ya existe una categoría con ese nombre para el empleado (violación de unique) |
| `INVALID_WEIGHT_RANGE` | `400` | El peso debe estar entre 0 y 100 |
| `GOAL_NOT_DELETABLE_IN_PHASE` | `403` | No se puede eliminar el objetivo en la fase actual (más específico que `PHASE_RESTRICTED`) |
| `KPI_LINKED_CANNOT_DELETE` | `409` | No se puede eliminar un KPI que está vinculado a objetivos |
| `CONCURRENT_MODIFICATION` | `409` | El recurso fue modificado por otro proceso (optimistic locking, version mismatch) |
| `IDEMPOTENCY_KEY_REUSE` | `409` | El `Idempotency-Key` ya fue usado con un payload diferente |
| `RATE_LIMIT_EXCEEDED` | `429` | Límite de requests superado (200 lecturas / 50 escrituras) |

### Criterios de `WEIGHT_SUM_INVALID`

El error `WEIGHT_SUM_INVALID` debe incluir en `details` una estructura que permita al frontend mostrar exactamente dónde falla la validación:

```json
{
  "details": [
    {
      "type": "category_sum",
      "expected": 100.0,
      "actual": 95.0,
      "categories": [ ... ]
    },
    {
      "type": "goal_sum",
      "category_id": "...",
      "category_name": "Crecimiento",
      "expected": 100.0,
      "actual": 110.0,
      "goals": [ ... ]
    }
  ]
}
```

## 6. Non-Goals

- **No se implementa evaluación final ni scoring:** La calificación de objetivos (cuánto se cumplió, scoring ponderado) es responsabilidad de C6.
- **No se implementa gestión de competencias:** C3 cubre las competencias por separado.
- **No se implementa autenticación ni RBAC:** C7 provee el middleware que inyecta `employeeId` y `roles` en el contexto de Chi. Este change asume que esa información ya está disponible.
- **No se implementan notificaciones:** C8 se encarga de enviar emails/push cuando se crean o solicitan cambios de objetivos.
- **No se implementa importación/exportación Excel:** Esto es un feature futuro que requiere su propio change y diseño de parsing de archivos.

## 7. Dependencies

- **C1: data-model-core** — Proporciona el esquema Ent (`Goal`, `GoalCategory`, `KPI`, `GoalKpiLink`, `GoalAssignment`) e índices de base de datos.
- **B3: goals-and-weighting** — Especificación funcional que define las reglas de negocio de doble 100%, fases de edición, y jerarquía de edición.
- **C2: evaluation-lifecycle-api** — Se consulta para determinar la fase actual del ciclo (`asignacion`, `avance`, `cierre`). El goals-api valida operaciones contra esta fase.

## 8. Success Criteria

- [ ] Todos los endpoints CRUD + validación están implementados y cubiertos por tests unitarios.
- [ ] La validación de doble 100% se ejecuta en tiempo real en el endpoint `validate-weights` y también se aplica como guard en operaciones de escritura que afectan peso.
- [ ] Las restricciones de fase se aplican correctamente: errores `PHASE_RESTRICTED` o `GOAL_NOT_DELETABLE_IN_PHASE` cuando se intentan operaciones no permitidas.
- [ ] El campo `version` de `Goal` se incrementa en cada escritura y se rechazan requests con version obsoleta (`409 CONCURRENT_MODIFICATION`).
- [ ] El endpoint `POST /api/v1/goals/batch` procesa 50 objetivos en menos de 500ms (incluyendo validación de peso) en entorno de staging.
- [ ] El endpoint `GET /api/v1/kpis` responde en menos de 50ms con cache activo (Redis) y menos de 200ms con cache cold (query a BD).
- [ ] Load test: 200 req/s sostenidos en lecturas, 50 req/s en escrituras, con p95 < 300ms y 0% de errores de validación de peso por race conditions.
- [ ] La especificación OpenAPI 3.1 está completa, validada (`openapi validate`) y genera tipos TypeScript correctamente.
- [ ] Tests de integración con PostgreSQL 15+ que verifican comportamiento concurrente (optimistic locking, suma de pesos, advisory locks).
