## Context

El sistema actual de metas tiene `targetValue` y `currentValue` en la tabla `goals`, pero:
- No hay campo `direction` — todas las metas asumen "más es mejor".
- No hay `baselineValue` — no se puede calcular progreso invertido para metas descendentes.
- El scoring ponderado completo (categoría × meta × cumplimiento) no existe — solo un promedio simple en frontend.
- Los KPIs (`kp_is`) no tienen `currentValue` ni `direction` — son catálogos estáticos sin tracking de progreso.
- No hay endpoint para actualizar KPIs externamente.

El cambio toca: modelo de datos (Ent schema + migración), servicio de metas, servicio de KPIs, handlers REST, store Svelte, y componentes UI.

## Goals / Non-Goals

**Goals:**
- Score ponderado completo: `Σ(cat_weight × Σ(goal_weight × progress%))` calculable en backend y frontend.
- Dirección ascendente/descendente en metas con `baselineValue` para inversión.
- KPIs con `currentValue`, `direction`, y endpoint de actualización externa.
- Indicadores +/- en frontend cuando `currentValue` supera o empeora el target.
- Migración de datos existentes (metas seed sin dirección → default `ascendente`).

**Non-Goals:**
- Cambiar la lógica de evaluación final (cierre) — solo se agrega el score como dato adicional.
- Dashboard de analytics o reportes — el score es un número, no un dashboard.
- Autenticación de endpoints de KPI — se implementará cuando auth esté listo.
- Cambiar la regla de doble ponderación 100% — se mantiene intacta.

## Decisions

### D1: Campo `direction` como enum en Goal y KPI

**Decisión**: Agregar `direction` (`ascendente` | `descendente`) como campo enum en la tabla `goals` y `kp_is`.

**Alternativa considerada**: Usar un campo booleano `invertProgress` — más simple pero menos expresivo. Se descartó porque `direction` es más claro semánticamente y permite extender a más direcciones en el futuro (ej: `estable`).

**Razón**: El enum es idéntico en Go (`field.Enum`) y TypeScript (`'ascendente' | 'descendente'`). Default `ascendente` para no romper datos existentes.

### D2: `baselineValue` solo para dirección descendente

**Decisión**: `baseline_value` es nullable en BD. Solo se requiere cuando `direction === 'descendente'`. Para ascendente, se ignora (nullable).

**Fórmula de cumplimiento**:
```
// Ascendente (default)
progress% = min(currentValue / targetValue * 100, 100)

// Descendente
progress% = min((baselineValue - currentValue) / (baselineValue - targetValue) * 100, 100)
// Si currentValue <= targetValue → 100% cumplido
// Si currentValue >= baselineValue → 0% cumplido
```

**Alternativa considerada**: Siempre requerir `baselineValue` — más consistente pero obliga a los usuarios a definir un valor inicial innecesario para metas ascendentes. Se descartó porque la mayoría de metas son ascendentes y el campo sería confuso.

### D3: Scoring ponderado como cálculo derivado (no persistido)

**Decisión**: El score se calcula on-demand, no se almacena en BD. Razón: el `currentValue` cambia durante `avance`, y persistir el score crearía problemas de consistencia.

**Fórmula**:
```
employee_score = Σ(
  category.weight / 100 * Σ(
    goal.weight / 100 * goalProgressPercent(goal)
  )
)
```

**Dónde se calcula**:
- Backend: `ScoringService.GetEmployeeScore()` — para reporting y cierre.
- Frontend: `getWeightedScore()` en goalsStore — para UI en tiempo real.

### D4: Endpoint `PATCH /kpis/{kpiId}/value` para actualización externa

**Decisión**: Los KPIs se actualizan vía un endpoint PATCH dedicado, no en el CRUD de KPIs. Esto separa la actualización de valores (externa, periódica) de la administración de catálogo (manual, rara).

**Request body**:
```json
{ "current_value": 85.5 }
```

**Validación**: `current_value >= 0`. No se requiere autenticación en este change (se agregará cuando auth esté listo).

**Alternativa considerada**: Usar el mismo endpoint `PUT /kpis/{kpiId}` — más simple pero mezcla concerns (admin vs data feed). Se descartó porque el usuario pidió explícitamente un campo separado para el avance.

### D5: Indicadores +/- como cálculo frontend-only

**Decisión**: Los indicadores de exceso/déficit (`+15%` o `-8.2`) son puramente visuales en el frontend. No se persisten ni se envían al backend.

