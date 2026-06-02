## 1. Scaffold web

- [x] 1.1 Inicializar SvelteKit en `web/` con TypeScript, Tailwind, DaisyUI y pnpm (`ssr = false` en layout raíz)
- [x] 1.2 Configurar ESLint + Prettier alineados al repo
- [x] 1.3 Actualizar `web/README.md` con comandos `pnpm dev`, `pnpm lint`, `pnpm check`

## 2. Design tokens

- [x] 2.1 Definir tema DaisyUI `sed` en Tailwind/app.css (primary, secondary, accent, radius, tipografía)
- [x] 2.2 Aplicar reglas sin box-shadow decorativo y sentence case en componentes base
- [x] 2.3 Verificar contraste WCAG AA y foco visible en botones y enlaces del shell

## 3. Tipos y contexto dev

- [x] 3.1 Crear `src/lib/types/evaluation.ts` (`EvaluationProfile`, `CyclePhase`, labels español)
- [x] 3.2 Implementar `devContext.svelte.ts` con runes, defaults y persistencia sessionStorage (solo DEV)
- [x] 3.3 Crear fixture estático de usuario simulado por perfil (nombre visible en barra)

## 4. Shell y navegación

- [x] 4.1 Implementar `menuConfig.ts` con matriz perfil × rutas según design.md
- [x] 4.2 Crear `AppShell.svelte`, `Sidebar.svelte` (drawer móvil, prefers-reduced-motion)
- [x] 4.3 Integrar shell en `+layout.svelte` con barra superior (usuario simulado + badge fase)
- [x] 4.4 Implementar `DevToolbar.svelte` (8 perfiles + 3 fases, solo DEV)

## 5. Rutas placeholder

- [x] 5.1 Página inicio `/` con resumen de contexto dev activo y demo de estados UI
- [x] 5.2 Rutas placeholder: `/objetivos/asignacion`, `/objetivos/avance`, `/mi-evaluacion`, `/mis-evaluados`, `/evaluacion/9x9`, `/rh/competencias`, `/rh/evaluaciones`
- [x] 5.3 Página `+error.svelte` 404 en español con enlace al inicio

## 6. Componentes de estado UI

- [x] 6.1 Crear `PageSkeleton.svelte`, `EmptyState.svelte`, `ErrorState.svelte`, `ForbiddenState.svelte` en `src/lib/components/ui/`
- [x] 6.2 Documentar uso en comentario breve o sección en página inicio

## 7. Verificación

- [x] 7.1 Ejecutar `pnpm check` y `pnpm lint` sin errores
- [x] 7.2 Probar manualmente: cambio de perfil filtra menú; cambio de fase actualiza badge; recarga restaura contexto en dev
- [x] 7.3 Confirmar DevToolbar ausente en `pnpm build` + preview producción
- [x] 7.4 Ejecutar `openspec validate --all` desde raíz del repo