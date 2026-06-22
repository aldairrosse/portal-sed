# Proposal: Backend API para métricas de jerarquía RRHH

## Change

`backend-rh-hierarchy-metrics`

## Intent

El frontend `rhHierarchyStore.svelte.ts` actualmente lee datos de fixtures JSON estáticos (`goals.json`, `assignments.json`, `rh-evaluations.json`). Necesitamos un endpoint backend que devuelva las mismas métricas agregadas pero consultando la base de datos real.

## Scope

### Incluido

- Nuevo endpoint `GET /api/v1/org-nodes/{nodeId}/area-metrics` que retorna:
  - Cantidad de empleados en el nodo (solo nivel directo, NO subárbol)
  - Cantidad de empleados con al menos una meta válida
  - Promedio de avance de metas (excluyendo targetValue === 0 y empleados sin metas)
  - Metas completadas vs pendientes
  - Promedio de calificación RH (solo empleados con al menos una evaluación RH)
  - Cantidad de evaluaciones RH
  - Lista de empleados del nodo (id, nombre, puesto, perfil)

- Handler, Service, Repository, DTO para el nuevo endpoint
- Query SQL que busca empleados cuyo `org_node_id` coincide con el nodeId (sin recursión)

### Scope de métricas: solo nivel directo

Las métricas se calculan **únicamente sobre los empleados cuyo `org_node_id` coincide con el `nodeId` solicitado**, incluyendo al manager del área (el empleado dueño del nodo). **NO se incluyen subnodos ni subordinados递归.**

**Ejemplo — Area de Capacitación (nodeId = emp-director-capacitacion):**
- Empleados directos: Director del área + 2 colaboradores + 1 jefe de bienvenida
- Promedio: agrega avance/calificación de esos 4 empleados únicamente
- **NO incluye** al personal debajo del jefe de bienvenida

**Ejemplo — Area de Bienvenida (nodeId = emp-jefe-bienvenida):**
- Empleados directos: Jefe de bienvenida + 4 colaboradores de ese nodo
- Promedio: agrega avance/calificación de esos 5 empleados únicamente

Esto simplifica la query SQL (sin CTE recursiva) y es más performante.

### No incluido

- Modificar el frontend (se hace en un cambio separado)
- Modificar endpoints existentes de org-hierarchy
- Nuevos campos en el schema Ent existente

## Why

El frontend actualmente hardcodea fixtures JSON. Para producción, necesita consultar datos reales. El endpoint existente `GET /api/v1/org-nodes/{nodeId}` devuelve datos del nodo pero no métricas agregadas de sus empleados. Necesitamos un endpoint dedicado que haga la agregación en el servidor (más eficiente que traer todos los datos al frontend y calcular ahí).

## Approach

### Nuevo endpoint

```
GET /api/v1/org-nodes/{nodeId}/area-metrics
```

**Query params opcionales:**
- `cycleId` (UUID) — ciclo a consultar. Si no se provee, usa el ciclo activo de la organización.

**Response:**
```json
{
  "nodeId": "emp-director-01",
  "employeeCount": 4,
  "employeesWithGoals": 3,
  "avgProgress": 68.3,
  "completedGoals": 5,
  "pendingGoals": 7,
  "avgRating": 3.5,
  "ratingsCount": 3,
  "employees": [
    {
      "id": "emp-director-01",
      "name": "Carmen Jiménez Castro",
      "position": "Director",
      "profileId": "director"
    },
    {
      "id": "emp-colaborador-01",
      "name": "Ana Pérez López",
      "position": "Colaborador",
      "profileId": "colaborador"
    }
  ]
}
```

### Lógica de agregación (Server-side)

1. **Empleados del nodo**: Query simple: `SELECT * FROM employees WHERE org_node_id = $1 AND is_active = true`. Sin recursión.
2. **Metas**: Join `employees` → `goal_assignments` → `goal_categories` → `goals` filtrado por cycle_id. Calcular progress = (current_value / target_value) * 100, excluyendo target_value = 0
3. **Evaluaciones RH**: Join `employees` → `evaluations` → `evaluation_competencies` donde el competency tiene un `rh_rating` no nulo
4. **Manager del área**: El empleado cuyo `org_node_id` es el nodeId es incluido en la lista (es el lider del area)

### Capas

| Capa | Archivo | Responsabilidad |
|------|---------|-----------------|
| Handler | `internal/handler/org/handler.go` (extender) | Parsear requestId, llamar service |
| Service | `internal/service/org/metrics_service.go` (nuevo) | Orquestar queries, calcular agregaciones |
| Repository | `internal/repository/org/metrics_repo.go` (nuevo) | Queries sobre org_nodes + goals + evaluations (scope directo, sin recursión) |
| DTO | `internal/dto/org/metrics_dto.go` (nuevo) | Request/Response types |

### Query SQL clave (sin recursión)

```sql
-- 1. Empleados directos del nodo
SELECT e.id, e.first_name, e.last_name, e.profile_id
FROM employees e
WHERE e.org_node_id = $1
  AND e.is_active = true

-- 2. Metas de esos empleados (filtrado por cycle_id)
SELECT g.*,
       ga.employee_id,
       (g.current_value::float / NULLIF(g.target_value, 0)) * 100 AS progress_pct
FROM goals g
JOIN goal_categories gc ON gc.id = g.category_id
JOIN employees emp ON emp.id = gc.employee_id
LEFT JOIN goal_assignments ga ON ga.employee_id = emp.id AND ga.cycle_id = $2
WHERE emp.org_node_id = $1
  AND emp.is_active = true
  AND g.target_value > 0

-- 3. Evaluaciones RH de esos empleados
SELECT ec.rh_rating, ec.employee_id
FROM evaluation_competencies ec
JOIN evaluations e ON e.id = ec.evaluation_id
JOIN employees emp ON emp.id = e.employee_id
WHERE emp.org_node_id = $1
  AND emp.is_active = true
  AND ec.rh_rating IS NOT NULL
  AND e.cycle_id = $2
```

### Decisiones técnicas

- **Reutilizar handler existente**: Extender `OrgHandler` con el nuevo método, no crear un handler nuevo
- **Service separado**: Crear `metrics_service.go` para no ensuciar el service de org existente
- **Sin cache por ahora**: La métrica se calcula on-demand. Cache se puede agregar después si hay problemas de performance
- **Cycle por defecto**: Si no se pasa `cycleId`, buscar el ciclo activo de la organización del empleado

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Muchos JOINs en una sola query | Low | Usar CTEs separadas, agregar índices compuestos |
| Sin autenticación RBAC clara para este endpoint | Low | Requerir profile `rh` en el middleware, seguir patrón existente |
| Performance con muchos empleados por nodo | Low | El scope es directo (no recursivo), así que el volumen es acotado |

## Out of scope

- Frontend migration de fixtures a API (cambio separado)
- Cache de métricas
- Export de métricas
- Filtros avanzados (solo un area específica, solo empleados con metas, etc.)
