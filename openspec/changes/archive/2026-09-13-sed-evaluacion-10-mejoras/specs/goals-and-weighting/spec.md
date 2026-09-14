# goals-and-weighting — Delta spec (sed-evaluacion-10-mejoras)

## ADDED Requirements

### Requirement: Empty state de mi-evaluación con acceso a metas

Cuando las categorías no tienen metas o las metas no han sido enviadas, la vista mi-evaluación SHALL mostrar un empty state con gate `assignmentStatus` y un botón `Ir a metas` que navega a la gestión de metas (sin redirect auto).

#### Scenario: Categorías sin metas

- **WHEN** las categorías no contienen metas
- **THEN** se muestra el empty state con el botón `Ir a metas`

#### Scenario: Metas no enviadas

- **WHEN** existen metas pero ninguna ha sido enviada
- **THEN** se muestra el empty state con el botón `Ir a metas`

### Requirement: Labels diferenciados de avance y cierre

El sistema SHALL usar el label `Evaluación de avance de medio año` con botón `Guardar avance` en fase `avance`, y el label `Evaluación de cierre de año` con botón `Guardar cierre` en fase `cierre`.

#### Scenario: Labels de avance

- **WHEN** la evaluación está en fase `avance`
- **THEN** el título es `Evaluación de avance de medio año` y el botón es `Guardar avance`

#### Scenario: Labels de cierre

- **WHEN** la evaluación está en fase `cierre`
- **THEN** el título es `Evaluación de cierre de año` y el botón es `Guardar cierre`

### Requirement: Comentarios por rol y fase en metas (verdad del código)

Los comentarios de metas SHALL mostrarse y editarse por (rol,fase) de forma independiente, respetando los permisos de ver/escribir de `EmployeeEvaluationDetail` (incluida la jefa). El comentario enviado en la fase actual SHALL persistir visible en esa fase sin alterar otras fases.

#### Scenario: Comentario con botón enviar persiste en fase actual

- **WHEN** se envía un comentario de meta en la fase actual
- **THEN** persiste visible en esa fase y las demás fases quedan intactas

### Requirement: Avatar con iniciales correctas

El avatar SHALL mostrar las iniciales del nombre correcto del empleado y SHALL no renderizarse en blanco (sin migración S3).

#### Scenario: Iniciales visibles

- **WHEN** se renderiza el avatar del empleado
- **THEN** muestra las iniciales derivadas de su nombre correcto

### Requirement: Textarea de comentarios con una línea por defecto

El textarea de comentarios SHALL renderizarse con una línea por defecto y SHALL exponer el control de resize visible.

#### Scenario: Altura inicial y resize

- **WHEN** se abre el campo de comentarios
- **THEN** ocupa una línea por defecto y el resize es visible y usable

### Requirement: Toasts claros y ErrorState con dueño único

Ante un fallo, el sistema SHALL mostrar un toast con el `error.code` claro. Ante un guardado exitoso, el sistema SHALL mostrar un toast de confirmación. El componente `ErrorState` SHALL tener dueño único en este change (no duplicarlo en otros changes).

#### Scenario: Toast de error con código

- **WHEN** el guardado falla con un código de error
- **THEN** el toast muestra el `error.code` de forma clara

#### Scenario: Toast de éxito

- **WHEN** el guardado es exitoso
- **THEN** el toast confirma la operación
