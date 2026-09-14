# goals-and-weighting Specification

## Purpose

Define las **reglas de negocio para metas y KPIs**: unidades de medida, sistema de doble ponderación 100%, vinculación N:M entre metas y KPIs, reglas de edición por fase del ciclo, y restricciones de borrado. Esta spec es la fuente de verdad para el dominio de metas que alimenta las pantallas A3 (asignación anual), A4 (medio año) y el backend C4.

**Decisiones reflejadas:** #1 (doble ponderación 100%), #5 (categorías de metas custom independientes de pilares), #6 (KPIs como indicadores vinculables a 1+ metas), #7 (todos los perfiles evaluables, incluido RH), #8 (jerarquía de edición: ver + solicitar cambios, no agregar/borrar).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **GoalCategory** | `id`, `name`, `description`, `weight` (0–100) | Categoría custom definida por el empleado. Independiente de pilares de competencias (decisión #5). El peso representa el porcentaje de la categoría dentro del total del empleado. |
| **Goal** | `id`, `categoryId`, `name`, `description`, `unit` (`porcentaje` \| `moneda` \| `numero`), `weight` (0–100), `targetValue` (> 0), `state` (ver `evaluation-lifecycle`) | Meta dentro de una categoría. El peso representa el porcentaje de la meta dentro de su categoría. |
| **KPI** | `id`, `name`, `unit` (`porcentaje` \| `moneda` \| `numero`), `description` | Indicador reutilizable. Independiente de cualquier meta o categoría. |
| **GoalKpiLink** | `goalId`, `kpiId` | Join N:M. Una meta puede tener 0..N KPIs; un KPI puede alimentar 1..N metas (decisión #6). |
| **GoalAssignment** | `id`, `employeeId`, `categoryIds[]`, `goalIds[]` | Mapa empleado → sus categorías y metas. Una asignación por empleado por ciclo. |

### Reglas de ponderación (doble 100%)

```
┌─────────────────────────────────────────────────────┐
│  Empleado                                           │
│  Suma de pesos de CATEGORÍAS = 100%                │
│                                                     │
│  ┌─ Categoría A (peso: 40%) ──────────────────┐    │
│  │  Suma de pesos de METAS dentro = 100%       │    │
│  │  Meta 1 (peso: 60%) + Meta 2 (peso: 40%)   │    │
│  └──────────────────────────────────────────────┘    │
│                                                     │
│  ┌─ Categoría B (peso: 35%) ──────────────────┐    │
│  │  Suma de pesos de METAS dentro = 100%       │    │
│  │  Meta 3 (peso: 100%)                        │    │
│  └──────────────────────────────────────────────┘    │
│                                                     │
│  ┌─ Categoría C (peso: 25%) ──────────────────┐    │
│  │  Suma de pesos de METAS dentro = 100%       │    │
│  │  Meta 4 (peso: 50%) + Meta 5 (peso: 50%)   │    │
│  └──────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────┘
```

## Requirements

### Requirement: Unidades de medida para metas

Cada meta SHALL tener una `unit` que define su tipo de medida: `porcentaje` (0–100), `moneda` (monto con símbolo) o `numero` (entero/decimal). La unidad determina cómo se expresa el `targetValue` y el avance.

#### Scenario: Crear meta con unidad porcentaje

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Reducir tasa de rotación" con unit `porcentaje` y targetValue `15`
- THEN la meta se muestra con sufijo "%" y el avance se registra como porcentaje

#### Scenario: Crear meta con unidad moneda

- GIVEN empleado en fase `asignacion`
- WHEN crea meta "Incrementar ingresos" con unit `moneda` y targetValue `500000`
- THEN la meta se muestra con formato monetario y el avance se registra como monto

#### Scenario: Unidad es independiente del KPI

- GIVEN KPI "Ingresos trimestrales" con unit `moneda`
- WHEN se vincula a meta "Crecimiento sostenible" con unit `porcentaje`
- THEN la vinculación es válida
- AND la meta mantiene su unidad propia (`porcentaje`), no hereda la del KPI

### Requirement: Doble ponderación 100% (decisión #1)

El sistema SHALL implementar doble ponderación: las categorías suman 100% del empleado, y las metas dentro de cada categoría suman 100% de esa categoría. Ambas sumas SHALL ser validadas independientemente.

#### Scenario: Suma de categorías = 100%

- GIVEN empleado con 3 categorías de pesos 40, 35, 25
- WHEN valida su asignación
- THEN la suma de categorías es 100% ✓
- AND el sistema permite guardar

#### Scenario: Suma de categorías ≠ 100%

- GIVEN empleado con 2 categorías de pesos 60, 30
- WHEN valida su asignación
- THEN la suma es 90%, no cumple
- AND el sistema bloquea el guardado
- AND muestra feedback indicando el déficit (faltan 10%)

#### Scenario: Suma de metas dentro de categoría = 100%

- GIVEN categoría "Resultados de negocio" con peso 40% y 2 metas de pesos 60 y 40
- WHEN valida la categoría
- THEN la suma de metas es 100% ✓

#### Scenario: Suma de metas dentro de categoría ≠ 100%

- GIVEN categoría con 2 metas de pesos 70 y 20
- WHEN valida la categoría
- THEN la suma es 90%, no cumple
- AND el sistema bloquea el guardado de la categoría
- AND muestra feedback indicando el déficit

#### Scenario: Categoría vacía (sin metas)

- GIVEN categoría recién creada sin metas
- WHEN valida la categoría
- THEN la suma de metas es 0%
- AND se muestra badge de advertencia "Sin metas"
- AND la categoría vacía no bloquea la validación global (decisión transitoria)

#### Scenario: Tolerancia flotante

- GIVEN pesos que suman 99.99 o 100.01
- WHEN valida la suma
- THEN el sistema acepta como válido (tolerancia ε = 0.01)

### Requirement: Vinculación KPIs a metas (decisión #6)

El sistema SHALL permitir vincular KPIs reutilizables (numérico, porcentaje o moneda) a 1..N metas. Cada meta SHALL poder tener 0..N KPIs asociados.

#### Scenario: Vincular KPI existente a meta

- GIVEN KPI "NPS clientes" (unit: `porcentaje`) y meta "Mejorar satisfacción"
- WHEN se vincula el KPI a la meta
- THEN la meta muestra el badge del KPI
- AND el KPI puede ser consultado desde la meta y viceversa

#### Scenario: KPI alimenta múltiples metas

- GIVEN KPI "Ingresos trimestrales" vinculado a 3 metas
- WHEN se consulta el KPI
- THEN muestra las 3 metas asociadas
- AND eliminar una meta elimina solo el vínculo, no el KPI

#### Scenario: Meta sin KPI

- GIVEN meta "Desarrollo personal"
- WHEN se guarda sin vincular ningún KPI
- THEN la meta es válida
- AND no muestra badges de KPI

### Requirement: Restricciones de edición por fase (decisión #3)

El sistema SHALL restringir la edición de metas según la fase del ciclo activo. En medio de año (`avance`), el sistema SHALL prohibir eliminar metas. El registro de avances SHALL estar permitido en `avance` (incl. alias `medio-anio`) y en `cierre`; en `cierre` el resto de campos de meta siguen siendo de solo lectura.

#### Scenario: Inicio de año — CRUD completo

- GIVEN ciclo en fase `asignacion`
- WHEN empleado edita sus metas
- THEN puede: crear, editar (todos los campos), eliminar metas y categorías, modificar ponderaciones, vincular/desvincular KPIs

#### Scenario: Medio de año — solo edición parcial

- GIVEN ciclo en fase `avance`
- WHEN empleado edita sus metas
- THEN puede: editar campos de meta (nombre, descripción, targetValue, KPIs), registrar avances
- AND NO puede: crear metas nuevas, eliminar metas, crear/eliminar categorías, modificar ponderaciones

#### Scenario: Fin de año — sin edición de metas

- GIVEN ciclo en fase `cierre`
- WHEN empleado accede a sus metas
- THEN las metas son de solo lectura, excepto el registro de avances que sigue permitido
- AND solo puede realizar autoevaluación (calificar competencias 1–5)

#### Scenario: Cierre — avances permitidos, resto solo lectura

- GIVEN ciclo en fase `cierre`
- WHEN empleado registra avance en su meta
- THEN el avance persiste
- AND NO puede: crear/editar/eliminar metas, categorías, ponderaciones ni KPIs

#### Scenario: Asignación — avances bloqueados

- GIVEN ciclo en fase `asignacion`
- WHEN empleado intenta registrar avance
- THEN el sistema lo rechaza (fase no permite progreso)

### Requirement: Jerarquía de edición (decisión #8)

El sistema SHALL permitir que cada empleado defina sus propias metas, categorías, ponderaciones y KPIs. El sistema SHALL permitir a jefes/directores/gerentes VER las definiciones de personas a cargo y SOLICITAR CAMBIOS, pero NO SHALL permitir borrar ni agregar metas ajenas.

#### Scenario: Dueño tiene control total

- GIVEN empleado `colaborador` en fase `asignacion`
- WHEN accede a su asignación
- THEN tiene acceso completo de CRUD sobre sus categorías y metas

#### Scenario: Jefe ve definiciones ajenas

- GIVEN jefe con 3 evaluados
- WHEN accede a la asignación de un evaluado
- THEN ve las categorías y metas del evaluado en modo lectura
- AND tiene botón "Solicitar cambio" por categoría y meta
- AND NO tiene botones de crear, editar o eliminar

#### Scenario: Solicitud de cambio (mock)

- GIVEN jefe en modo lectura sobre la meta "Reducir costos" del evaluado X
- WHEN hace clic en "Solicitar cambio"
- THEN se abre modal con la meta en read-only y textarea de feedback
- WHEN confirma
- THEN se registra la solicitud (mock local, sin persistencia, sin email)

#### Scenario: RH como dueño (decisión #7)

- GIVEN perfil `rh` activo en fase `asignacion`
- WHEN accede a su asignación
- THEN tiene control total sobre sus propias metas
- AND NO tiene acceso de edición sobre metas de otros (RH administra catálogo de competencias, no metas ajenas)

### Requirement: Validación de integridad de ponderación

El sistema SHALL calcular y validar la integridad de la doble ponderación antes de permitir el guardado de una asignación completa.

#### Scenario: Guardado bloqueado por categoría incompleta

- GIVEN empleado con categorías que suman 100%, pero una categoría con metas que suman 80%
- WHEN intenta guardar
- THEN el guardado está bloqueado
- AND se muestra indicador de qué categoría falla

#### Scenario: Guardado exitoso

- GIVEN empleado con categorías que suman 100% y todas las categorías con metas que suman 100%
- WHEN intenta guardar
- THEN el guardado se realiza
- AND se muestra confirmación

### Requirement: Separación total por fase avance/cierre

El sistema SHALL tratar `avance` y `cierre` como evaluaciones independientes: progreso, snapshot y comentarios de una fase SHALL NOT contaminar la otra. `phaseKind` deriva de la fase del ciclo (`EmployeeEvaluationDetail`); `GoalClosureCard` SHALL renderizar single card de la fase actual.

#### Scenario: Fases independientes

- WHEN se registra avance o comentario en `cierre`
- THEN `avance` permanece inmutable y viceversa.

### Requirement: Comentarios de metas separados por fase y rol

Los comentarios de metas SHALL separarse por fase (`avance`/`cierre`) y rol (empleado/jefa). Solo la fase actual SHALL ser editable; la fase previa SHALL ser inmutable (solo lectura).

#### Scenario: Fase actual editable, previa inmutable

- GIVEN ciclo en `cierre`
- WHEN empleado o jefa accede a comentarios de `avance`
- THEN los ve en solo lectura y NO puede editarlos; solo los comentarios de `cierre` son editables.

### Requirement: Permisos jefa en comentarios de metas

Jefa SHALL ver comentarios de metas de sus evaluados en `avance` y `cierre`; SHALL escribir comentarios de jefa solo en la fase actual. Empleado SHALL ver/escribir solo sus comentarios en fase actual (código actual = verdad).

#### Scenario: Matriz de permisos metas

- GIVEN rol `jefa` y fase actual `cierre`
- WHEN accede a comentarios
- THEN ve comentarios empleado+jefa de `avance` (inmutables) y ve/escribe comentarios de `cierre`.
- GIVEN rol `empleado` y fase actual `cierre`
- WHEN accede a comentarios
- THEN ve comentarios de `avance` (inmutables) y escribe solo sus comentarios de `cierre`.

### Requirement: Cierre activo hasta nuevo ciclo (sin finished_at)

`UpdatePhase` SHALL NOT setear `finished_at` al entrar a `cierre`. `finished_at` SHALL setearse solo al crear el nuevo ciclo anual en `asignacion`. El ciclo en `cierre` SHALL seguir resoluble por `GetActiveCycleID WHERE finished_at IS NULL`. Botón de cierre manual fuera de scope.

#### Scenario: Cierre sigue activo

- GIVEN ciclo en fase `cierre` con `finished_at IS NULL`
- WHEN se llama `GetActiveCycleID`
- THEN retorna el ciclo (válido para avances, autoeval y 9-box).
- WHEN se crea el nuevo ciclo anual en `asignacion`
- THEN el ciclo anterior se marca `finished_at` y deja de ser activo.

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

## Non-goals

- **Persistencia**: esta spec define las reglas de negocio; la implementación en BD (C4) es un change separado.
- **API de metas**: no se expone REST para CRUD de metas en esta fase.
- **Evaluación de metas**: la calificación final de metas es scope de A5 (fin de año).
- **Metas inter-employee**: no se soporta agregación de metas entre personas o comparación de rendimiento.
- **Plantillas de metas**: no se soporta crear metas desde plantillas predefinidas.
- **Wizard multi-step**: el formulario de metas es una sola pantalla, no un asistente paso a paso.
- **Importación desde Excel**: no se soporta carga masiva de metas.
- **Historial de cambios**: no se registra audit log de ediciones de metas (scope de C4/C7).
