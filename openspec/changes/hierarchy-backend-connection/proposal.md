## Why

Las vistas de jerarquía (`/rh/jerarquia` y `/evaluacion/9x9/jerarquia`) usan datos fixture. El backend ya tiene endpoints relevantes (`/evaluator-scopes`, `/org-trees/{treeId}/nodes`, `/employees/{empId}/evaluatees`) pero el frontend no los consume. Además, `area-metrics` existe como handler pero no está en el spec OpenAPI. Necesitamos conectar frontend con backend y agregar filtrado por rol para que cada perfil vea solo su alcance.

## What Changes

- Agregar `GET /org-nodes/{nodeId}/area-metrics` al spec OpenAPI `org-hierarchy.yaml`
- Agregar tipo `OrgHierarchyPaths` al `client.ts` del frontend para consumir endpoints de jerarquía
- Modificar `orgHierarchyStore` para consumir API real y filtrar por rol usando `evaluator-scopes`
- Modificar `rhHierarchyStore` para consumir `area-metrics` del backend en vez de fixtures
- Agregar manejo de perfil `jefe` en vista de jerarquía (actualmente no tiene manejo especial)
- Eliminar dependencia de fixtures en producción para ambas rutas

## Capabilities

### New Capabilities

- `hierarchy-api-contract`: Definición OpenAPI de `area-metrics` y tipos TypeScript generados para consumo del frontend
- `role-based-hierarchy-filter`: Filtrado de nodos visibles según perfil del usuario (DG/RH ven todo, Director ve su subárbol, Jefe ve reportes directos)

### Modified Capabilities

- `org-hierarchy`: Agregar requirement de consumo de API y filtrado por rol
- `rh-org-hierarchy-metrics`: Cambiar de fixtures a API real para métricas agregadas

## Impact

**Backend:**
- `api/openapi/org-hierarchy.yaml`: Agregar schema `AreaMetrics` y endpoint `GET /org-nodes/{nodeId}/area-metrics`
- `api/internal/handler/org/org_handler.go`: Ya existe `GetAreaMetrics`, solo necesita documentación OpenAPI

**Frontend:**
- `web/src/lib/api/client.ts`: Agregar `OrgHierarchyPaths` al tipo `AppPaths`
- `web/src/lib/api/schemas/org-hierarchy.d.ts`: Generar tipos desde OpenAPI (nuevo archivo)
- `web/src/lib/stores/orgHierarchyStore.svelte.ts`: Agregar filtrado por rol usando evaluator-scopes
- `web/src/lib/stores/rhHierarchyStore.svelte.ts`: Cambiar de fixtures a llamada API
- `web/src/routes/evaluacion/9x9/jerarquia/+page.svelte`: Manejar perfil jefe
- `web/src/routes/rh/jerarquia/+page.svelte`: Consumir métricas reales

**Dependencias:**
- `openapi-typescript`: Ya instalado, generar tipos con `openapi-typescript api/openapi/org-hierarchy.yaml -o web/src/lib/api/schemas/org-hierarchy.d.ts`

**Non-Goals (for this change):**
- Autenticación y RBAC completos (asume dev persona existente)
- Modificación de la estructura organizacional desde UI
- Exportación de métricas (CSV/PDF)
- Histórico de métricas o tendencias
