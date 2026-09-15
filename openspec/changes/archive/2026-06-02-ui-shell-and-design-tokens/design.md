## Context

Repositorio SED en fase spec-only; `web/` existe como placeholder. El roadmap UI-first (A1 en `SPEC-ROADMAP.md`) requiere shell navegable antes de pantallas de negocio. Stack obligatorio: Vite, Svelte 5, TypeScript, DaisyUI, pnpm.

## Goals / Non-Goals

**Goals:**

- Scaffold `web/` con SvelteKit (modo SPA, `ssr = false`) + Tailwind + DaisyUI.
- Layout autenticado simulado reutilizable por todos los changes A2–A7.
- Stores globales de perfil y fase de ciclo consumibles por rutas futuras.
- Tema SED documentado en tokens DaisyUI.
- Componentes de estado UI (skeleton, vacío, error, 403).

**Non-Goals:**

- Login, SSO, cookies, API.
- Datos de evaluación reales o fixtures de negocio.
- i18n completo (solo español hardcoded).
- Tests E2E (solo preparar estructura para Vitest en tasks).

## Decisions

### 1. SvelteKit SPA vs router manual

**Decisión:** SvelteKit con `ssr = false` y rutas en `src/routes/`.

**Rationale:** Lazy loading nativo por `+page.svelte`, alineado a Vite + Svelte 5; evita mantener router aparte.

**Alternativa descartada:** `svelte-spa-router` — menos convención de proyecto para escala futura.

### 2. Estado dev: stores Svelte 5 runes

**Decisión:** `src/lib/stores/devContext.svelte.ts` con `$state` para `persona` y `cyclePhase`; persistir en `sessionStorage` solo en dev.

**Rationale:** Svelte 5 runes; sessionStorage permite recargar sin perder contexto en desarrollo.

### 3. Menú por perfil: configuración declarativa

**Decisión:** `src/lib/nav/menuConfig.ts` — array de ítems con `profiles: EvaluationProfile[]` y `phases?: CyclePhase[]`.

**Rationale:** Un solo lugar para matriz perfil × fase; changes futuros solo añaden entradas.

**Matriz inicial (placeholder routes):**

| Ruta | Etiqueta | Perfiles |
|------|----------|----------|
| `/` | Inicio | todos |
| `/objetivos/asignacion` | Asignación anual | colaborador, jefe, vendedor, gerente-tienda, divisional, regional, director |
| `/objetivos/avance` | Avance medio año | idem |
| `/mi-evaluacion` | Mi evaluación | todos excepto rh |
| `/mis-evaluados` | Mis evaluados | jefe, gerente-tienda, divisional, regional, director |
| `/evaluacion/9x9` | Matriz 9×9 | jefe, divisional, regional, director |
| `/rh/competencias` | Competencias | rh |
| `/rh/evaluaciones` | Evaluaciones RH | rh |

### 4. Design tokens: tema DaisyUI custom

**Decisión:** Tema `sed` en `tailwind.config` / `app.css` con variables:

- `--color-primary`: `#1e3a5f` (azul corporativo placeholder)
- `--color-secondary`: `#4a90a4`
- `--color-accent`: `#c4a035`
- `--radius-box`: `0.375rem`
- Fuente: `system-ui, "Segoe UI", sans-serif`
- Desactivar sombras: `--depth: 0` o override clases DaisyUI sin `shadow-*` decorativo

**Rationale:** Cumple `principles/styles-and-ui.md`; valores placeholder revisables con marca real.

### 5. Dev toolbar

**Decisión:** Componente `DevToolbar.svelte` fijo en footer o barra superior, renderizado solo si `import.meta.env.DEV`.

**Rationale:** Selectores de persona (8) y fase (3) siempre visibles en dev sin contaminar prod build.

**8 perfiles:** `colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `rh`.

**3 fases:** `inicio-anio`, `medio-anio`, `fin-anio`.

### 6. Estructura de carpetas `web/`

```
web/
├── src/
│   ├── routes/
│   │   ├── +layout.svelte          # AppShell
│   │   ├── +page.svelte            # Dashboard placeholder
│   │   ├── objetivos/...
│   │   ├── mi-evaluacion/...
│   │   ├── mis-evaluados/...
│   │   ├── evaluacion/9x9/...
│   │   └── rh/...
│   ├── lib/
│   │   ├── components/
│   │   │   ├── AppShell.svelte
│   │   │   ├── Sidebar.svelte
│   │   │   ├── DevToolbar.svelte
│   │   │   └── ui/                 # EmptyState, ErrorState, ForbiddenState, PageSkeleton
│   │   ├── nav/menuConfig.ts
│   │   ├── stores/devContext.svelte.ts
│   │   └── types/evaluation.ts     # EvaluationProfile, CyclePhase
│   └── app.css
├── tailwind.config.js
├── package.json
└── svelte.config.js
```

## Risks / Trade-offs

| Riesgo | Mitigación |
|--------|------------|
| Menú por perfil simula RBAC | Comentario en código + spec: no es seguridad; change `identity-access` reemplaza |
| Valores de tokens placeholder | Documentar en README web que marca final pendiente |
| SvelteKit vs SPA pura Vite | `ssr = false` en layout raíz; validar build estático |
| sessionStorage en dev filtra a prod | Guard `import.meta.env.DEV` en lectura/escritura |

## Migration Plan

1. Implementar scaffold y shell (este change).
2. Changes A2–A7 montan contenido en rutas placeholder.
3. Change `identity-access`: ocultar DevToolbar en prod; reemplazar persona por sesión API.
4. Rollback: revertir carpeta `web/`; sin impacto en `api/`.

## Open Questions

- Colores corporativos SED definitivos (pendiente marca).
- ¿Incluir `director-general` como noveno perfil en change futuro o mapear a `director`?
