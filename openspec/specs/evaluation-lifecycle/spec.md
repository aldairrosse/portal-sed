# evaluation-lifecycle Specification

## Purpose

Define el **ciclo anual SED**: sus tres fases (inicio de año, medio de año, fin de año), las transiciones de estado permitidas, quién puede actuar en cada fase y las restricciones de edición. Esta spec es la fuente de verdad para el behavior de ciclo en todas las pantallas (A3, A4, A5, A6) y el backend futuro (C2).

**Decisiones reflejadas:** #3 (medio año: editar metas y avances, prohibido eliminar metas), #4 (fin de año: autoevaluación + evaluación RH + jefe 9×9 como vías paralelas).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **Cycle** | `id`, `year` (YYYY), `currentPhase` (`asignacion` \| `avance` \| `cierre`) | Un ciclo por año. `currentPhase` determina qué pantallas y mutaciones están habilitadas. |
| **PhaseDefinition** | `phase` (`asignacion` \| `avance` \| `cierre`), `label`, `order` (1–3), `allowedActors` (`string[]`), `allowedActions` (`string[]`), `blockedActions` (`string[]`) | Catálogo estático de fases. No es editable por el usuario. |
| **PhaseTransition** | `fromPhase`, `toPhase`, `trigger` (`auto` \| `manual-rh`), `conditions?` | Define el grafo de transiciones válidas. |

### Estados de meta por fase

| Phase | Meta states permitidos | Acciones habilitadas |
|-------|----------------------|---------------------|
| `asignacion` | `borrador` → `fijada` | Crear, editar, eliminar metas; fijar metas; crear/editar/eliminar categorías; vincular KPIs; definir ponderaciones |
| `avance` | `fijada` → `en-seguimiento` | Editar campos de meta (nombre, descripción, targetValue, KPIs); registrar avances; **NO eliminar metas**; **NO crear metas nuevas** |
| `cierre` | `en-seguimiento` → `evaluada` → `cerrada` | Autoevaluación (empleado); calificación 9×9 (jefe); evaluación formal (RH); cierre de ciclo |

### Estados de evaluación por fase

| Phase | Evaluation states | Quién actúa |
|-------|------------------|-------------|
| `asignacion` | `pendiente-asignacion` | Empleado fija metas; RH asigna competencias |
| `avance` | `pendiente-avance` | Empleado registra avances |
| `cierre` | `pendiente-evaluacion-final` → `completada` | Empleado (autoevaluación), jefe (9×9), RH (evaluación formal) |

## Requirements

### Requirement: Transiciones de fase

El sistema SHALL soportar exactamente tres fases en orden: `asignacion` → `avance` → `cierre`. Las transiciones SHALL ser lineales (sin retroceso) y estar gobernadas por `PhaseTransition`.

#### Scenario: Transición inicio → medio año

- GIVEN ciclo en fase `asignacion`
- WHEN RH activa la transición (o se cumple condición temporal automática)
- THEN `currentPhase` cambia a `avance`
- AND las metas en estado `fijada` pasan a `en-seguimiento`
- AND la UI refleja las nuevas acciones habilitadas

#### Scenario: Transición medio → fin de año

- GIVEN ciclo en fase `avance`
- WHEN se activa la transición
- THEN `currentPhase` cambia a `cierre`
- AND se habilitan las tres vías paralelas de evaluación (autoevaluación, 9×9, evaluación RH)

#### Scenario: Sin retroceso de fase

- GIVEN ciclo en fase `avance`
- WHEN se intenta volver a `asignacion`
- THEN el sistema rechaza la transición
- AND muestra error "No es posible retroceder de fase"

### Requirement: Restricciones de edición en medio de año (decisión #3)

En fase `avance`, el sistema SHALL permitir editar metas existentes y registrar avances, pero SHALL bloquear la eliminación de metas y la creación de metas nuevas.

#### Scenario: Editar meta en medio año

- GIVEN ciclo en fase `avance`, meta en estado `en-seguimiento`
- WHEN empleado edita nombre, descripción, targetValue o KPIs de la meta
- THEN los cambios se persisten
- AND la meta mantiene su estado `en-seguimiento`

#### Scenario: Registrar avance en meta

