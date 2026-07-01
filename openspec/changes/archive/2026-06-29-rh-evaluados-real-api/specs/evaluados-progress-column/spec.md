# Delta for evaluados-progress-column

## ADDED Requirements

### Requirement: Soporte de fuente de datos dual

`EmployeeEvaluationTable` SHALL aceptar `EmployeeListItem[]` como fuente alternativa de filas para la vista RH, manteniendo la ruta `EmployeeAssignment[]` + `goalsStore` para la vista manager "Mis evaluados".

#### Scenario: Vista RH con EmployeeListItem

- GIVEN `EmployeeEvaluationTable` recibe `rows: EmployeeListItem[]`
- WHEN renderiza
- THEN columna "Empleado" muestra `firstName lastName`
- AND columna "Perfil" muestra `profileName`
- AND columna "Progreso global" muestra "—" (sin datos de metas)
- AND columna "Estado" muestra "—" (sin datos de evaluación)

#### Scenario: Vista manager sin cambios

- GIVEN `EmployeeEvaluationTable` recibe `rows: EmployeeAssignment[]`
- WHEN renderiza
- THEN "Progreso global" se calcula con `getGoals()` y `getAssignments()`
- AND "Estado" usa `getEvaluationStatus()`
- AND énfasis visual de filas pendientes funciona igual

#### Scenario: Tabla vacía

- GIVEN `rows` es array vacío
- WHEN renderiza
- THEN muestra empty state "Sin empleados para mostrar"
- AND oculta controles de paginación
