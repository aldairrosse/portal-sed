## Context

A3 entregó la pantalla de inicio de año (`/objetivos/asignacion`) con modo editor (dueño) y lector (jefe). El store `goalsStore.svelte.ts` maneja CRUD de categorías, metas, KPIs y change requests con fixtures JSON. Los tipos están en `goal.ts`.

A4 extiende esta misma pantalla para la fase "avance" del ciclo: el empleado edita avances en sus metas, y el jefe puede editar avances y comentar en metas de subordinados. No se puede eliminar ni editar pesos.

## Goals / Non-Goals

**Goals:**
- Modo "avance" en `/objetivos/asignacion` que se activa cuando `cyclePhase === 'avance'`
- Campo `progress` (number 0–100 o valor absoluto según `GoalUnit`) en cada meta
- Indicador visual de progreso (barra + badge) por meta y por categoría
- Campo `comments` (texto libre) por meta, accesible por empleado y jefe
- Bloqueo de eliminar metas y editar pesos/valores objetivo en fase avance
- Reutilizar componentes existentes (CategoryCard, GoalRow, WeightIndicator) con variantes condicionales

**Non-goals:**
- Evaluación 1-5 de competencias (A5)
- Matriz 9×9 (A6)
- Eliminar metas (prohibido en medio año, decisión #3)
- Persistencia, API, auth
- Edición de KPIs vinculados (solo avance y comentarios)

## Decisions

### 1. Extender Goal con campos de avance

Agregar al tipo `Goal`:
```ts
progress?: number;        // 0–100 para porcentaje, valor absoluto para moneda/numero
progressUpdatedAt?: string; // ISO date
comments?: GoalComment[];
```

Nuevo tipo:
```ts
interface GoalComment {
  id: string;
  authorId: string;
  authorName: string;
  content: string;
  createdAt: string;
}
```

**Razón:** Mantener avance y comentarios en la misma entidad Goal simplifica el store y evita joins innecesarios. `progress` es opcional para no romper fixtures de inicio de año.

### 2. CyclePhase como tipo separado

```ts
type CyclePhase = 'asignacion' | 'avance' | 'cierre';
```

Agregar `cyclePhase` al store (no a cada goal) como estado global. La página lee `getCyclePhase()` para decidir el modo.

**Razón:** La fase es una propiedad del ciclo, no de cada meta. Un solo state global es más simple que propagar por entidad.

### 3. Permisos por rol en el store

Función `getGoalPermissions(role, isOwner)` que retorna:
```ts
{ canEditProgress: boolean; canComment: boolean; canEditWeight: boolean; canDelete: boolean; }
```

En fase `asignación`: empleado tiene permisos completos (CRUD). En fase `avance`: solo avance + comentarios. El jefe siempre tiene `canEditProgress: true` y `canComment: true` en subordinados.

**Razón:** Centralizar permisos en el store evita lógica condicional dispersa en componentes.

### 4. Componentes nuevos vs modificar existentes

| Componente | Acción | Razón |
|---|---|---|
| `GoalRow.svelte` | Modificar | Agregar campo avance inline y botón comentario. Ya renderiza la fila. |
| `CategoryCard.svelte` | Modificar | Agregar progress indicator por categoría. Ya es el contenedor. |
| `ProgressIndicator.svelte` | **Crear** | Barra de progreso reutilizable (meta individual y global). |
| `CommentPopover.svelte` | **Crear** | Popover con lista de comentarios y textarea para agregar. |

**Razón:** No duplicar la estructura de tabla. GoalRow ya tiene la fila; agregar variantes es más limpio que crear GoalProgressRow separado.

### 5. Fixtures extendidos

`goals.json` se extiende con `progress` y `comments` opcionales. Nuevo archivo `cycle.json`:
```json
{ "year": 2026, "phase": "avance" }
```

El store carga `cycle.json` para determinar el modo. En dev, se puede cambiar la fase manualmente.

### 6. Indicador de progreso por categoría

`WeightIndicator` ya existe para pesos. Para avance, crear variante en `ProgressIndicator` que muestre:
- Barra de progreso (promedio de avance de metas de la categoría)
- Badge con porcentaje promedio
- Color: rojo (<40%), amarillo (40–79%), verde (≥80%)

**Razón:** Separar indicador de peso (A3) de indicador de avance (A4) mantiene componentes con una sola responsabilidad.

## Risks / Trade-offs

- **[Riesgo]** GoalRow acumula mucha lógica (peso, avance, comentarios, permisos). → **Mitigación:** Extraer `CommentPopover` y `ProgressIndicator` como componentes separados que GoalRow compone.
- **[Riesgo]** Conflicto entre modo editor (A3) y modo avance (A4) en la misma página. → **Mitigación:** El store expone `getCyclePhase()` y la página usa un `{#if}` block para renderizar el modo correcto. No hay superposición de modos.
- **[Trade-off]** `progress` como campo en Goal vs tabla separada GoalProgress. → **Decisión:** Campo en Goal porque es UI-first sin persistencia; cuando llegue la API se puede normalizar.
- **[Trade-off]** Comentarios inline vs modal separado. → **Decisión:** Popover inline (click en ícono de comentario) para reducir fricción y mantener contexto.
