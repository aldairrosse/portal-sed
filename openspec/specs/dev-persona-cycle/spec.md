# dev-persona-cycle Specification

## Purpose
TBD - created by archiving change ui-shell-and-design-tokens. Update Purpose after archive.
## Requirements
### Requirement: Selector de persona en desarrollo

El sistema SHALL mostrar un selector de perfil de evaluación visible únicamente cuando `import.meta.env.DEV` es verdadero.

#### Scenario: Modo desarrollo

- **WHEN** la aplicación corre en modo desarrollo
- **THEN** el selector lista exactamente 8 perfiles: colaborador, jefe, vendedor, gerente-tienda, divisional, regional, director, rh

#### Scenario: Modo producción

- **WHEN** la aplicación corre en build de producción
- **THEN** el selector de persona no se renderiza

### Requirement: Selector de fase de ciclo en desarrollo

El sistema SHALL permitir elegir la fase del ciclo anual en modo desarrollo entre: inicio de año, medio año y fin de año.

#### Scenario: Cambio de fase

- **WHEN** el desarrollador selecciona una fase distinta en el selector
- **THEN** el store global de fase se actualiza y la UI refleja la fase activa (badge o etiqueta en barra)

### Requirement: Persistencia de contexto dev en sesión

En modo desarrollo, el sistema SHALL persistir perfil y fase seleccionados en `sessionStorage` para sobrevivir recargas de página.

#### Scenario: Recarga de página en dev

- **WHEN** el desarrollador recarga el navegador tras elegir perfil y fase
- **THEN** el perfil y la fase activos se restauran desde sessionStorage

### Requirement: Contexto consumible por rutas hijas

El perfil y la fase activos SHALL estar disponibles para cualquier componente o ruta hijo mediante store compartido sin prop drilling.

#### Scenario: Ruta hija lee contexto

- **WHEN** una página placeholder accede al store de contexto dev
- **THEN** obtiene el perfil y la fase actuales de forma reactiva

### Requirement: Usuario simulado en barra

La barra superior SHALL mostrar nombre y perfil del usuario simulado derivado del perfil activo (datos fixture estáticos por perfil).

#### Scenario: Cambio de perfil

- **WHEN** el desarrollador cambia el perfil en el selector
- **THEN** la barra superior actualiza nombre simulado y etiqueta de perfil en español

