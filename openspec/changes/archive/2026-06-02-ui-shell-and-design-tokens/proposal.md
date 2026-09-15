## Why

El portal SED aún no tiene aplicación frontend; sin un shell común no se pueden validar flujos ni el roadmap UI-first (A2–A7). Este change establece layout, navegación por perfil, tokens visuales y herramientas de desarrollo (selector de persona y fase de ciclo) para iterar pantallas con fixtures antes de backend o autenticación real.

## What Changes

- Scaffold inicial de `web/` (Vite + Svelte 5 + TypeScript + DaisyUI + Tailwind) si no existe.
- Layout autenticado simulado: barra superior, menú lateral, área de contenido con rutas lazy-loaded.
- Menú dinámico según **perfil de evaluación** activo (ítems visibles/ocultos por perfil, sin RBAC servidor).
- **Design tokens** SED en tema DaisyUI (colores primario/secundario, radius, tipografía base).
- **Selector de persona** (solo `import.meta.env.DEV`): 8 perfiles de evaluación.
- **Selector de fase de ciclo:** inicio de año, medio año, fin de año (contexto global en store).
- Páginas placeholder por módulo del roadmap (inicio, medio, fin, RH, mis evaluados, 9×9) enlazadas desde menú.
- Estados UI reutilizables: skeleton, vacío, error de red, sin permiso (403 simulado).
- Convención **sentence case** en etiquetas de UI.

## Non-goals

- SSO, login, sesión httpOnly, JWT.
- API REST, OpenAPI implementado, PostgreSQL, persistencia.
- Fixtures de evaluación, metas o competencias (change A2+).
- RBAC en backend; la UI no debe asumir que ocultar menú es seguridad.
- Notificaciones por correo.
- Matriz 9×9 funcional (solo enlace de menú si aplica al perfil jefe).

## Capabilities

### New Capabilities

- `ui-shell`: layout autenticado simulado, navegación lazy, menú por perfil, integración tokens.
- `dev-persona-cycle`: selector de persona (8 perfiles) y fase de ciclo en desarrollo; stores Svelte compartidos.
- `design-tokens`: tema DaisyUI/Tailwind SED (colores, radius, tipografía, reglas sin box-shadow decorativo).

### Modified Capabilities

- _(ninguna — `openspec/specs/` vacío)_

## Actores y flujo

| Actor | Precondición | Flujo feliz |
|-------|--------------|-------------|
| Desarrollador / diseñador | App en modo dev | Abre app → elige perfil y fase → navega menú → ve placeholder del módulo |
| Usuario final (futuro) | Fuera de alcance este change | — |

**Errores:** ruta inexistente → página 404; perfil sin ítem de menú → ítem no visible (no 403 real hasta auth).

## Impact

- Nuevo paquete `web/` con estructura `src/routes`, `src/lib/components`, `src/lib/stores`, `src/lib/dev`.
- `principles/styles-and-ui.md` — tokens concretos referenciados en design.
- `openspec/specs/` — specs duraderas al archivar: `ui-shell`, `dev-persona-cycle`, `design-tokens`.
- Sin cambios en `api/`.
- Dependencia para changes A2–A7 del [SPEC-ROADMAP](../SPEC-ROADMAP.md).
