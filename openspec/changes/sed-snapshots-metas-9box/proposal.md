## Why

Las metas pierden su valor capturado por fase (avance/cierre) y el 9-box deriva calificaciones con reglas inconsistentes (promedios sin ponderar, ceros por ausencias, derivación `final_rating`). Se necesita snapshot por fase con valor directo según `unit` y reglas 9-box/competencias unificadas.

## What Changes

- REQ1: `evaluation_goals` gana `avance_progress` y `cierre_progress` (FLOAT NULL); valor directo de la meta según su `unit` (porcentaje/moneda/numero/binario) capturado en cada fase.
- REQ2: `goals.current_value = cierre_progress ?? avance_progress`; si el módulo de metas actualiza `current_value` (`UpdateGoalProgress`), se actualiza también el snapshot de la fase activa (`cycles.current_phase`: avance o cierre).
- REQ3: `PUT goal-state` (cierre) persiste el valor directo en el snapshot de la fase activa y sincroniza `current_value`; se elimina la derivación `final_rating = int(finalProgress*5)`. La UI muestra el valor por `unit`.
- REQ4: 9-box metas mantiene/valida el escalado `direction` (ascendente/descendente) → % completado (clamp 0–100) → escala 1–3 (`<34=1, <67=2, else 3`).
- REQ5: `RecomputeMatrix` usa promedio ponderado `0.8 RH / 0.2 self` (igual que `ComputeMatrixView`), ignorando valores ausentes en vez de tratarlos como 0.
- REQ6: comentarios de metas y competencias (self/jefe/rh) se guardan y muestran con nombre de autor, fecha y hora; autor desde la sesión.

## Capabilities

### New Capabilities

- `evaluations`: snapshots por fase (`avance_progress`/`cierre_progress`), sincronización `current_value`, cierre con valor directo y comentarios con autor/fecha.

### Modified Capabilities

- `ninebox-computed`: escalado de metas por `direction` y promedio ponderado 0.8/0.2 ignorando ausentes en `RecomputeMatrix`.
- `competency-framework`: promedio ponderado 0.8 RH / 0.2 self ignorando ausentes como regla de dominio de competencias.

## Impact

- BD: migración `evaluation_goals` (+2 columnas FLOAT NULL); backfill: snapshots NULL, `current_value` sin cambios hasta próxima escritura.
- API: `PUT goal-state` (cierre) cambia contrato (valor directo por `unit`, sin `final_rating` derivado); `UpdateGoalProgress` escribe snapshot de fase activa.
- UI: mostrar valor por `unit` en cierre; comentarios con autor/fecha/hora.
- Actores: empleado (self), jefe, RH; sesión aporta autor (REQ6). Precondición: ciclo con `current_phase` en avance/cierre. Flujo feliz: capturar snapshot → sincronizar `current_value` → 9-box lee snapshots/ponderado. Errores: fase desconocida → 400; `unit` no soportada → 400; sin sesión → 401. Sin notificaciones email en este change.
- Referencias: `principles/evaluations-domain.md`, `principles/data-and-orm.md`, `principles/contracts-api.md`. Decisión abierta a cerrar: backfill histórico de snapshots (se dejan NULL por defecto).

## Non-goals

- Sin cambios de ponderación de metas ni de catálogo de pilares; sin triggers email; sin migración de `final_rating` histórico.