- GIVEN ciclo en fase `avance`, meta en estado `en-seguimiento`
- WHEN empleado registra un valor de avance (% o monto según `unit`)
- THEN el avance se actualiza
- AND el semáforo/indicador de avance se recalcula

#### Scenario: Bloquear eliminación de meta en medio año

- GIVEN ciclo en fase `avance`
- WHEN empleado intenta eliminar una meta
- THEN la acción está bloqueada (botón deshabilitado o no renderizado)
- AND no existe flujo de confirmación para eliminar meta en esta fase

#### Scenario: Bloquear creación de meta en medio año

- GIVEN ciclo en fase `avance`
- WHEN empleado intenta crear una meta nueva
- THEN la acción está bloqueada (botón "Nueva meta" no disponible)

### Requirement: Acciones permitidas por fase

El sistema SHALL habilitar o deshabilitar acciones CRUD según la fase activa del ciclo.

#### Scenario: Fase inicio — CRUD completo

- GIVEN ciclo en fase `asignacion`
- WHEN empleado accede a su asignación
- THEN tiene acceso a: crear/editar/eliminar categorías, crear/editar/eliminar metas, vincular KPIs, definir ponderaciones, fijar metas

#### Scenario: Fase medio — solo edición parcial

- GIVEN ciclo en fase `avance`
- WHEN empleado accede a su asignación
- THEN tiene acceso a: editar campos de metas existentes, registrar avances
- AND NO tiene acceso a: crear metas, eliminar metas, crear categorías, eliminar categorías, modificar ponderaciones

#### Scenario: Fase fin — solo evaluación

- GIVEN ciclo en fase `cierre`
- WHEN empleado accede a su asignación
- THEN tiene acceso a: autoevaluación (calificar competencias 1–5, comentarios de cierre)
- AND NO tiene acceso a: editar metas, registrar avances, modificar ponderaciones

### Requirement: Vías paralelas en fin de año (decisión #4)

En fase `cierre`, el sistema SHALL soportar tres vías de evaluación en paralelo, cada una independiente:

1. **Autoevaluación del empleado**: califica competencias en escala 1–5 y cierra sus metas.
2. **Evaluación RH**: evaluación formal del empleado (competencias / cierre).
3. **9×9 del jefe**: califica desempeño y potencial para la matriz 9×9 (no sustituye evaluación RH).

#### Scenario: Autoevaluación del empleado

- GIVEN ciclo en fase `cierre`, empleado con metas y competencias asignadas
- WHEN empleado completa su autoevaluación
- THEN registra calificación 1–5 por competencia y comentarios de cierre de metas
- AND su evaluación pasa a estado `completada`

#### Scenario: Jefe califica 9×9

- GIVEN ciclo en fase `cierre`, jefe con evaluados
- WHEN jefe abre la matriz 9×9
- THEN puede calificar desempeño y potencial de cada evaluado
- AND las calificaciones 9×9 son independientes de la evaluación RH

#### Scenario: RH evalúa formalmente

- GIVEN ciclo en fase `cierre`, RH con empleados asignados
- WHEN RH completa la evaluación formal de un empleado
- THEN registra calificación de competencias y cierre
- AND la evaluación formal es la definitiva para el empleado

### Requirement: Calendario y visualización de fase

El sistema SHALL mostrar la fase actual del ciclo de forma prominente en la UI y SHALL indicar qué fases están disponibles, completadas o pendientes.

#### Scenario: Indicador de fase visible

- GIVEN cualquier fase activa
- WHEN empleado navega a cualquier pantalla del ciclo
- THEN se muestra un indicador de fase actual (badge o timeline)
- AND las fases completadas se muestran con check o estilo differente

## Non-goals

- **Persistencia**: esta spec define el behavior del ciclo; la implementación en BD (C2) es un change separado.
- **API de ciclo**: no se expone REST para CRUD de ciclos en esta fase.
- **Configuración de fechas**: las transiciones se definen por触发 manual o condición temporal; no se implementa calendario de fechas específicas.
- **Notificaciones de cambio de fase**: el envío de email al cambiar de fase es scope de C7/C8.
- **Múltiples ciclos activos**: solo un ciclo por año está activo simultáneamente.
- **Retroceso de fase**: explícitamente bloqueado; no se soporta undo de transición.
