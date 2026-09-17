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

### Requirement: Evaluation con fila propia por employee, cycle y phase

El sistema SHALL buscar y crear Evaluations por la tupla (employee,cycle,phase) mediante `Ensure`/`FindBy(employee,cycle,phase)`. Escribir en `avance` SHALL afectar solo la fila `avance`; escribir en `cierre` SHALL afectar solo la fila `cierre`. El sistema SHALL NOT reutilizar la fila de otra fase.

#### Scenario: Avance escribe solo avance

- **GIVEN** existen filas `avance` y `cierre` para (employee,cycle)
- **WHEN** se escribe avance/ratings en fase `avance`
- **THEN** solo la fila `avance` cambia y la fila `cierre` queda intacta

#### Scenario: Cierre escribe solo cierre

- **GIVEN** existen filas `avance` y `cierre` para (employee,cycle)
- **WHEN** se escribe en fase `cierre`
- **THEN** solo la fila `cierre` cambia y la fila `avance` queda intacta

#### Scenario: Auto-creación de Evaluation al abrir cierre

- **GIVEN** no existe Evaluation para (employee,cycle,`cierre`)
- **WHEN** se abre la vista de cierre
- **THEN** el sistema crea la fila `cierre` automáticamente (idempotente, concurrente OK) y la muestra sin error 404

### Requirement: Avances, ratings y comentarios por rol y fase separados e independientes

Los avances/ratings y los comentarios SHALL almacenarse por (rol,fase) de forma separada e independiente (self, jefe, RH × avance/cierre). Editar el comentario de un rol en una fase SHALL NOT alterar el de otro rol ni el de otra fase. Los permisos de ver/escribir comentarios de metas/competencias de la jefa SHALL regirse por el código del repo (`EmployeeEvaluationDetail`).

#### Scenario: Comentario de jefa en avance no toca cierre

- **WHEN** la jefa escribe un comentario de meta/competencia en fase `avance`
- **THEN** el comentario de fase `cierre` del mismo rol queda intacto

#### Scenario: Permisos de jefa según código

- **GIVEN** el código de `EmployeeEvaluationDetail` define ver/escribir comentarios para la jefa
- **WHEN** hay conflicto entre spec y código
- **THEN** prevalece el código y la spec se corrige

### Requirement: Revert cierre→avance conserva ambas fases y deja editable la fase actual

El sistema SHALL permitir revertir de `cierre` a `avance` conservando snapshot y data de ambas fases (sin borrados). Tras el revert, la fase `avance` SHALL quedar editable (avances/ratings + comentarios de la fase actual editables).

#### Scenario: Revert conserva ambas filas

- **GIVEN** una Evaluation con filas `avance` y `cierre` con data
- **WHEN** RH revierte `cierre`→`avance`
- **THEN** ambas filas conservan su data y la fase activa es `avance` editable

#### Scenario: Editable tras revert

- **GIVEN** se revirtió `cierre`→`avance`
- **WHEN** se escribe avance o comentario en fase `avance`
- **THEN** el sistema acepta la escritura en la fila `avance` sin error de fase

### Requirement: Gates de escritura por fase

La escritura de competencias (self y manager/RH) SHALL estar permitida en fases `avance`/`medio-anio` y `cierre`; Submit/Finalize SHALL seguir exigiendo `cierre`. Fuera de fases escribibles el sistema SHALL responder 409 PHASE_NOT_ADVANCEABLE. La escritura RH (`POST`/`PUT /evaluations/{id}/rh-evaluation`) SHALL estar autorizada solo a RH (permiso `eval:rh`) o al jefe asignado; sin sesión SHALL responder 401, no autorizado SHALL responder 403.

#### Scenario: Self 1-5 en avance OK

- **WHEN** el empleado actualiza su autoevaluación con rating 1-5 en fase `avance`
- **THEN** el sistema acepta la escritura sin error de fase

#### Scenario: Submit en avance responde 409

- **WHEN** se intenta Submit (self o RH) o Finalize en fase `avance`
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE (operación solo `cierre`)

#### Scenario: Fuera de fase escribible responde 409

- **WHEN** se actualiza self o RH fuera de `avance`/`medio-anio`/`cierre` (p. ej. `asignacion`)
- **THEN** el sistema responde 409 PHASE_NOT_ADVANCEABLE

#### Scenario: RH o jefe asignado escribe OK; tercero 403; sin sesión 401

- **WHEN** un usuario con permiso `eval:rh` o el manager del evaluado escribe la evaluación RH
- **THEN** el sistema acepta (sujeto a fase)
- **WHEN** un tercero autenticado no autorizado intenta escribirla
- **THEN** responde 403; sin sesión responde 401

### Requirement: Avance con progreso, autoeval y comentarios visibles

En fase `avance`, el sistema SHALL habilitar registro de progreso y autoevaluación del empleado, y SHALL mostrar los comentarios del jefe y de RH.

#### Scenario: Progreso y autoevaluación habilitados en avance

- **WHEN** el empleado abre su evaluación en fase `avance`
- **THEN** puede registrar progreso en sus metas y completar su autoevaluación

#### Scenario: Comentarios de jefe y RH visibles en avance

- **GIVEN** existen comentarios del jefe o de RH
- **WHEN** el empleado consulta su evaluación en fase `avance`
- **THEN** los comentarios del jefe y de RH son visibles

### Requirement: Ciclo activo único por organización
El sistema SHALL mantener exactamente un ciclo con `is_active = true` por organización (o ninguno antes de la primera activación). Activar un ciclo SHALL poner `is_active = false` en todos los demás ciclos de la misma organización dentro de la misma transacción, bajo lock pesimista de la fila del ciclo a activar. La unicidad SHALL estar respaldada por un unique partial index `(organization_id) WHERE is_active = true`.

#### Scenario: activar un ciclo viejo desactiva el actual
- GIVEN organización con ciclo 2025 activo y ciclo 2024 cerrado
- WHEN RH activa el ciclo 2024
- THEN el ciclo 2024 queda con `is_active = true`
- AND el ciclo 2025 queda con `is_active = false`

#### Scenario: activación concurrente deja un solo ganador
- GIVEN dos peticiones simultáneas de activación sobre ciclos distintos de la misma organización
- WHEN ambas se procesan
- THEN exactamente un ciclo termina con `is_active = true`
- AND la perdedora recibe error o converge al mismo estado sin duplicar activos

#### Scenario: escrituras fuera del ciclo activo son rechazadas
- GIVEN un ciclo cerrado (`is_active = false`)
- WHEN se intenta crear o modificar meta, competencia asignada o evaluación sobre ese ciclo
- THEN el backend rechaza con error 409 y código `cycle-not-active`
- AND no se persiste ningún cambio

## Non-goals

- **Persistencia**: esta spec define el behavior del ciclo; la implementación en BD (C2) es un change separado.
- **API de ciclo**: no se expone REST para CRUD de ciclos en esta fase.
- **Configuración de fechas**: las transiciones se definen por触发 manual o condición temporal; no se implementa calendario de fechas específicas.
- **Notificaciones de cambio de fase**: el envío de email al cambiar de fase es scope de C7/C8.
- **Múltiples ciclos activos**: solo un ciclo por año está activo simultáneamente.
- **Retroceso de fase**: explícitamente bloqueado; no se soporta undo de transición.
