# Accessibility — Rule (WCAG 2.1 AA)

Scope: `web/src/**/*.svelte`, `web/src/app.css`

## Required
- Semantic HTML: `button` for actions, `a` for navigation, `label` ↔ `input` association.
- Keyboard: all interactive elements reachable via Tab, operable via Enter/Space, visible `:focus-visible`.
- Contrast: text ≥ 4.5:1, large text ≥ 3:1. Verify against both `sed` and `sed-dark` themes.
- Images/icons: `alt` for meaningful, `aria-hidden="true"` for decorative. `Icon.svelte` must pass `aria-label` when meaningful.
- Forms: `aria-invalid` + `aria-describedby` on error, error text linked via `id`.

## Responsive
- Mobile-first. No horizontal scroll at 320px. Touch targets ≥ 44×44px.

## Checks
- [ ] Tab order logical, focus visible
- [ ] Contrast checked (both themes)
- [ ] Labels/alt/aria correct
- [ ] No `div`/`span` with click handler without `role="button"` + keyboard handler (prefer `button`)
- [ ] Responsive 320px / 768px / 1280px

## Prohibited
- `outline: none` without replacement focus style
- Color-only conveyance of information
- `aria-label` that duplicates visible text unnecessarily
