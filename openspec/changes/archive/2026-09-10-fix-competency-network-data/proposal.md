# Proposal: Fix Competency Network Data Loading

## Intent

La ruta `/evaluacion/9x9/competencias/{employeeId}` siempre muestra "No hay evaluaciones registradas para Empleado" porque `evaluationStore.load()` nunca se invoca desde la ruta, el store solo carga datos del usuario logueado, y no existe endpoint backend para obtener competencias de un empleado arbitrario.

## Scope

### In Scope
- Endpoint backend `GET /evaluations/employee/{employeeId}?cycle_id=...` que retorne competencias evaluadas
- Modificar `evaluationStore` para soportar carga por `employeeId` externo
- Invocar `load()` en `+page.ts` / `+layout.ts` de la ruta de competencias con el `employeeId` correcto
- Skeleton de carga mientras llegan los datos

### Out of Scope
- Cambios al modelo de datos o esquema `evaluation_competencies`
- Modificar `CompetencyNetworkView` (ya funciona correctamente con datos)
- Nuevos permisos RBAC (se reutilizan los existentes)
- Paginación o filtros adicionales

## Capabilities

### New Capabilities
- `employee-competency-ratings`: Endpoint y store para consultar las calificaciones de competencias de un empleado específico por ciclo

### Modified Capabilities
- `evaluation-lifecycle`: La ruta de competencias debe invocar carga de datos al montar; el store debe aceptar `employeeId` como parámetro

## Approach

1. **Backend**: Nuevo handler en Chi que consulta `evaluation_competencies` JOIN `competencies` JOIN `pillars` filtrado por `employee_id` y `cycle_id`. Retorna estructura compatible con `CompetencyNetworkView`.
2. **Store**: Extender `evaluationStore.load()` con parámetro opcional `employeeId`. Si se provee, llama al nuevo endpoint; si no, mantiene comportamiento actual (usuario logueado).
3. **Route**: `+page.ts` en `/evaluacion/9x9/competencias/[employeeId]` invoca `evaluationStore.load(employeeId)` en `onMount`.
4. **Skeleton**: Componente skeleton reutilizando patrón existente (ej. `CompetencyNetworkSkeleton`).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `api/internal/handler/` | New | Handler `GET /evaluations/employee/{employeeId}` |
| `api/internal/service/` | New | Servicio de consulta de competencias por empleado |
| `web/src/lib/stores/evaluationStore.ts` | Modified | `load()` acepta `employeeId` opcional |
| `web/src/routes/evaluacion/9x9/competencias/[employeeId]/+page.ts` | New/Modified | Invoca `load(employeeId)` en mount |
| `web/src/lib/components/` | New | Skeleton component para carga |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Endpoint expone datos de empleados sin autorización | Low | Reutilizar middleware RBAC existente; validar que el requester tiene acceso al empleado (jerarquía o rol RH) |
| Store rompe comportamiento actual de usuario logueado | Low | `employeeId` es opcional; path sin parámetro mantiene lógica actual |

## Rollback Plan

Revertir el commit del change. El endpoint nuevo es aditivo; no hay cambios a endpoints existentes. El store mantiene retrocompatibilidad con la firma original de `load()`.

## Dependencies

- `competency-framework` spec (pilares, competencias, niveles de aceptación ya definidos)
- `evaluation-lifecycle` spec (ciclos y fases existentes)
- Tabla `evaluation_competencies` ya poblada en BD

## Success Criteria

- [ ] `/evaluacion/9x9/competencias/{employeeId}` muestra las competencias evaluadas del empleado
- [ ] Skeleton visible durante carga de datos
- [ ] Endpoint responde en <200ms para un empleado con ~30 competencias
- [ ] Comportamiento actual de usuario logueado no se ve afectado
