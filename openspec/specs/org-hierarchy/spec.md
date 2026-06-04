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

El sistema SHALL determinar quién evalúa a quién basándose en la posición en el árbol. El evaluador es el **nodo padre directo** en la jerarquía.

#### Scenario: Jefe evalúa a sus colaboradores

- GIVEN jefe con 3 colaboradores directos en árbol corporativo
- WHEN se calcula el scope del jefe
- THEN sus evaluatees son los 3 colaboradores directos
- AND NO incluye colaboradores de otros jefes

#### Scenario: Gerente evalúa a sus vendedores

- GIVEN gerente de tienda con 4 vendedores en árbol retail
- WHEN se calcula el scope del gerente
- THEN sus evaluatees son los 4 vendedores de su tienda
- AND NO incluye vendedores de otras tiendas

#### Scenario: Director evalúa a sus jefes

- GIVEN director con 2 jefes directos
- WHEN se calcula el scope del director
- THEN sus evaluatees son los 2 jefes
- AND NO incluye colaboradores de esos jefes (evaluación en cascada, no directa)

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

El `profileId` de cada nodo en el árbol determina las **reglas de escala y niveles de aceptación** que aplican para ese empleado. El catálogo de pilares y competencias es el mismo para todos (decisión #2), pero los criterios de evaluación varían por perfil.

#### Scenario: Mismo catálogo, diferentes criterios

- GIVEN perfil `colaborador` y perfil `jefe` en el mismo pilar "Liderazgo"
- WHEN se comparan los niveles de aceptación de "Liderazgo de equipo"
- THEN `colaborador` puede tener nivel 2, `jefe` nivel 4
- AND ambos usan las mismas competencias y definiciones de nivel

#### Scenario: Perfil determina menú y pantallas

- GIVEN perfil `rh` activo
- WHEN observa el menú
- THEN ve ítems de administración de competencias que no ven otros perfiles
- AND NO ve "Mis evaluados" (RH administra, no evalúa directamente en esta spec)

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

## Non-goals

- **Autenticación y RBAC**: esta spec define la estructura; la implementación de auth y permisos es scope de C7.
- **API de jerarquía real**: en fase UI-first, la jerarquía es mock. La API real es C5.
- **Editor de árbol**: no se soporta crear/modificar la estructura organizacional desde la UI (es dato maestro de RH).
- **Multi-tenant**: la spec asume una única empresa; partición por organización es scope futuro.
- **Cascada de evaluación**: un director NO evalúa directamente a los colaboradores de sus jefes; solo evalúa a sus jefes directos.
- **Transferencias**: no se soporta mover empleados entre árboles o niveles durante un ciclo.
- **Historial de cambios**: no se registra audit log de cambios en la jerarquía (scope de C5/C7).
