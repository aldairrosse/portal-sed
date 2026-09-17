# cycle-activation Specification

## Purpose
Permite activar un ciclo de evaluación por organización y operar siempre sobre el ciclo activo, desde el endpoint de activación hasta los badges y guards del frontend.

## Requirements

### Requirement: Endpoint de activación de ciclo
El sistema SHALL exponer `POST /cycles/{id}/activate` (solo rol RH) que marca el ciclo indicado como activo y desactiva los demás de la organización. La respuesta SHALL incluir el ciclo activado con su flag `is_active = true`.

#### Scenario: RH activa un ciclo
- WHEN RH llama `POST /cycles/{id}/activate` con un ciclo existente de su organización
- THEN la respuesta es 200 con el ciclo e `is_active = true`
- AND `GET /cycles/current` devuelve ese ciclo

#### Scenario: sin permiso o ciclo ajeno
- WHEN un usuario sin rol RH (o de otra organización) llama al endpoint
- THEN la respuesta es 403 (o 404 si el ciclo no pertenece a su organización)
- AND ningún flag cambia

### Requirement: Store de ciclo activo y badges en gestión
El frontend SHALL exponer `cycleStore.loadCurrent()` que carga `GET /cycles/current` y lo reutilizan los stores de metas, competencias y evaluación en lugar de recibir el ciclo por parámetro. La vista de gestión de ciclos SHALL mostrar badge "Activo" en el ciclo activo y "Cerrado" en los demás, con acción de activación solo para RH.

#### Scenario: gestión muestra el activo
- WHEN RH abre gestión de ciclos con el ciclo 2025 activo
- THEN el ciclo 2025 muestra badge "Activo" y los demás "Cerrado"

#### Scenario: stores reutilizan el ciclo activo
- WHEN el store de metas necesita el ciclo para crear una meta
- THEN usa el ciclo cargado por `cycleStore.loadCurrent()`
- AND si no hay ciclo activo muestra estado vacío con mensaje "No hay ciclo activo"

### Requirement: Guards de escritura y 9-box sobre ciclo activo
Los stores de metas, competencias y evaluación SHALL validar que el ciclo objetivo es el activo antes de enviar la escritura y SHALL mostrar error legible si el backend responde `cycle-not-active`. La vista 9-box SHALL operar solo sobre el ciclo activo y mostrar tabs "Avance" y "Cierre".

#### Scenario: escritura bloqueada sin ciclo activo
- WHEN no hay ciclo activo y el usuario intenta crear una meta
- THEN la UI bloquea la acción con mensaje "No hay ciclo activo"
- AND no se emite petición al backend

#### Scenario: 9-box con tabs del ciclo activo
- WHEN el jefe abre 9-box con ciclo activo en fase `avance`
- THEN ve los tabs "Avance" y "Cierre" con datos del ciclo activo
- AND no puede consultar ciclos cerrados desde esa vista
