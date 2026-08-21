# Tailwind 4 + DaisyUI — Rule

Scope: `web/src/app.css`, `web/src/**/*.svelte`

## Setup (project reality)
- Tailwind 4 via `@tailwindcss/vite` in `web/vite.config.ts` — no `tailwind.config.ts`.
- Single entry: `web/src/app.css` with `@import "tailwindcss"`, `@plugin "@tailwindcss/typography"`, `@plugin "daisyui" { themes: sed --default, sed-dark; }`.
- Themes defined via `@plugin "daisyui/theme" { name: "sed"; ... }` — all palette/radius/border vars there.

## Usage
- DaisyUI classes first: `btn`, `card`, `input`, `navbar`, `timeline`, etc. Extend via `@layer components` or `@utility` in `app.css`, not inline `<style>`.
- Global resets already in `app.css` (`.card`, `.btn`, `.input`, `.navbar` → `box-shadow: none`). Keep flat style; use `soft-shadow`/`card-shadow`/`card-elevated` only when elevation needed.
- Custom tokens → add CSS var in theme block + `@utility` if needed. Never hardcode hex in component.

## Tokens strategy (Tailwind 4 compatible)
- Colors: `--color-*` inside theme block
- Radius: `--radius-selector/field/box`
- Shadows/breakpoints/motion: `@utility` or CSS var in `app.css`
- Typography: `@layer base` h1-h3 + `@utility font-binjay`

## Checks
- [ ] No `tailwind.config.*` created
- [ ] No hardcoded color/hex in `.svelte`
- [ ] DaisyUI used before custom CSS
- [ ] New utility added to `app.css`, not scattered

## Prohibited
- Creating `tailwind.config.ts/js`
- `box-shadow` / `border-left/right` decorative only
- Inline `<style>` that duplicates theme var
