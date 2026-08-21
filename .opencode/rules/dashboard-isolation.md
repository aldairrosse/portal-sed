# Dashboard Isolation — Rule

> Flujo INVERSO código → Figma para dashboard. Copia fiel del diseño existente usando data real. Scope: dashboard (`web/src/routes/+page.svelte` y similares `objetivos/`, `evaluacion/`, `mis-evaluados/`). Excluye `login`, redirects y auth guard (`web/src/routes/+layout.svelte`).

## Propósito
Migrar/refactorizar UI de dashboard sin inventar data: identificar primero la data real de la vista → generar Figma 1:1 → luego refactorizar. Preservar lógica, APIs, stores y rutas.

## Paso 1 — IDENTIFY DATA (antes de Figma)

Inspeccionar en orden, documentar en proposal:

| Fuente | Path real |
|--------|-----------|
| Page + load | `web/src/routes/+page.svelte` , `+page.ts` / `+layout.ts` (si existe `load`) |
| Stores | `web/src/lib/stores/*.svelte.ts` — ej. `goalsStore.svelte.ts`, `competencyStore.svelte.ts`, `cycleStore.svelte.ts`, `devContext.svelte.ts`, `evaluationStore.svelte.ts` |
| Session/cycle API | `web/src/lib/api/session.svelte.ts`, `web/src/lib/api/cycle.svelte.ts` |
| Schemas | `api/openapi/*.yaml` (`goals-api.yaml`, `cycle.yaml`, `evaluations-and-9x9.yaml`, etc.) |
| Props derivadas | `$derived` en `+page.svelte` (ej. `myGoals`, `myCompetencies`, `phase`, `personalStats`) |

Checklist:
- [ ] Listar cada `$derived`/`store` que alimenta la vista (ej. `getGoals()`, `getAssignments()`, `getPillars()`)
- [ ] Anotar tipos (`web/src/lib/types/*`) y schemas OpenAPI correspondientes
- [ ] Confirmar qué data ya existe en stores vs. qué requeriría nuevo fetch (si requiere fetch nuevo → ARCHITECTURE)
- [ ] No inventar props: toda prop de Figma debe mapear 1:1 a campo real

## Paso 2 — ISOLATE VIEW (View Component puro)

Extraer UI de `+page.svelte` a componente puro sin crear lógica nueva:

```
web/src/lib/views/[Nombre]View.svelte   // ej. DashboardView.svelte
```

Patrón:
```svelte
<script lang="ts">
  let { user, myGoals, pillars, phase, ... } = $props(); // solo data real
</script>
<!-- markup copiado 1:1 de +page.svelte, sin lógica nueva -->
```

- `+page.svelte` queda como contenedor delgado: importa stores/data real y hace `<DashboardView {user} {myGoals} ... />`.
- View no importa stores directamente; recibe todo por `$props()`.
- No mock inventado: si necesitas renderizar sin auth, el contenedor inyecta la misma data (ver Paso 4).
- Clasificación: `COMPONENT` si solo mueve markup; `ARCHITECTURE` si cambia data flow.

## Paso 3 — CODE → FIGMA (con data real renderizada)

Ejecutar con Figma MCP sobre la vista renderizada con data real:

1. `figma_get_libraries` + `figma_search_design_system` — reutilizar DS existente.
2. `figma_generate_figma_design` — captura pixel-perfect de la vista con data real.
3. `figma_use_figma` — construir en Figma desde DS (DaisyUI tokens de `web/src/app.css` → `bg-primary`, `text-base-content`, etc.).
4. Si hay Code Connect (`figma_get_code_connect_map`), preferir `web/src/lib/components/**` reales.
5. Refinar `use_figma` contra captura; borrar captura tras refinar.

Tokens: usar solo vars de `app.css` (`@plugin "daisyui/theme"`). No hardcodear hex.

Done de este paso: Figma refleja 1:1 estructura, tokens y data de la vista real.

## Mock aislado (solo si hace falta bypass auth)

Para renderizar dashboard sin pasar por `+layout.svelte` guard (`goto('/login')`):

- Opción: `VITE_MOCK_SESSION=1` o `?mock=1` que inyecta `session.user` y `devContext` en dev — **solo en `web/src/routes/dev/` o guard condicional `import.meta.env.DEV`**.
- Marcado como `ARCHITECTURE` — requiere aprobación explícita. Limitado a dashboard, no tocar `login` ni redirects.
- Alternativa sin código: usar usuario de dev ya existente en `devContext.svelte.ts` / `activityLogStore` si disponible.

## Qué NO hacer

- No inventar mock data distinta a stores/schemas reales.
- No tocar `web/src/routes/login/`, `+layout.svelte` guard, ni redirects (fuera de scope).
- No crear `tailwind.config.*` (proyecto usa `@tailwindcss/vite`, ver `tailwind-daisyui` rule).
- No crear abstracciones/helpers de un solo uso.
- No añadir dependencia UI nueva.

## Criterio de done

- [ ] Figma 1:1 con vista real: misma estructura, mismos tokens DaisyUI, misma data (no props inventadas)
- [ ] View Component existe en `web/src/lib/views/` con `$props()` mapeando a stores reales
- [ ] `+page.svelte` solo inyecta stores/data real
- [ ] Si hubo mock, marcado `ARCHITECTURE` y limitado a dev/dashboard
- [ ] Sin cambios en `login`/auth guard sin aprobación
