# Design System — Rule

Scope: `web/src/app.css`, `web/src/lib/components/**`, `web/src/routes/**`

## Tokens (single source)
- All color / spacing / radius / shadow / typography via DaisyUI theme variables in `web/src/app.css` (`@plugin "daisyui/theme"`). No hex/rgb/hsl literals in Svelte files.
- Spacing/radius/shadow/size → Tailwind scale or CSS var. No arbitrary `12px`, `17px`, `box-shadow:` literals. Use `p-4`, `rounded-box`, `soft-shadow` utility.

## Component reuse
- Check `web/src/lib/components/ui/*` and existing `web/src/lib/components/**` before creating new component. Prefer extending DaisyUI (`btn`, `card`, `input`, etc.).
- No duplicate component that DaisyUI already provides.

## Change types — must label PR/change
- `TOKEN` — only `app.css` theme vars / `@utility`
- `COMPONENT` — `web/src/lib/components/**` (isolated, no route logic)
- `PAGE` — `web/src/routes/**` composition only
- `ARCHITECTURE` — routes, stores (`*.svelte.ts`), api schemas, layout → requires approval, STOP and ask.

## Order
Fix Design System inconsistencies before page-level redesign. Pages consume tokens, not invent them.

## Checks
- [ ] No hardcoded color/spacing/radius/shadow/size outside `app.css`
- [ ] No duplicated component vs DaisyUI/existing
- [ ] Change labeled TOKEN|COMPONENT|PAGE|ARCHITECTURE
- [ ] If ARCHITECTURE → approval obtained

## Prohibited
- Hardcoded tokens in `.svelte`/`.ts`
- New UI lib, React import, or custom equivalent of DaisyUI component
- PAGE change that introduces new token (move to TOKEN first)
