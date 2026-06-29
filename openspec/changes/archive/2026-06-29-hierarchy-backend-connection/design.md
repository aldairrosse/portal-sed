## Context

El frontend tiene dos rutas de jerarquía:
- `/evaluacion/9x9/jerarquia`: usa `orgHierarchyStore` (dual-mode: fixtures en DEV, API en prod) pero sin filtrado por rol
- `/rh/jerarquia`: usa `rhHierarchyStore` (puramente client-side desde fixtures)

El backend tiene endpoints relevantes:
- `GET /evaluator-scopes?evaluatorId={me}`: retorna scope del evaluador (orgNodeIds[], employeeIds[])
- `GET /org-trees/{treeId}/nodes?format=nested`: retorna árbol completo anidado
- `GET /employees/{empId}/evaluatees`: retorna reportes directos
- `GET /org-nodes/{nodeId}/area-metrics`: handler existe pero NO está en spec OpenAPI

El `client.ts` del frontend solo tiene tipos para auth, cycle, goals y competency — falta `OrgHierarchyPaths`.

## Goals / Non-Goals

**Goals:**
- Consumir API real en ambas rutas de jerarquía (eliminando fixtures en producción)
- Filtrar nodos visibles según perfil del usuario usando evaluator-scopes
- Agregar `area-metrics` al spec OpenAPI y generar tipos TypeScript
- Manejar perfil `jefe` en vista de jerarquía (actualmente ignorado)

**Non-Goals:**
- Autenticación completa (asume dev persona existente)
- RBAC en backend (este change solo conecta frontend con backend existente)
- Modificación de estructura organizacional
- Exportación de métricas
- Histórico o tendencias

## Decisions

### D1: Generar tipos TypeScript desde OpenAPI

**Decisión:** Usar `openapi-typescript` para generar `org-hierarchy.d.ts` desde `org-hierarchy.yaml`

**Alternativas consideradas:**
- Tipos manuales: Requiere mantener sincronización manual, propenso a errores
- `openapi-typescript` (elegido): Genera tipos automáticamente desde spec, ya instalado en el proyecto

**Razón:** El proyecto ya usa `openapi-typescript` para otros schemas. Mantener un solo source of truth (OpenAPI) reduce drift.

### D2: Filtrado por rol en frontend usando evaluator-scopes

**Decisión:** Frontend llama `GET /evaluator-scopes?evaluatorId={me}` y usa `scopeData.orgNodeIds` para filtrar el árbol completo

**Flujo:**
1. Frontend obtiene el árbol completo con `GET /org-trees/{treeId}/nodes?format=nested`
2. Frontend obtiene el scope del usuario con `GET /evaluator-scopes?evaluatorId={me}`
3. Frontend filtra nodos del árbol según `orgNodeIds` del scope

**Alternativas consideradas:**
- Backend retorna árbol filtrado: Requiere nuevo endpoint o parámetros, más complejo
- Frontend filtra (elegido): Reutiliza endpoints existentes, lógica simple en store

**Razón:** El endpoint `evaluator-scopes` ya existe y retorna los nodos que el usuario puede ver. Filtrar en frontend es más simple y no requiere cambios en backend.

### D3: Manejo de perfil jefe

**Decisión:** Para perfil `jefe`, usar endpoint `GET /employees/{empId}/evaluatees` para obtener reportes directos, NO el árbol completo

**Flujo:**
1. Jefe llama `GET /employees/{me}/evaluatees` para obtener sus reportes directos
2. Jefe ve tabla de reportes directos (no árbol jerárquico)

**Razón:** El spec `org-hierarchy` define que jefes ven "mis evaluados" como lista, no como árbol. El endpoint `/evaluatees` ya existe y retorna exactamente lo necesario.

### D4: Consumo de area-metrics

**Decisión:** Frontend llama `GET /org-nodes/{nodeId}/area-metrics?cycleId={cycleId}` para métricas agregadas

**Flujo:**
1. Usuario selecciona nodo en árbol
2. Frontend llama `area-metrics` con nodeId y cycleId
3. Backend retorna métricas agregadas (avg progress, completed goals, etc.)

**Razón:** El handler ya existe en backend. Solo necesita documentación OpenAPI y tipos generados.

## Risks / Trade-offs

**[R1] Performance al cargar árbol completo** → Mitigación: El árbol completo se carga una vez y se cachea en el store. El filtrado por rol es operación local (sin llamadas adicionales).

**[R2] Fallo en evaluator-scopes** → Mitigación: Si falla, mostrar EmptyState con error. No asumir scope vacío.

**[R3] Desfase entre spec OpenAPI y handler** → Mitigación: Verificar que el handler `GetAreaMetrics` cumple el contract antes de agregar al spec.

**[R4] Perfil jefe sin jerarquía visual** → Mitigación: Mostrar tabla de reportes directos en vez de árbol. Es consistente con el spec existente.

## Migration Plan

1. Generar tipos TypeScript desde OpenAPI
2. Agregar `OrgHierarchyPaths` al `client.ts`
3. Modificar `orgHierarchyStore` para filtrar por rol
4. Modificar `rhHierarchyStore` para consumir `area-metrics`
5. Agregar manejo de perfil `jefe` en rutas
6. Verificar que `pnpm run check` pasa

**Rollback:** Revertir cambios en stores; los fixtures siguen disponibles en DEV.
