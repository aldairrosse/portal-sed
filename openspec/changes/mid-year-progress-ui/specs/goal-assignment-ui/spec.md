# goal-assignment-ui — Delta spec (A4: mid-year-progress-ui)

## MODIFIED Requirements

### Requirement: Jerarquía de edición (decisión #8)

El sistema SHALL detectar si el usuario activo es **dueño** de la asignación o **jefe/lector**. En modo lector SHALL ocultar todos los botones de crear/editar/eliminar y SHALL mostrar "Solicitar cambio" por categoría y por meta. En **fase avance**, el jefe SHALL poder editar avances y agregar comentarios en metas de subordinados (sin poder editar pesos ni eliminar metas).

#### Scenario: Modo editor para el dueño

- GIVEN dev persona activa = `colaborador`
- WHEN navega a `/objetivos/asignacion`
- THEN ve su propia asignación con todos los botones (Nueva categoría, Nueva meta, Editar, Eliminar)

#### Scenario: Modo lector para jefe en fase asignación

- GIVEN dev persona activa = `jefe` y `cyclePhase = 'asignacion'`
- WHEN navega a `/objetivos/asignacion`
- THEN `MANAGER_MAP['colaborador'] === 'jefe'`, por lo que el sistema detecta que el `jefe` ve la asignación de `colaborador`
- AND muestra `ReadOnlyBanner` "Estás viendo las metas de María López García. Solo puedes solicitar cambios."
- AND oculta los botones Nueva/Editar/Eliminar
- AND muestra "Solicitar cambio" en cada categoría y meta

#### Scenario: Modo avance para jefe en fase avance

- GIVEN dev persona activa = `jefe` y `cyclePhase = 'avance'`
- WHEN navega a `/objetivos/asignacion`
- THEN ve la asignación de `colaborador` con campos de avance editables
- AND puede editar `progress` en cada meta del subordinado
- AND puede agregar comentarios en cada meta del subordinado
- AND NO puede editar pesos, valores objetivo, nombre ni descripción
- AND NO puede eliminar metas ni categorías
- AND `ReadOnlyBanner` muestra "Puedes editar avances y agregar comentarios."

#### Scenario: Solicitar cambio (mock)

- GIVEN modo lector activo sobre la meta X de la persona Y
- WHEN hace clic en "Solicitar cambio" en la meta X
- THEN se abre `RequestChangeModal` con el contexto de la meta X (read-only) y un textarea de feedback
- WHEN confirma el envío
- THEN se inserta un `ChangeRequest` en el store local
- AND se muestra `alert-success` "Tu solicitud fue registrada"
- AND el modal se cierra (sin email, sin API)

#### Scenario: RH como dueño (decisión #7)

- GIVEN dev persona activa = `rh`
- WHEN navega a `/objetivos/asignacion`
- THEN ve su propia asignación en **modo editor** (RH también tiene metas, decisión #7)

#### Scenario: RH viendo a otro

- GIVEN dev persona activa = `rh`
- WHEN (futuro) abre la asignación de otro perfil
- THEN entra en modo lector (RH administra catálogo de competencias, no metas ajenas en esta fase)

## ADDED Requirements

### Requirement: Restricciones de edición en fase avance

Cuando `cyclePhase === 'avance'`, el sistema SHALL bloquear la edición de `weight`, `targetValue`, `name`, `description`, `unit` de metas y categorías. Solo `progress` y `comments` SHALT ser editables. Los botones de crear y eliminar SHALL ocultarse.

#### Scenario: Campos bloqueados para empleado

- GIVEN `cyclePhase = 'avance'` y empleado es dueño
- WHEN renderiza `GoalRow`
- THEN campos peso, targetValue, nombre, unidad son read-only
- AND botones "Editar" y "Eliminar" no existen
- AND campo avance es editable
- AND ícono de comentario es accesible

#### Scenario: Campos bloqueados para jefe

- GIVEN `cyclePhase = 'avance'` y jefe viendo subordinado
- WHEN renderiza `GoalRow` del subordinado
- THEN jefe puede editar avance y agregar comentarios
- AND NO puede editar pesos ni valores objetivo
- AND NO hay botones de eliminar
