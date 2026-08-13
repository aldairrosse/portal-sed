# Tasks: Hidden Settings Page + Light-Only Theme

- [ ] 1. `web/src/app.html`: remove the inline `prefers-color-scheme` script (lines 13-23); keep `<html data-theme="sed">` static
- [ ] 2. New `web/src/lib/stores/theme.ts`: writable store `'light' | 'dark'`, default `'light'`, `localStorage` persistence (`sed-theme` key), sets `document.documentElement.dataset.theme = 'sed' | 'sed-dark'`
- [ ] 3. `web/src/routes/+layout.svelte`: initialize/apply theme store (restore persisted value before render; no flash)
- [ ] 4. `web/src/app.css`: add `--color-sidebar` + `--color-sidebar-content` to `sed` (light: strong solid, e.g. `#101C82`-family) and to `sed-dark` (dark slate, e.g. `#0f172a`-family)
- [ ] 5. `web/src/lib/components/Sidebar.svelte`: `<aside>` background → new sidebar tokens; adjust nav hover/active + profile badge + logout/footer to use `--color-sidebar-content` and translucent variants for contrast on the strong background
- [ ] 6. New `web/src/routes/ajustes/+page.svelte`: section "Apariencia" with light/dark toggle bound to theme store; minimal skeleton for future settings (no nav entry yet)
- [ ] 7. Verify: `pnpm build` (or `pnpm lint` + `tsc` per AGENTS.md) passes; manual check that default is light and toggle persists on reload

## Verification notes

- `/ajustes` must inherit AppShell (auth guard) — not standalone.
- No changes to `web/src/lib/nav/menuConfig.ts` (page stays URL-only).
- RadarChart must still read `data-theme` correctly after the store takes over.
