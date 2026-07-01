# rh-evaluados-employee-list Specification

## Purpose

Store reactivo que consume `GET /api/v1/employees` con paginación por cursor y búsqueda server-side para la tabla RH de evaluaciones.

## Requirements

### Requirement: Store reactiva de empleados

El store SHALL exponer `items: EmployeeListItem[]`, `loading: boolean`, `error: string | null` como `$state` reactivo. SHALL llamar `GET /api/v1/employees` al inicializar.

#### Scenario: Carga exitosa

- GIVEN el store se inicializa
- WHEN `GET /api/v1/employees` retorna 200 con `{ data: [...], nextCursor, hasMore }`
- THEN `items` contiene los empleados retornados
- AND `loading` es false
- AND `error` es null

#### Scenario: Error de red

- GIVEN la API no responde
- WHEN el store intenta cargar
- THEN `error` contiene mensaje legible
- AND `loading` es false
- AND `items` está vacío

### Requirement: Paginación por cursor

El store SHALL usar cursores `nextCursor` y `prevCursor` del response. SHALL exponer `next()` / `prev()` y booleans `hasMore` / `hasPrev`. Paginación SHALL reemplazar página completa, no concatenar.

#### Scenario: Avanzar página

- GIVEN response incluye `nextCursor: "abc"`, `hasMore: true`
- WHEN se llama `next()`
- THEN fetch incluye `?cursor=abc`
- AND `items` se reemplaza con la nueva página

#### Scenario: Última página

- GIVEN response incluye `hasMore: false`
- THEN `next()` es no-op
- AND control "Siguiente" aparece deshabilitado

### Requirement: Búsqueda server-side con debounce

El store SHALL exponer `search(q: string)` con debounce de 300ms. Al buscar, SHALL resetear cursor a primera página e incluir `q` como query param.

#### Scenario: Búsqueda tras debounce

- GIVEN usuario escribe "María"
- WHEN transcurren 300ms sin nuevas teclas
- THEN se llama `GET /api/v1/employees?q=María`
- AND cursor vuelve a primera página

#### Scenario: Teclas rápidas no disparan múltiples llamadas

- GIVEN usuario escribe "M", "a", "r" con <100ms entre cada tecla
- WHEN transcurren 300ms desde la última tecla
- THEN se realiza UNA sola llamada con `q=Mar`

### Requirement: Etiqueta de perfil desde API

El store SHALL incluir `profileName` del response en cada `EmployeeListItem`. La tabla SHALL mostrar este valor directamente sin lookup local.

#### Scenario: Render de columna Perfil

- GIVEN API retorna `{ profileName: "Gerente Senior" }`
- WHEN la tabla renderiza la fila
- THEN columna "Perfil" muestra "Gerente Senior"

### Requirement: Estados de UI

El consumidor SHALL manejar tres estados: loading (skeleton), error (banner con reintento), empty (mensaje "Sin empleados").

#### Scenario: Skeleton durante carga

- GIVEN `loading` es true
- WHEN se renderiza la tabla
- THEN se muestran 5 filas skeleton
- AND paginación y búsqueda están deshabilitadas

#### Scenario: Empty state

- GIVEN búsqueda retorna 0 resultados
- WHEN `items` está vacío y `loading` es false
- THEN se muestra "Sin empleados para mostrar"
