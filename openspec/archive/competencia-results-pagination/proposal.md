# Proposal: Resultados de competencias — Paginación, búsqueda y filtro de equipo

## Intent

La vista "Resultados de competencias" (`/evaluacion/9x9/competencias`) carga todos los empleados del scope en el cliente y calcula promedios de competencias en el navegador. Esto no escala: a medida que crece el catálogo de empleados, la página se vuelve lenta y no ofrece búsqueda ni paginación.

Este change aplica el patrón OFFSET/LIMIT de `rhEvaluadosStore` + "Evaluaciones RH" a esta vista, agrega búsqueda server-side y un filtro "mi equipo" (excluyendo al usuario actual), siguiendo el patrón del selector de empleados de "Asignación de Metas".

## What Changes

### Backend

1. **Nuevo endpoint** `GET /api/v1/evaluations/competency-results?cycle_id=&offset=&limit=&q=&scope=`
   - `scope` = `all` (default) | `team` (excluye usuario actual)
   - Retorna por cada empleado: id, nombre, perfil, promedio autoevaluación, promedio RH, estado
   - Los promedios se calculan en BD con `AVG(self_rating)` y `AVG(rh_rating)` sobre `evaluation_competencies` JOIN `evaluations`
   - Estado se deriva en backend: `completada` (ambos ratings presentes), `autoevaluacion` (solo self), `pendiente` (solo rh), `sin-datos` (ninguno)
2. **Repository**: `ListCompetencyResults(cycleID, query, scope, currentUserId, offset, limit)` + `CountCompetencyResults(cycleID, query, scope, currentUserId)` — método COUNT separado
3. **DTO response**: `{ data: [...], meta: { hasMore, total, offset, limit } }` donde `hasMore = offset + len(rows) < total`
4. **Handler**: parsea `offset`, `limit`, `q`, `scope` de query string; clampa limit 1-200, default 50

### Frontend Store

Nuevo `competencyResultsStore.svelte.ts` — misma estructura que `rhEvaluadosStore`:
- `$state` a nivel de módulo: `items`, `loading`, `error`, `currentPage`, `apiTotal`, `hasMore`, `hasPrev`, `currentQ`, `scopeFilter`
- `load()` lee `currentPage`, `currentQ`, `scopeFilter` del estado, llama `GET /evaluations/competency-results`
- `next()`/`prev()` incrementan/decrementan `currentPage`, llaman `load()`
- `search(query)` con debounce 300ms, resetea `currentPage = 0`
- `setScope(scope)` cambia filtro `all`/`team`, resetea página
- `cycleId` se obtiene del contexto activo (mismo patrón que otras vistas 9x9)

### Frontend UI

Reescribir `evaluacion/9x9/competencias/+page.svelte` siguiendo el patrón de `rh/evaluaciones/+page.svelte`:
- Header: "Resultados de competencias" + descripción
- Barra: input de búsqueda (SIN `disabled={loading}`) + toggle "Mi equipo" + "Viendo X de Y empleados" / "Viendo X resultado(s)" si hay búsqueda activa
- Controles de paginación: Anterior / Página N / Siguiente — siempre visibles fuera del condicional de loading
- Skeleton SOLO en carga inicial: `{#if loading && items.length === 0}`
- Tabla con las mismas 6 columnas: Empleado, Perfil, Autoevaluación, RH, Estado, Acción
- `titleCase()` para labels de perfil
- Link de acción apunta a `/evaluacion/9x9/competencias/{empId}` (sin cambio)

### Filtro "Mi equipo"

- Toggle en la barra de controles, activo por defecto OFF
- Cuando está ON: envía `scope=team` al backend
- Backend excluye el `employee_id` del usuario actual (obtenido de la sesión)
- Para perfil `jefe`: filtra a solo reportes directos del manager (excluyendo self)
- Para perfil `director`/`director-general`/`rh`: filtra a todos los descendientes o全体员工 (excluyendo self)
- Referencia: `GET /employees/{empId}/team` en `evaluatee_service.go` ya resuelve membresía de equipo

### Validación de columnas vs BD/API

| Columna UI | Fuente actual (cliente) | Fuente nueva (BD via API) |
|---|---|---|
| Empleado | `orgHierarchyStore` → `node.name` | `employees.first_name`, `employees.last_name` |
| Perfil | `orgHierarchyStore` → `node.profileId` → `PROFILE_LABELS` | `evaluation_profiles.name` (JOIN) |
| Autoevaluación | `evaluationStore` → avg de `selfRating` | `AVG(ec.self_rating)` WHERE NOT NULL |
| RH | `evaluationStore` → avg de `rhRating` | `AVG(ec.rh_rating)` WHERE NOT NULL |
| Estado | Derivado en cliente (lógica if/else) | Derivado en backend (CASE WHEN) |
| Acción | Link a detalle | Sin cambio — link a `/evaluacion/9x9/competencias/{id}` |

## Scope Boundaries

- **In scope**: Backend endpoint paginado, frontend store, reescritura UI, filtro "mi equipo", validación de columnas
- **Out of scope**: Cambios a `rhEvaluadosStore`, cambios a "Mis Evaluados", cambios a la página de detalle de competencias por empleado, nuevas columnas en la tabla, exportación, ordenamiento por columna

## Chained PR Plan

| PR | Layer | Content |
|----|-------|---------|
| 1 | Backend | Repository + Service + Handler para `GET /evaluations/competency-results` con OFFSET/LIMIT, búsqueda, scope |
| 2 | Frontend Store | `competencyResultsStore.svelte.ts` con patrón offset/limit y filtro de scope |
| 3 | Frontend UI | Reescritura de `+page.svelte` con búsqueda, paginación, toggle "mi equipo", skeleton condicional |

## Risks

| Riesgo | Probabilidad | Mitigación |
|--------|-------------|------------|
| `AVG()` sobre muchos rows por empleado puede ser lento | Media | Índice compuesto en `evaluation_competencies(evaluation_id)` + filtro por `cycle_id` vía JOIN a `evaluations` |
| La columna `self_rating`/`rh_rating` puede no existir en el schema Ent actual | Baja | Verificar migraciones existentes; el SQL crudo en `metrics_repo.go` ya las usa — si Ent no las expone, usar query crudo |
| El toggle "mi equipo" puede confundir si el usuario no entiende qué excluye | Baja | Tooltip corto: "Excluye tu propio registro de la lista" |

## Open Questions

None — el patrón está bien establecido por `rhEvaluadosStore` y el endpoint `GET /employees/{empId}/team`.
