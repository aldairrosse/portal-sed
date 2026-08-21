# Svelte 5 — Rule

Scope: `web/src/**/*.svelte`, `web/src/**/*.svelte.ts`

## Stack
- Svelte 5 runes only: `$state`, `$derived`, `$effect`, `$props`, `$inspect`. No `export let`, no `on:` legacy, no `createEventDispatcher` legacy.
- SvelteKit file-based routing: `web/src/routes/**/+page.svelte`, `+layout.svelte`, `+page.ts`.
- Stores: `*.svelte.ts` with runes (`$state`/`$derived`). No `writable`/`readable` stores for new code.
- Types from `web/src/lib/api/schemas/*.d.ts` (generated via `openapi-typescript`).

## Checks
- [ ] No Svelte 4 syntax (`export let`, `$$props`, `on:click`)
- [ ] Props typed via `$props()` with interface
- [ ] Async state handled via runes, not manual subscriptions
- [ ] Reuses existing component under `web/src/lib/components/` when applicable

## Prohibited
- React / JSX / `lucide-react` in Svelte files (use `@lucide/svelte`)
- New store that duplicates existing `web/src/lib/stores/*.svelte.ts`
- Modifying `api/` from a Svelte rule
