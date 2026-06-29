## MODIFIED Requirements

### Requirement: API de jerarquía real

El sistema SHALL consumir la API real de jerarquía en producción. En desarrollo sin `VITE_USE_API`, se usarán fixtures. El frontend SHALL llamar `GET /org-trees/{treeId}/nodes?format=nested` para obtener el árbol completo.

#### Scenario: Producción consume API real

- GIVEN `VITE_USE_API=true` o entorno de producción
- WHEN `orgHierarchyStore.load()` se ejecuta
- THEN se llama `GET /org-trees?type=corporate` para descubrir el árbol
- AND se llama `GET /org-trees/{treeId}/nodes?format=nested&depth=-1` para obtener nodos
- AND se llama `GET /evaluator-scopes?evaluatorId={currentUserId}` para obtener alcance
- AND el árbol se filtra según `scopeData.orgNodeIds`
- AND el árbol se renderiza con datos del backend

#### Scenario: Desarrollo usa fixtures

- GIVEN `import.meta.env.DEV && !import.meta.env.VITE_USE_API`
- WHEN `orgHierarchyStore.load()` se ejecuta
- THEN se carga fixture desde `$lib/fixtures/org-hierarchy/org-tree.json`
- AND NO se realizan llamadas a la API
- AND NO se aplica filtrado por rol

### Requirement: Consumo de OrgHierarchyPaths en client.ts

El frontend SHALL importar `OrgHierarchyPaths` desde `./schemas/org-hierarchy.d.ts` y agregarlo al tipo `AppPaths` en `client.ts`.

#### Scenario: Client tiene tipos de jerarquía

- WHEN `client.ts` se importa en cualquier store o componente
- THEN `client.GET('/org-trees')` está tipado correctamente
- AND `client.GET('/org-trees/{treeId}/nodes')` está tipado correctamente
- AND `client.GET('/evaluator-scopes')` está tipado correctamente
- AND `client.GET('/employees/{empId}/evaluatees')` está tipado correctamente
- AND `client.GET('/org-nodes/{nodeId}/area-metrics')` está tipado correctamente

### Requirement: Alcance de "mis evaluados"

Cada persona con subordinados directos SHALL ver una lista de "mis evaluados" con acceso a la información de fijación y seguimiento de metas de sus subordinados en inicio y medio año. El frontend SHALL usar `GET /employees/{empId}/evaluatees` para obtener reportes directos.

#### Scenario: Jefe ve mis evaluados

- GIVEN jefe con 3 colaboradores
- WHEN navega a "Mis evaluados"
- THEN se llama `GET /employees/{me}/evaluatees`
- AND ve lista con nombre, puesto y estado de evaluación de cada colaborador
- AND cada fila es seleccionable para ver detalle

#### Scenario: Colaborador sin mis evaluados

- GIVEN perfil `colaborador` sin subordinados
- WHEN navega a la aplicación
- THEN NO ve el ítem "Mis evaluados" en el menú
- AND si accede por URL directa, ve mensaje "No tienes evaluados asignados"

#### Scenario: Evaluatees response format

- WHEN `GET /employees/{me}/evaluatees` retorna datos
- THEN response 200 includes:
  ```json
  {
    "data": [
      {
        "id": "uuid",
        "firstName": "María",
        "lastName": "García",
        "email": "maria@company.com",
        "employeeNumber": "E001",
        "orgNodeId": "uuid",
        "managerId": "uuid",
        "profileId": "uuid",
        "isActive": true
      }
    ]
  }
  ```
- AND only active employees are returned (isActive: true)
- AND employees are sorted by firstName, lastName

### Requirement: Store state management

El `orgHierarchyStore` SHALL mantener los siguientes estados reactivos:

#### Scenario: Loading state

- WHEN any API call is in progress
- THEN `isLoading()` returns true
- AND UI shows skeleton or loading indicator

#### Scenario: Error state

- WHEN any API call fails
- THEN `getError()` returns error message string
- AND UI shows EmptyState with error
- AND tree data is null

#### Scenario: Data state

- WHEN all API calls succeed
- THEN `getRoot()` returns filtered OrgNode
- AND `getChildren(nodeId)` returns children of given node
- AND `getDescendants(nodeId)` returns all descendants
- AND `getNodeById(nodeId)` returns specific node

### Requirement: Manejo de perfil jefe en jerarquía

El frontend SHALL manejar el perfil `jefe` de forma especial en la vista de jerarquía. Actualmente la vista no tiene manejo especial para este perfil.

#### Scenario: Jefe en /evaluacion/9x9/jerarquia

- GIVEN user with profile `jefe`
- WHEN navigates to `/evaluacion/9x9/jerarquia`
- THEN page detects profile is `jefe`
- AND calls `GET /employees/{me}/evaluatees` to get direct reports
- AND renders table of direct reports (not tree)
- AND table has columns: nombre, puesto, perfil
- AND selecting a row navigates to employee detail or shows panel

#### Scenario: Jefe con evaluatees vacíos

- GIVEN user with profile `jefe` but no direct reports
- WHEN navigates to `/evaluacion/9x9/jerarquia`
- THEN page shows EmptyState "No tienes evaluados asignados"
- AND no table or tree is rendered

### Requirement: Eliminación de dependencia de fixtures en producción

El sistema NO SHALL usar fixtures en producción para datos de jerarquía. Los fixtures SOLO se usarán en desarrollo cuando `VITE_USE_API` no esté definido.

#### Scenario: Producción sin fixtures

- WHEN app runs in production (not DEV mode)
- THEN `orgHierarchyStore` always calls API
- AND no fixture files are imported in production bundle
- AND bundle does not include fixture JSON data

#### Scenario: Desarrollo con API flag

- GIVEN `import.meta.env.DEV && import.meta.env.VITE_USE_API`
- WHEN `orgHierarchyStore.load()` se ejecuta
- THEN se llama a la API real (misma ruta que producción)
- AND NO se usan fixtures

## Acceptance Criteria

1. `client.ts` imports `OrgHierarchyPaths` and includes it in `AppPaths`
2. `orgHierarchyStore.load()` calls API in production, fixtures in DEV
3. Tree is filtered by evaluator scope in production
4. Jefe profile renders table of direct reports
5. "Mis evaluados" menu item only shows for profiles with subordinates
6. Loading/error/data states handled correctly
7. `pnpm run check` passes
8. Production bundle does not include fixture JSON
