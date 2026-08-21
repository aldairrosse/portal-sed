# Architecture — Rule

## Boundary
- `.opencode/**` changes ONLY. Never modify `web/` or `api/` without explicit `ARCHITECTURE` approval via Design System flow.
- Preserve existing logic, APIs (`web/src/lib/api/**`), routes (`web/src/routes/**`), store behavior (`web/src/lib/stores/*.svelte.ts`).

## Change separation
- `TOKEN` / `COMPONENT` / `PAGE` can proceed after DESIGN SYSTEM PROPOSAL approval.
- `ARCHITECTURE` (new route, store, schema, layout, data flow) → STOP, request approval, document in proposal.

## Workflow gate
- Analyze before modify. Reuse existing components first.
- No React, no new UI library. DaisyUI + Tailwind 4 only.

## Checks
- [ ] No `web/` or `api/` file modified without approval
- [ ] Change labeled correctly
- [ ] Existing component reused or justification documented
- [ ] No new dependency without approval

## Prohibited
- Refactor unrelated code
- Rename/move outside scope
- Add abstraction for single use
