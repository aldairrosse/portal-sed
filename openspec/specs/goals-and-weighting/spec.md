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

KPIs son **indicadores** reutilizables (numérico, porcentaje o moneda) que pueden vincularse a 1..N metas. Cada meta puede tener 0..N KPIs asociados.

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

Las reglas de edición de metas dependen de la fase del ciclo activo. En medio de año (`avance`), está **prohibido eliminar** metas.

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
- THEN las metas son de solo lectura
- AND solo puede realizar autoevaluación (calificar competencias 1–5)

### Requirement: Jerarquía de edición (decisión #8)

Cada empleado define sus propias metas, categorías, ponderaciones y KPIs. Jefes/directores/gerentes pueden **VER** las definiciones de personas a cargo y **SOLICITAR CAMBIOS**, pero NO pueden borrar ni agregar metas.

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

## Non-goals

- **Persistencia**: esta spec define las reglas de negocio; la implementación en BD (C4) es un change separado.
- **API de metas**: no se expone REST para CRUD de metas en esta fase.
- **Evaluación de metas**: la calificación final de metas es scope de A5 (fin de año).
- **Metas inter-employee**: no se soporta agregación de metas entre personas o comparación de rendimiento.
- **Plantillas de metas**: no se soporta crear metas desde plantillas predefinidas.
- **Wizard multi-step**: el formulario de metas es una sola pantalla, no un asistente paso a paso.
- **Importación desde Excel**: no se soporta carga masiva de metas.
- **Historial de cambios**: no se registra audit log de ediciones de metas (scope de C4/C7).
