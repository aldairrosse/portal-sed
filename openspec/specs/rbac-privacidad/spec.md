# rbac-privacidad Specification

## Purpose
Expose gerente and coordinador profiles as team-scoped visible roles with privacy-safe evaluation access.

## Requirements

### Requirement: Roles Gerente y Coordinador visibles

El sistema SHALL exponer los perfiles `gerente` y `coordinador` como roles visibles con alcance `team` (descendientes del nodo org del viewer), resolvibles vía `ProfileNameToRole`.

#### Scenario: Gerente ve su equipo

- **WHEN** un viewer con perfil `gerente` lista evaluaciones de su ciclo
- **THEN** recibe solo las de empleados descendientes de su nodo org

#### Scenario: Nombre de perfil desconocido

- **WHEN** `ProfileNameToRole` recibe un nombre no catalogado
- **THEN** retorna rol `colaborador`

### Requirement: Mapping job_title case-insensitive con fallback

El sync SHALL resolver `job_title` → perfil de forma case-insensitive; si no hay match SHALL usar el perfil del jefe (`manager_id`), y si no hay jefe SHALL asignar `colaborador`.

#### Scenario: Título en mayúsculas

- **WHEN** el sync recibe `job_title = "GERENTE"`
- **THEN** asigna perfil `gerente`

#### Scenario: Título sin match con jefe disponible

- **WHEN** el título no matchea y el empleado tiene jefe con perfil `jefe`
- **THEN** asigna el perfil del jefe como fallback

### Requirement: Privacidad colaborador en avance y medio-año

Un viewer `colaborador` SHALL ver solo sus propias evaluaciones en fases `avance` y `medio-anio`; el acceso directo a evaluaciones ajenas SHALL retornar `403`. El evaluador sugerido SHALL ser el jefe, con fallback a RH si no hay jefe.

#### Scenario: Colaborador lista avance

- **WHEN** un colaborador llama `GET /evaluations?phase=avance`
- **THEN** recibe 200 solo con sus propias evaluaciones

#### Scenario: Colaborador abre evaluación ajena de medio-año

- **WHEN** un colaborador llama `GET /evaluations/{idAjeno}?phase=medio-anio`
- **THEN** recibe 403
