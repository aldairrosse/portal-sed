# global-objectives Specification

## Purpose

Definir las **reglas de negocio para objetivos globales creados por RH**: metas que RH define y asigna masivamente a empleados, departamentos o grupos con cierto número de reportes directos. Las metas globales son de solo lectura para los empleados asignados. Soporta metas cualitativas y cuantitativas con ponderación variable por destinatario. Incluye formularios reutilizables (plantillas).

**Decisiones reflejadas:** RH como administrador de objetivos organizacionales, asignación masiva, pesos variables por grupo/regla/empleado, formularios reutilizables.

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **Goal** (extendido) | `type` (`personal` \| `global` \| `shared`), `goal_kind` (`qualitative` \| `quantitative`) | Nuevos campos en entidad existente. `type=global` para metas de RH. |
| **GoalTemplate** | `id`, `name`, `description`, `unit`, `direction`, `targetValue`, `goalKind`, `createdBy`, `isPublic` | Plantilla reutilizable. `isPublic=true` visible para todo RH. |
| **GoalTemplateKpiLink** | `templateId`, `kpiId` | KPIs vinculados a la plantilla. |
| **GlobalGoalAssignment** | `id`, `goalId`, `employeeId`, `weight`, `targetValue`, `baselineValue` | Asignación de meta global a empleado específico. Peso variable. |
| **GlobalGoalRule** | `id`, `goalId`, `ruleType` (`department` \| `min_direct_reports`), `departmentId`, `minDirectReports`, `defaultWeight` | Regla de asignación masiva. |
| **GoalCategory** (extendido) | `type` (`personal` \| `global` \| `shared`) | Nueva categoría: "Cualitativos" y "Cuantitativos" para metas globales. |

### Reglas de asignación global