**Lógica**:
```
// Para ascendente: si currentValue > targetValue → "+exceso"
delta = currentValue - targetValue
indicator = delta > 0 ? `+${delta}` : delta < 0 ? `${delta}` : ''

// Para descendente: si currentValue < targetValue → "+cumplido" (bien)
// si currentValue > baselineValue → "-excedido" (mal)
```

**Razón**: El backend no necesita esta información; es presentación pura. Si en el futuro se necesita para reporting, se puede derivar de los campos existentes.

### D6: Migración de datos existentes

**Decisión**: 
- `goals.direction` → default `'ascendente'` en migración.
- `goals.baseline_value` → NULL (no aplica para ascendente).
- `kp_is.direction` → default `'ascendente'`.
- `kp_is.current_value` → NULL (sin dato inicial).
- Seeds actuales se actualizan para incluir direction explícita.

### D7: Reutilizar lógica de cumplimiento para KPIs

**Decisión**: La función `goalProgressPercent()` se generaliza a `progressPercent(current, target, baseline, direction)` y se reutiliza tanto para metas como para KPIs. No se duplica la lógica.

**Ubicación**: 
- Backend: `internal/pkg/scoring/progress.go`
- Frontend: `$lib/utils/scoring.ts`

## Risks / Trade-offs

- **[R1] Complejidad del formulario de meta** → El radio de dirección y el input condicional de baseline增加了 cognitive load. Mitigación: solo mostrar `baselineValue` cuando se selecciona "descendente", con helper text que explique qué es.
- **[R2] Metas descendentes con baseline = target** → División por cero si `baselineValue == targetValue`. Mitigación: validación en backend y frontend que `baselineValue > targetValue` para descendente.
- **[R3] Endpoint de KPI sin auth** → Temporalmente abierto. Mitigación: documentado como temporary; se cerrará con el change de auth. Para producción, agregar rate limiting mínimo.
- **[R4] Score calculado vs persistido** → Si el `currentValue` cambia frecuentemente, el score se recalcula en cada request. Mitigación: el cálculo es O(n) con n = número de metas (típico: 5-15); no es un hotspot. Si en el futuro se necesita, se puede cachear con TTL corto.
- **[R5] Breaking change en store** → El tipo `Goal` cambia (campo nuevo `direction`). Mitigación: default `ascendente` en fixtures y normalización en `normalizeApiData()`. Los fixtures existentes funcionan sin cambios.

## Migration Plan

1. **Migración Ent**: Agregar campos `direction` (enum, default 'ascendente') y `baseline_value` (float64, nullable) a tabla `goals`. Agregar `direction` (enum, default 'ascendente') y `current_value` (float64, nullable) a tabla `kp_is`.
2. **Seeds**: Actualizar `seed/goals.go` para incluir `direction` explícita en metas descendentes (ej: "Reducir ausentismo" → descendente, baseline=100).
3. **Backend services**: 
   - `GoalService.CreateGoal`: validar `baselineValue` requerido para descendente.
   - `ProgressService.UpdateGoalProgress`: usar `progressPercent()` con dirección.
   - Nuevo `ScoringService`: calcular score ponderado.
   - Nuevo handler `PATCH /kpis/{kpiId}/value`.
4. **Frontend**:
   - Actualizar tipo `Goal` y `KPI` en `types/goal.ts`.
   - Formulario de meta: radio direction + input condicional baseline.
   - `GoalRow`: indicador +/-.
   - Nuevo componente `KpiProgressBar` o extender `KpiBadge`.
   - Store: `getWeightedScore()`, `progressPercent()`.
5. **Fixtures**: Actualizar `goals.json` y `kpis.json` con direction y currentValue.

**Rollback**: Migración reversible (campos added con defaults). Si el rollback se activa, los campos new se ignoran por el código anterior.

## Open Questions

- **Q1**: ¿Los KPIs seed actuales deben tener `currentValue` inicial? Probablemente sí para que la UI muestre datos realistas en dev.
- **Q2**: ¿El endpoint de KPI value update debe soportar batch (múltiples KPIs de una vez)? El usuario pidió "una vez por etapa", lo que sugiere batch. Se puede agregar después.
- **Q3**: ¿El scoring ponderado debe mostrar desglose por categoría en la UI de cierre, o solo el score final? El usuario no especificó; se puede decidir en implementación.
