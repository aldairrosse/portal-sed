---
name: design-system
description: Propone/actualiza Design Tokens y componentes base compatibles Tailwind 4 + DaisyUI. Fuente para TOKEN/COMPONENT changes.
license: MIT
compatibility: Tailwind 4 via @tailwindcss/vite, DaisyUI themes in app.css
metadata:
  author: portal-sed
  version: "1.0"
---

# Design System — Skill

Trigger: `design system`, `tokens`, `propuesta ds`, o paso DESIGN SYSTEM PROPOSAL.

## Input
- Audit report previo (requerido si existe).

## Steps
1. **Define tokens** compatibles Tailwind 4 + DaisyUI en `web/src/app.css`:
   - Colores → `--color-*` en `@plugin "daisyui/theme" { name: "sed" }` y `sed-dark`
   - Typography → `@layer base` + `@utility font-binjay`
   - Spacing/radius/shadows/breakpoints/motion → CSS var + `@utility` (ej. `soft-shadow`, `card-shadow`)
   - Estados (hover/focus/disabled) y motion (150ms ease) documentados.
2. **Componentes base**: mapea cada hallazgo COMPONENT a DaisyUI (`btn`, `card`, `input`, `navbar`, etc.). Solo crea nuevo si DaisyUI no cubre.
3. **Proposal doc**: tabla token actual → propuesto + componente afectado + archivo `app.css` target.
4. **No implementar** hasta aprobación USER FEEDBACK. Proposal es solo doc.

## Outputs
- `design-system-proposal.md`: tokens, typography, spacing, radius, shadows, breakpoints, estados, motion, componentes base.

## Checks
- [ ] Cada token en `app.css` theme, no hardcode en componente
- [ ] No `tailwind.config.*` creado
- [ ] DaisyUI preferido antes de custom

## Prohibited
- Hardcode fuera de `app.css`
- Nuevo componente que duplica DaisyUI
