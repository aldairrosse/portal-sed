# shared-goals Specification

## Purpose

Definir las **reglas de negocio para metas compartidas creadas por jefes/directores**: metas que un jefe define para su grupo directo, donde solo el creador puede editar y registrar avance para todos los miembros del grupo. Las metas compartidas soportan cualitativos y cuantitativos con ponderación variable entre empleados del grupo.

**Decisiones reflejadas:** jefe como creador de metas grupales, permisos exclusivos del creador, pesos variables por empleado, acordeones cualitativos/cuantitativos.

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **Goal** (extendido) | `type` (`personal` \| `global` \| `shared`), `goal_kind` (`qualitative` \| `quantitative`) | `type=shared` para metas de jefes. |
| **SharedGoalGroup** | `id`, `goalId`, `createdBy`, `name`, `description` | Agrupa metas compartidas por jefe/grupo. |
| **SharedGoalMember** | `id`, `groupId`, `employeeId`, `weight`, `targetValue`, `baselineValue` | Miembro del grupo con peso variable. |

### Reglas de permisos

```
┌─────────────────────────────────────────────────────────────┐
│  Jefe crea Meta Compartida                                  │
│  Nombre: "Proyecto Q1"                                      │
│  Tipo: cualitativa                                          │
│                                                              │
│  ┌─ Grupo de destino ───────────────────────────────────┐   │
│  │  Empleado A: peso 30%, target 100                     │   │
│  │  Empleado B: peso 40%, target 150                     │   │
│  │  Empleado C: peso 30%, target 120                     │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─ Permisos ───────────────────────────────────────────┐   │
│  │  Jefe creador: editar, registrar avance, eliminar     │   │
│  │  Empleados del grupo: solo lectura                    │   │
│  │  Otros jefes: no acceso                               │   │
│  │  RH: solo lectura (futuro)                            │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Requirements

### Requirement: Creación de metas compartidas por jefe/director

Solo perfiles `jefe`, `gerente-tienda`, `divisional`, `regional`, `director` SHALL poder crear metas compartidas. Las metas compartidas SHALL tener `type=shared` y `goal_kind` (`qualitative` o `quantitative`).

#### Scenario: Jefe crea meta compartida cualitativa

- GIVEN perfil `jefe` autenticado
- WHEN crea meta "Proyecto de innovación" con `goal_kind=qualitative`
- THEN se crea `SharedGoalGroup` con el jefe como `createdBy`
- AND se crea la meta con `type=shared`

#### Scenario: Director crea meta compartida cuantitativa

- GIVEN perfil `director` autenticado
- WHEN crea meta "Incrementar ingresos 20%" con `goal_kind=quantitative`
- THEN se crea `SharedGoalGroup` con el director como `createdBy`
- AND se crea la meta con `type=shared`

#### Scenario: Colaborador no puede crear metas compartidas

- GIVEN perfil `colaborador` autenticado
- WHEN intenta acceder a `/objetivos/compartidas`
- THEN el sistema bloquea el acceso

### Requirement: Grupo de destinatarios

El creador SHALL definir qué empleados de su grupo directo reciben la meta compartida. Cada miembro SHALL poder tener peso y target diferentes.

#### Scenario: Seleccionar empleados del grupo

- GIVEN jefe con 5 evaluados directos
- WHEN crea meta compartida y selecciona 3 evaluados
- THEN se crea `SharedGoalGroup` con 3 `SharedGoalMember`
- AND cada miembro tiene peso y target configurados

#### Scenario: Agregar miembro después

- GIVEN meta compartida con 3 miembros
- WHEN jefe agrega empleado D
- THEN se crea nuevo `SharedGoalMember`
- AND el empleado D ve la meta en su asignación

#### Scenario: Remover miembro

- GIVEN meta compartida con 4 miembros
- WHEN jefe remueve empleado B
- THEN se elimina el `SharedGoalMember` de B
- AND B ya no ve la meta

### Requirement: Pesos variables entre miembros

Cada `SharedGoalMember` SHALL poder tener peso y target diferentes. La suma de pesos de metas compartidas por empleado SHALL ser ≤100%.

#### Scenario: Pesos diferentes por miembro

- GIVEN meta compartida "Proyecto Q1"
- WHEN se asigna a empleado A con peso 30% y a empleado B con peso 40%
- THEN cada miembro respeta su peso
- AND la suma de metas compartidas + personales + globales ≤100%

#### Scenario: Validación de peso máximo

- GIVEN empleado con metas personales (60%) + metas globales (20%)
- WHEN jefe le asigna meta compartida con peso 25%
- THEN el sistema calcula que excede 100%
- AND muestra advertencia pero permite guardar

### Requirement: Permisos exclusivos del creador

Solo el creador de la meta compartida SHALL poder editarla y registrar avance. Los miembros del grupo SHALL ver la meta en modo solo lectura.

#### Scenario: Creador edita meta compartida

- GIVEN jefe que creó "Proyecto Q1"
- WHEN edita nombre, peso o target
- THEN los cambios se reflejan en todos los miembros
- AND se mantiene la integridad

#### Scenario: Creador registra avance

- GIVEN jefe que creó "Proyecto Q1"
- WHEN registra avance para empleado A
- THEN se actualiza `current_value` de la meta para A
- AND el avance es visible para A

#### Scenario: Miembro no puede editar

- GIVEN empleado A asignado a meta compartida "Proyecto Q1"
- WHEN accede a `/objetivos/asignacion`
- THEN ve la meta en modo solo lectura
- AND NO existe botón de editar o eliminar
- AND puede registrar avance si el ciclo lo permite

#### Scenario: Otro jefe no puede acceder

- GIVEN jefe B que no creó "Proyecto Q1"
- WHEN intenta acceder a la meta
- THEN el sistema bloquea el acceso

### Requirement: Acordeones cualitativos/cuantitativos

Las metas compartidas SHALL mostrarse en dos acordeones separados en la vista del creador. En la vista del miembro, se muestran mezcladas con sus metas personales.

#### Scenario: Vista del creador

- GIVEN jefe con 2 metas compartidas cualitativas y 1 cuantitativa
- WHEN accede a `/objetivos/compartidas`
- THEN ve acordeón "Cualitativos" con 2 metas
- AND ve acordeón "Cuantitativos" con 1 meta
- AND cada acordeón muestra su ponderación total

#### Scenario: Vista del miembro

- GIVEN empleado con meta compartida asignada
- WHEN accede a `/objetivos/asignacion`
- THEN ve la meta compartida mezclada con sus metas personales
- AND la meta compartida tiene badge "Compartida"
- AND es de solo lectura

### Requirement: API para metas compartidas

El sistema SHALL exponer endpoints REST para CRUD de metas compartidas y gestión de miembros.

#### Scenario: Crear meta compartida

- POST `/api/v1/goals/shared`
- Body: `{ name, description, unit, direction, targetValue, goalKind, members[] }`
- Response: `201 Created` con meta y grupo creados

#### Scenario: Listar metas compartidas del creador

- GET `/api/v1/goals/shared?createdBy=me`
- Response: `200 OK` con metas donde el usuario es creador

#### Scenario: Listar metas compartidas como miembro

- GET `/api/v1/goals/shared?memberOf=true`
- Response: `200 OK` con metas donde el usuario es miembro

#### Scenario: Actualizar meta compartida

- PUT `/api/v1/goals/shared/:id`
- Body: `{ name, description, unit, targetValue, goalKind }`
- Response: `200 OK` con meta actualizada
- AND solo el creador puede ejecutar

#### Scenario: Agregar miembro

- POST `/api/v1/goals/shared/:id/members`
- Body: `{ employeeId, weight, targetValue }`
- Response: `201 Created` con miembro agregado

#### Scenario: Remover miembro

- DELETE `/api/v1/goals/shared/:id/members/:employeeId`
- Response: `204 No Content`

#### Scenario: Registrar avance

- PUT `/api/v1/goals/shared/:id/progress/:employeeId`
- Body: `{ currentValue }`
- Response: `200 OK` con avance actualizado
- AND solo el creador puede ejecutar

### Requirement: UI para metas compartidas

La ruta `/objetivos/compartidas` SHALL mostrar el editor de metas compartidas con acordeones y gestión de miembros.

#### Scenario: Vista principal del creador

- GIVEN perfil `jefe` accede a `/objetivos/compartidas`
- THEN ve dos acordeones: "Cualitativos" y "Cuantitativos"
- AND ve botón "Nueva meta compartida"
- AND ve lista de metas creadas con indicador de miembros

#### Scenario: Crear meta compartida desde UI

- WHEN jefe hace clic en "Nueva meta compartida"
- THEN se abre formulario con campos: nombre, descripción, unidad, dirección, peso, target, KPIs
- AND al guardar, se abre panel de selección de miembros

#### Scenario: Panel de miembros

- GIVEN meta compartida creada
- WHEN jefe procede a seleccionar miembros
- THEN ve lista de sus evaluados directos con checkboxes
- AND puede configurar pesos variables por miembro
- AND ve preview de impacto en ponderación de cada miembro

#### Scenario: Gestionar miembros existentes

- GIVEN meta compartida con 3 miembros
- WHEN jefe hace clic en "Gestionar miembros"
- THEN ve lista de miembros actuales con pesos
- AND puede agregar/remover miembros
- AND puede editar pesos individualmente

## Non-goals

- **Metas globales**: scope de spec `global-objectives`.
- **Metas personales**: ya existente en `goals-and-weighting`.
- **Evaluación de metas**: calificación final es scope de A5.
- **Aprobación de metas**: las metas compartidas no requieren aprobación.
- **Notificaciones**: no se envían notificaciones al crear metas compartidas.
- **Historial de cambios**: no se registra audit log de ediciones (scope futuro).
- **Edición por miembros**: los miembros no pueden editar metas compartidas.
