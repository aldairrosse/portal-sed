# org-hierarchy Specification

## Purpose

Define la **jerarquía organizacional dual** (corporativa y retail), el concepto de **evaluador**, el alcance de **"mis evaluados"** y la relación entre jerarquía y perfiles de evaluación. Esta spec es la fuente de verdad para la estructura organizacional que alimenta las pantallas A7 (mis evaluados), A6 (9×9 del jefe) y el backend C5.

**Decisiones reflejadas:** #2 (catálogo único de pilares — la jerarquía define quién evalúa, no el catálogo de competencias), #8 (jerarquía de edición: jefe/director/gerente pueden ver + solicitar cambios en metas de personas a cargo).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **OrganizationTree** | `id`, `type` (`corporativa` \| `retail`), `name` | Árbol organizacional. Coexisten ambos tipos en la misma empresa. |
| **OrgNode** | `id`, `treeId`, `parentId?`, `employeeId`, `profileId`, `title` | Nodo en el árbol. `parentId = null` para la raíz. `profileId` determina el rol de evaluación. |
| **Employee** | `id`, `name`, `email`, `position`, `organizationId?` | Empleado básico. En fase UI-first: datos mock. |
| **EvaluatorScope** | `evaluatorId`, `evaluateeIds[]`, `treeType` (`corporativa` \| `retail`) | Derivado del árbol: quién evalúa a quién. Calculado, no almacenado explícitamente. |

### Jerarquía Corporativa

```
Director General
├── Director A
│   ├── Jefe 1
│   │   ├── Colaborador 1
│   │   ├── Colaborador 2
│   │   └── Colaborador 3
│   └── Jefe 2
│       ├── Colaborador 4
│       └── Colaborador 5
└── Director B
    └── Jefe 3
        └── Colaborador 6
```

**Niveles:** colaborador → jefe → director → director general.

### Jerarquía Retail

```
Regional
├── Divisional 1
│   ├── Gerente Tienda A
│   │   ├── Vendedor 1
│   │   └── Vendedor 2
│   └── Gerente Tienda B
│       ├── Vendedor 3
│       └── Vendedor 4
└── Divisional 2
    └── Gerente Tienda C
        └── Vendedor 5
```

**Niveles:** vendedor → gerente de tienda → divisional → regional.

### Perfiles de evaluación

| Perfil | Árbol | Nivel típico | Puede evaluar |
|--------|-------|-------------|---------------|
| `colaborador` | Corporativa | Hoja | — |
| `jefe` | Corporativa | Intermedio | Colaboradores directos |
| `vendedor` | Retail | Hoja | — |
| `gerente-tienda` | Retail | Intermedio | Vendedores de su tienda |
| `divisional` | Retail | Intermedio | Gerentes de tienda de su división |
| `regional` | Retail | Alto | Divisionales de su región |
| `director` | Corporativa | Alto | Jefes de su área |
| `director-general` | Corporativa | Raíz | Directores |
| `rh` | Ambos (acceso transversal) | Transversal | Todos (admin, no evaluación directa) |

## Requirements

### Requirement: Árbol dual corporativo y retail

El sistema SHALL soportar dos árboles organizacionales simultáneos: corporativo y retail. Cada empleado SHALL pertenecer a exactamente un árbol (o ser `rh` con acceso transversal).

#### Scenario: Empleado en árbol corporativo

- GIVEN árbol corporativo con estructura completa
- WHEN se consulta un empleado con perfil `colaborador`
- THEN se muestra en el árbol corporativo bajo su jefe directo
- AND su cadena de mando es: colaborador → jefe → director → director general

#### Scenario: Empleado en árbol retail

- GIVEN árbol retail con estructura completa
- WHEN se consulta un empleado con perfil `vendedor`
- THEN se muestra en el árbol retail bajo su gerente de tienda
- AND su cadena de mando es: vendedor → gerente → divisional → regional

#### Scenario: RH con acceso transversal

