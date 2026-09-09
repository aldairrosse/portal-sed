# Hidden Settings Page + Light-Only Theme

## Goal

Add a hidden `/ajustes` page (no visible nav entry yet) that manages the theme (light/dark toggle, persisted), make **light the only default** for regular users (stop following `prefers-color-scheme`), and give the sidebar a strong token-based background that differs between light and dark.

## Non-goals

- Changing the actual color values of the main tokens (`primary`, `secondary`, `accent`, etc.) — values stay as-is until the user defines the new palette.
- Adding a visible link to `/ajustes` in `MENU_ITEMS` or the profile badge yet — decided later ("al final en el side o en perfil").
- Backend changes: no API, no OpenAPI, no DB. Pure frontend.
- Replacing daisyUI themes; both `sed` (light) and `sed-dark` remain, but `sed` is the forced default for everyone unless toggled in `/ajustes`.

## Scope

| Layer | Change |
|-------|--------|
| Theme default | `web/src/app.html` — remove the inline `prefers-color-scheme` script (lines 13-23); keep `data-theme="sed"` static so regular users only ever see light |
| Theme store | New `web/src/lib/stores/theme.ts` — Svelte writable, default `'light'`, persists to `localStorage`, applies `document.documentElement.dataset.theme = 'sed' | 'sed-dark'` |
| Page | New `web/src/routes/ajustes/+page.svelte` — section "Apariencia" with light/dark toggle bound to the theme store; skeleton-ready for future settings; requires auth (AppShell layout) |
| Sidebar tokens | `web/src/app.css` — add `--color-sidebar` + `--color-sidebar-content` to both themes: light = strong solid (e.g. deep brand `#101C82`-family), dark = solid dark slate |
| Sidebar | `web/src/lib/components/Sidebar.svelte` — switch `<aside>` background from `bg-base-100` to the new sidebar tokens (`bg-[var(--color-sidebar)]` or daisyUI-compatible usage); adjust hover/active states for contrast on the strong background |

## Gotchas (documented constraints)

- **RadarChart.svelte:190** observes `data-theme` to pick chart colors; with light-only default it stays on the light palette — fine, but the store must set `data-theme` on the `<html>` element (same attribute RadarChart reads).
- **No SSR flash**: apply the persisted theme before hydration in `+layout.svelte` (or inline) so the store doesn't flash dark on reload; keep it minimal.
- **AppShell**: `/ajustes` must live inside the existing AppShell layout (auth + drawer) — do NOT make it standalone like `/login`. Verify `+layout.svelte` grouping so the route inherits the shell.
- **Sidebar contrast**: with a strong background, text/borders inside the sidebar must use `--color-sidebar-content` (and derived translucent states) instead of `base-content`; check the profile badge + nav items + logout/footer.
- **No menu entry**: do not touch `web/src/lib/nav/menuConfig.ts` in this change; the page is URL-only for now.

## Precondition

`web/src/app.css` daisyUI theme blocks exist (lines 9-110). `web/src/lib/components/Sidebar.svelte` currently uses `bg-base-100` (line 64). SvelteKit file-based routing with AppShell in `web/src/routes/+layout.svelte`.

## Status

Draft — proposal only.
