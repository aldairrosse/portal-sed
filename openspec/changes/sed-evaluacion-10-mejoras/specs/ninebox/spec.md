# ninebox — Delta spec (sed-evaluacion-10-mejoras)

## ADDED Requirements

### Requirement: Breadcrumb contextual por rol
El sistema SHALL mostrar el breadcrumb según el contexto de la vista: `Mi evaluación > [Rol]` si es self-view, `Evaluaciones > [Rol]` si el visor tiene rol RH, y `Mis evaluados > [Rol]` si el visor es jefe o director. El primer nivel SHALL usar `<a href={backHref}>`, nunca `<button onclick={history.back}>`, para preservar deep-link. `backHref` SHALL ser `/mi-evaluacion` si `employeeId === session.user.employeeId` (self), `/rh/evaluaciones` si el rol es RH, y `/mis-evaluados` si el visor es jefe o director. El label del primer nivel SHALL ser `Mi evaluación` (self), `Evaluaciones` (RH) o `Mis evaluados` (jefe/director). El segundo nivel SHALL ser el `profileName` en titleCase. Los hrefs SHALL ser rutas limpias sin query ni hash (los tabs son radio client-side).

#### Scenario: Breadcrumb self-view
- **WHEN** el empleado abre su propia evaluación
- **THEN** el breadcrumb muestra `Mi evaluación > [Rol]`

#### Scenario: Breadcrumb RH
- **WHEN** un usuario con rol RH abre la evaluación de otro empleado
- **THEN** el breadcrumb muestra `Evaluaciones > [Rol]`

#### Scenario: Breadcrumb jefe o director
- **WHEN** un jefe o director abre la evaluación de un evaluado
- **THEN** el breadcrumb muestra `Mis evaluados > [Rol]`

### Requirement: Columna Evaluación con visibilidad por contexto
El sistema SHALL renombrar la columna `RH` a `Evaluación` en las tablas de competencias. El sistema SHALL ocultar la columna `Evaluación` cuando `employeeId === session.user.employeeId` (self-view). La columna SHALL ser visible para un visor no colaborador que no sea el propio empleado.

#### Scenario: Columna oculta en self-view
- **WHEN** `employeeId === session.user.employeeId`
- **THEN** la columna `Evaluación` no se renderiza

#### Scenario: Columna visible para tercero no colaborador
- **WHEN** el visor no es colaborador del empleado ni es el propio empleado
- **THEN** la columna `Evaluación` es visible con el header `Evaluación`

### Requirement: Grid de columnas uniformes entre pilares
Las tablas de competencias de distintos pilares SHALL usar `sized cols` uniformes para que las columnas queden alineadas entre pilares.

#### Scenario: Alineación entre pilares
- **WHEN** se renderizan tablas de dos pilares distintos
- **THEN** las columnas tienen el mismo ancho definido por `sized cols`

### Requirement: Cálculo de brecha por vista
En self-view la brecha SHALL calcularse como `self vs expectedLevel`. En vista manager o RH la brecha SHALL calcularse como promedio ponderado (`auto` + `rh`) vs `expectedLevel`, usando el peso global provisto por el backend.

#### Scenario: Brecha en self-view
- **WHEN** el empleado ve su propia evaluación
- **THEN** la brecha es `self - expectedLevel`

#### Scenario: Brecha en vista manager o RH
- **WHEN** un manager o RH ve la evaluación de un empleado
- **THEN** la brecha es `promedio ponderado (auto + rh, peso global backend) - expectedLevel`

### Requirement: Radar oculta línea RH en self-view
El gráfico radar SHALL ocultar el dataset de la línea RH cuando la vista es self-view.

#### Scenario: Radar sin línea RH en self-view
- **WHEN** el empleado abre su propio radar
- **THEN** solo se muestran sus datasets propios y el dataset RH queda oculto
