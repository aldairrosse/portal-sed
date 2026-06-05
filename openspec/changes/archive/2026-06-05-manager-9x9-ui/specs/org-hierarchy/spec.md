# Delta for org-hierarchy

> Modifies: `openspec/specs/org-hierarchy/spec.md`

## ADDED Requirements

### Requirement: Recorrido recursivo del árbol jerárquico

El sistema SHALL soportar operaciones de recorrido de árbol: `getChildren(nodeId)` retorna hijos directos, `getDescendants(nodeId)` retorna todos los nodos bajo un ancestro, y `getSubtree(nodeId)` retorna el subárbol completo con anidamiento.

#### Scenario: Obtener descendientes de director-general

- GIVEN árbol corporativo con 4 niveles (DG → 2 directores → 3 jefes → 6 colaboradores)
- WHEN se llama `getDescendants(dgNodeId)`
- THEN retorna 11 nodos (2 directores + 3 jefes + 6 colaboradores)
- AND los nodos incluyen `parentId` y `profileId` para reconstruir jerarquía

#### Scenario: Obtener subárbol anidado

- GIVEN árbol corporativo completo
- WHEN se llama `getSubtree(dgNodeId)`
- THEN retorna estructura anidada con `children[]` recursivo
- AND cada nivel preserva `profileId`, `title` y `employeeCount`

#### Scenario: Nodo hoja sin descendientes

- GIVEN nodo `colaborador` sin subordinados
- WHEN se llama `getDescendants(leafNodeId)`
- THEN retorna array vacío
- AND `getChildren(leafNodeId)` también retorna array vacío

### Requirement: Drill-down jerárquico en UI

El sistema SHALL renderizar un árbol jerárquico expandible desde la vista `/evaluacion/9x9/jerarquia`. La expansión es lazy: los hijos de un nodo se cargan al expandir, no al renderizar el árbol completo.

#### Scenario: Árbol colapsado al iniciar

- GIVEN director-general accede a jerarquía
- WHEN carga la vista
- THEN solo se muestra el nodo raíz (DG) expandido con hijos directos visibles
- AND los niveles inferiores (jefe → colaborador) aparecen colapsados

#### Scenario: Expansión lazy de un director

- GIVEN árbol con Director A colapsado
- WHEN se hace clic en expandir
- THEN se cargan sus jefes (hijos directos) sin cargar colaboradores
- AND cada jefe muestra indicador de subordinados pendientes de expandir

#### Scenario: Nodo hoja sin controles de expansión

- GIVEN nodo colaborador (hoja) visible en el árbol
- WHEN se renderiza
- THEN no muestra ícono de expandir/colapsar
- AND al hacer clic muestra información del empleado (sin navegación a subnivel)

### Requirement: Scope multi-nivel del evaluador para matriz 9×9

El sistema SHALL calcular el scope de evaluatees para la matriz 9×9 según perfil y posición en jerarquía. Un perfil superior SHALL poder ver empleados de niveles inferiores en modo visualización, aunque la edición de scores se limite a reportes directos.

#### Scenario: Director-general ve todos los empleados

- GIVEN director-general en fase `cierre`
- WHEN accede a la matriz 9×9
- THEN el scope incluye `getDescendants(dgNodeId)` — todos los empleados del árbol
- AND puede ver scores de todos pero solo editar scores de directores (reportes directos)

#### Scenario: Jefe ve solo reportes directos (sin cambios de scope)

- GIVEN jefe con 3 colaboradores directos y 0 indirectos
- WHEN se calcula su scope
- THEN `getChildren(jefeNodeId)` retorna los 3 colaboradores
- AND `getDescendants(jefeNodeId)` coincide con `getChildren` (sin niveles extra)

## MODIFIED Requirements

### Requirement: Perfil de evaluación determina reglas

El `profileId` de cada nodo en el árbol determina las reglas de escala y niveles de aceptación que aplican para ese empleado. El catálogo de pilares y competencias es el mismo para todos (decisión #2), pero los criterios de evaluación varían por perfil. Adicionalmente, el perfil `director-general` tiene acceso de solo lectura a toda la jerarquía bajo su mando.

(Previously: Director-general evaluaba solo directores; ahora puede ver toda la jerarquía.)

#### Scenario: Mismo catálogo, diferentes criterios (sin cambios)

- GIVEN perfil `colaborador` y perfil `jefe` en el mismo pilar "Liderazgo"
- WHEN se comparan los niveles de aceptación
- THEN `colaborador` puede tener nivel 2, `jefe` nivel 4
- AND ambos usan las mismas competencias y definiciones de nivel

#### Scenario: Director-general accede a toda la jerarquía (nuevo)

- GIVEN perfil `director-general` activo
- WHEN consulta el árbol organizacional
- THEN ve todos los niveles: directores, jefes y colaboradores
- AND puede hacer drill-down desde su nodo raíz hasta cualquier hoja

### Requirement: Cálculo de evaluador

El sistema SHALL determinar quién evalúa a quién basándose en la posición en el árbol. Para perfiles `jefe`: el evaluador es el nodo padre directo. Para `director` y `director-general`: el scope de visualización de matriz 9×9 incluye todos los descendientes (`getDescendants`), aunque la edición de calificaciones se limita a hijos directos (`getChildren`).

(Previously: Solo se consideraban hijos directos como evaluatees; ahora perfiles superiores ven todos los descendientes en la matriz 9×9.)

#### Scenario: Jefe evalúa a sus colaboradores (sin cambios)

- GIVEN jefe con 3 colaboradores directos
- WHEN se calcula el scope del jefe
- THEN sus evaluatees son los 3 colaboradores directos
- AND NO incluye colaboradores de otros jefes

#### Scenario: Director-general ve toda la jerarquía (modificado)

- GIVEN director-general con 12 empleados en su árbol
- WHEN se calcula el scope de visualización para matriz 9×9
- THEN ve los 12 empleados (todos los descendientes)
- AND solo puede editar scores de sus directores (hijos directos)

#### Scenario: Director evalúa a managers bajo su tramo (modificado)

- GIVEN director con 2 jefes directos y 5 colaboradores indirectos
- WHEN se calcula el scope de visualización
- THEN ve 7 evaluatees en la matriz 9×9
- AND solo puede editar scores de los 2 jefes directos

## REMOVED Requirements

(Ninguno)

### Non-goals actualizados

El non-goal "Cascada de evaluación: un director NO evalúa directamente a los colaboradores de sus jefes; solo evalúa a sus jefes directos" se **mantiene para edición de scores**. Se **extiende para visualización**: director y director-general SÍ pueden ver colaboradores indirectos en la matriz 9×9 en modo lectura.
