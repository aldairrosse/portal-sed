# Visual Quality — Rule

Scope: `web/src/**/*.svelte`, `web/src/app.css`

## Principles
- Flat by default (no shadow/border on `card`/`btn`/`input`/`navbar` per `app.css` reset). Elevate only with `card-elevated` / `soft-shadow` / `card-shadow`.
- No decorative `box-shadow`, no `border-left/right` only. Use tokens.
- Typography: `Geist` body, `Binjay` (`font-binjay`) for brand/logo only. Hierarchy via `@layer base` h1-h3.

## Checks
- [ ] No ad-hoc `box-shadow` or single-side border
- [ ] Shadows only via `soft-shadow` / `card-shadow` / `card-elevated`
- [ ] Spacing uses Tailwind scale, not arbitrary px
- [ ] Icons from `@lucide/svelte`, consistent size
- [ ] Visual diff vs Figma ≤ 2px where Figma exists

## Prohibited
- Hardcoded `style="box-shadow: ..."` / `style="border-left: ..."`
- Mixing `lucide-react` in Svelte (use `@lucide/svelte`)
- Per-page color overrides that should be tokens