- GIVEN perfil `rh` activo
- WHEN consulta la estructura organizacional
- THEN puede ver ambos árboles (corporativo y retail)
- AND no pertenece a ningún árbol como nodo evaluado

### Requirement: Cálculo de evaluador

El sistema SHALL determinar quién evalúa a quién basándose en la posición en el árbol. Para perfiles `jefe`: el evaluador es el nodo padre directo. Para `director` y `director-general`: el scope de visualización de matriz 9×9 incluye todos los descendientes (`getDescendants`), aunque la edición de calificaciones se limita a hijos directos (`getChildren`).

(Previously: Solo se consideraban hijos directos como evaluatees; ahora perfiles superiores ven todos los descendientes en la matriz 9×9.)

#### Scenario: Jefe evalúa a sus colaboradores (sin cambios)

- GIVEN jefe con 3 colaboradores directos en árbol corporativo
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

### Requirement: Alcance de "mis evaluados"

Cada persona con subordinados directos SHALL ver una lista de "mis evaluados" con acceso a la información de fijación y seguimiento de metas de sus subordinados en inicio y medio año.

#### Scenario: Jefe ve mis evaluados

- GIVEN jefe con 3 colaboradores
- WHEN navega a "Mis evaluados"
- THEN ve lista con nombre, puesto y estado de evaluación de cada colaborador
- AND puede acceder a la asignación de metas de cada uno (modo lectura + solicitud de cambio)

#### Scenario: Gerente de tienda ve mis evaluados

- GIVEN gerente de tienda con 4 vendedores
- WHEN navega a "Mis evaluados"
- THEN ve lista de sus 4 vendedores
- AND puede acceder a sus metas en fase inicio y medio año

#### Scenario: Colaborador sin mis evaluados

- GIVEN perfil `colaborador` sin subordinados
- WHEN navega a la aplicación
- THEN NO ve el ítem "Mis evaluados" en el menú
- AND si accede por URL directa, ve mensaje "No tienes evaluados asignados"

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

### Requirement: Conexión jerarquía ↔ metas (decisión #8)

La jerarquía define quién puede **ver** y **solicitar cambios** en las metas de personas a cargo. El jefe/director/gerente SHALL poder ver la asignación de metas de sus evaluatees y SHALL poder enviar solicitudes de cambio, pero NO SHALL poder crear, editar o eliminar metas ajenas.

#### Scenario: Jefe solicita cambio en meta ajena

- GIVEN jefe con evaluado "María" en modo lectura
- WHEN hace clic en "Solicitar cambio" en la meta "Reducir costos" de María
- THEN se abre modal con la meta en read-only y textarea de feedback
- WHEN confirma
- THEN la solicitud se registra (mock local)
- AND María recibe notificación (futuro: C7)

#### Scenario: Director ve metas de jefe a cargo

- GIVEN director con jefe "Carlos" como evaluatee
- WHEN accede a la asignación de Carlos
- THEN ve las metas de Carlos en modo lectura
- AND puede solicitar cambios igual que un jefe directo

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

## Non-goals

- **Autenticación y RBAC**: esta spec define la estructura; la implementación de auth y permisos es scope de C7.
- **API de jerarquía real**: en fase UI-first, la jerarquía es mock. La API real es C5.
- **Editor de árbol**: no se soporta crear/modificar la estructura organizacional desde la UI (es dato maestro de RH).
- **Multi-tenant**: la spec asume una única empresa; partición por organización es scope futuro.
- **Cascada de evaluación**: un director NO evalúa directamente a los colaboradores de sus jefes; solo evalúa a sus jefes directos. Esta regla se mantiene para edición de scores, pero se extiende para visualización: director y director-general SÍ pueden ver colaboradores indirectos en la matriz 9×9 en modo lectura.
- **Transferencias**: no se soporta mover empleados entre árboles o niveles durante un ciclo.
- **Historial de cambios**: no se registra audit log de cambios en la jerarquía (scope de C5/C7).