```
┌─────────────────────────────────────────────────────────────┐
│  RH crea Meta Global                                        │
│  Tipo: cualitativa o cuantitativa                           │
│  Peso base: 20%                                             │
│                                                              │
│  ┌─ Asignación directa ─────────────────────────────────┐   │
│  │  Empleado A: peso 25%, target 500                     │   │
│  │  Empleado B: peso 15%, target 300                     │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─ Regla por departamento ─────────────────────────────┐   │
│  │  Departamento "Ventas": peso 20%, target 400         │   │
│  │  → Aplica a todos los empleados del departamento     │   │
│  └───────────────────────────────────────────────────────┘   │
│                                                              │
│  ┌─ Regla por reportes directos ────────────────────────┐   │
│  │  Empleados con ≥3 reportes directos: peso 30%        │   │
│  │  → Aplica a jefes/directores con ese mínimo          │   │
│  └───────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

## Requirements

### Requirement: Creación de metas globales por RH

Solo el perfil `rh` SHALL poder crear metas globales. Las metas globales SHALL tener `type=global` y `goal_kind` (`qualitative` o `quantitative`).

#### Scenario: RH crea meta global cualitativa

- GIVEN perfil `rh` autenticado
- WHEN crea meta "Mejorar clima laboral" con `goal_kind=qualitative`, `unit=porcentaje`, `targetValue=80`
- THEN la meta se guarda con `type=global`
- AND se muestra en la sección "Objetivos globales"

#### Scenario: No-RH no puede crear metas globales

- GIVEN perfil `colaborador` autenticado
- WHEN intenta acceder a `/objetivos/globales`
- THEN el sistema bloquea el acceso
- AND muestra mensaje "No tienes permisos para acceder a esta sección"

### Requirement: Asignación masiva de metas globales

RH SHALL poder asignar metas globales a: empleados específicos, departamento completo, o empleados con cierto número de reportes directos.

#### Scenario: Asignar a empleados específicos

- GIVEN meta global "Reducir rotación" creada
- WHEN RH selecciona empleados A, B, C
- THEN se crean `GlobalGoalAssignment` para cada empleado
- AND cada asignación puede tener peso y target diferentes

#### Scenario: Asignar por departamento

- GIVEN meta global "Incrementar ventas" creada
- WHEN RH selecciona departamento "Ventas" con regla
- THEN se crea `GlobalGoalRule` con `ruleType=department`
- AND al ejecutar la regla, se crean asignaciones para todos los empleados del departamento
- AND el peso por defecto se aplica a cada empleado

#### Scenario: Asignar por mínimo de reportes directos

- GIVEN meta global "Desarrollar liderazgo" creada
- WHEN RH configura regla `minDirectReports=3`, `defaultWeight=30`
- THEN se crea `GlobalGoalRule` con `ruleType=min_direct_reports`
- AND al ejecutar, se crean asignaciones para empleados con ≥3 reportes directos

### Requirement: Pesos variables por destinatario

Cada `GlobalGoalAssignment` SHALL poder tener peso y target diferentes. La suma de pesos de metas globales por empleado SHALL ser ≤100% (las metas personales complementan).

#### Scenario: Pesos diferentes por empleado

- GIVEN meta global "Proyectos estratégicos" con peso base 20%
- WHEN se asigna a empleado A con peso 25% y a empleado B con peso 15%
- THEN cada asignación respeta su peso
- AND la suma total de metas globales + personales del empleado ≤100%

#### Scenario: Validación de peso máximo

- GIVEN empleado con metas personales que suman 80%
- WHEN se le asigna meta global con peso 25%
- THEN el sistema calcula que excede 100%
- AND muestra advertencia pero permite guardar (las metas globales son obligatorias)

### Requirement: Metas globales de solo lectura

Los empleados asignados SHALL ver las metas globales pero NO podrán editarlas. Solo RH puede modificarlas.

#### Scenario: Empleado ve metas globales

- GIVEN empleado con meta global asignada
- WHEN accede a `/objetivos/asignacion`
- THEN ve las metas globales en acordeón separado
- AND los campos son de solo lectura
- AND NO existe botón de editar o eliminar

#### Scenario: RH edita meta global

- GIVEN meta global "Reducir costos" creada
- WHEN RH edita nombre, peso o target
- THEN los cambios se reflejan en todas las asignaciones
- AND se mantiene la integridad de pesos

### Requirement: Formularios reutilizables (plantillas)

RH SHALL poder crear plantillas de metas que reutilicen nombre, descripción, unidad, dirección y KPIs vinculados.

#### Scenario: Crear plantilla desde meta existente

- GIVEN meta global "Mejorar satisfacción" con 2 KPIs vinculados
- WHEN RH selecciona "Guardar como plantilla"
- THEN se crea `GoalTemplate` con los campos de la meta
- AND los KPIs se vinculan via `GoalTemplateKpiLink`

#### Scenario: Crear meta desde plantilla

- GIVEN plantilla "Meta de ventas" con campos predefinidos
- WHEN RH selecciona "Usar plantilla"
- THEN se abre formulario con campos prellenados
- AND RH puede modificar antes de guardar

#### Scenario: Plantilla pública vs privada

- GIVEN plantilla creada por RH 1
- WHEN se marca como `isPublic=true`
- THEN todos los usuarios RH pueden usarla
- AND si es `isPublic=false`, solo su creador puede usarla

### Requirement: Acordeones cualitativos/cuantitativos

Las metas globales SHALL mostrarse en dos acordeones separados: "Cualitativos" y "Cuantitativos". Cada acordeón muestra las metas del tipo correspondiente.

#### Scenario: Separación por tipo

- GIVEN 3 metas globales: 2 cualitativas, 1 cuantitativa
- WHEN RH accede a `/objetivos/globales`
- THEN ve acordeón "Cualitativos" con 2 metas
- AND ve acordeón "Cuantitativos" con 1 meta
- AND cada acordeón muestra su ponderación total

#### Scenario: Validación por acordeón

- GIVEN acordeón cualitativo con metas que suman 60%
- AND acordeón cuantitativo con metas que suman 40%
- WHEN RH intenta guardar
- THEN la suma total es 100%
- AND el guardado se realiza

### Requirement: API para objetivos globales

El sistema SHALL exponer endpoints REST para CRUD de metas globales y asignaciones.

#### Scenario: Crear meta global

- POST `/api/v1/goals/global`
- Body: `{ name, description, unit, direction, targetValue, goalKind, assignments[], rules[] }`
- Response: `201 Created` con meta creada

#### Scenario: Listar metas globales

- GET `/api/v1/goals/global?cycleId=...`
- Response: `200 OK` con lista de metas globales y sus asignaciones

#### Scenario: Actualizar meta global

- PUT `/api/v1/goals/global/:id`
- Body: `{ name, description, unit, targetValue, goalKind }`
- Response: `200 OK` con meta actualizada

#### Scenario: Eliminar meta global

- DELETE `/api/v1/goals/global/:id`
- Response: `204 No Content`
- AND elimina en cascada todas las asignaciones

#### Scenario: Ejecutar regla de asignación

- POST `/api/v1/goals/global/:id/execute-rules`
- Response: `200 OK` con número de asignaciones creadas

### Requirement: UI para objetivos globales

La ruta `/objetivos/globales` SHALL mostrar el editor de metas globales con acordeones, formularios reutilizables y herramientas de asignación masiva.

#### Scenario: Vista principal

- GIVEN perfil `rh` accede a `/objetivos/globales`
- THEN ve dos acordeones: "Cualitativos" y "Cuantitativos"
- AND ve botón "Nueva meta global"
- AND ve sección de "Plantillas" con opción de crear/usar

#### Scenario: Crear meta global desde UI

- WHEN RH hace clic en "Nueva meta global"
- THEN se abre formulario con campos: nombre, descripción, unidad, dirección, peso, target, KPIs
- AND puede seleccionar "Usar plantilla" para prellenar
- AND al guardar, se abre panel de asignación

#### Scenario: Panel de asignación

- GIVEN meta global creada
- WHEN RH procede a asignar
- THEN ve opciones: "Empleados específicos", "Por departamento", "Por reportes directos"
- AND puede configurar pesos variables por destinatario
- AND ve preview de empleados afectados

## Non-goals

- **Metas compartidas**: scope de spec `shared-goals`.
- **Metas personales**: ya existente en `goals-and-weighting`.
- **Evaluación de metas**: calificación final es scope de A5.
- **Aprobación de metas**: las metas globales no requieren aprobación.
- **Notificaciones**: no se envían notificaciones al asignar metas globales.
- **Historial de cambios**: no se registra audit log de ediciones (scope futuro).
