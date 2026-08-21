---
name: redesign
description: Workflow completo de rediseño UI/UX — 8 pasos estrictos. No modifica web/api hasta aprobación.
---

# Workflow: Redesign (8 pasos)

> Secuencia exacta: ANALYZE → UI AUDIT → DESIGN SYSTEM PROPOSAL → FIGMA → USER FEEDBACK → IMPLEMENT → VISUAL VALIDATION → ACCESSIBILITY CHECK + FINAL REVIEW

## 1 — ANALYZE
- Inventaria `web/src/routes`, `components`, `stores/*.svelte.ts`, `app.css`, `api/openapi/**`.
- Identifica componentes reutilizables, APIs y rutas a preservar. Output: `analyze.md` (scope + reuse list).

## 2 — UI AUDIT
- Delega a `ui-audit` skill. Inconsistencias DS primero (TOKEN/COMPONENT antes de PAGE). Output: `audit-report.md`.

## 3 — DESIGN SYSTEM PROPOSAL
- Delega a `design-system` skill. Tokens Tailwind4+DaisyUI (`app.css` theme), typography, spacing, radius, shadows, breakpoints, estados, motion, componentes base. Output: `design-system-proposal.md`. No implementar aún.

## 4 — FIGMA
- Delega a `figma` skill. Detecta MCP (`figma_whoami`); si ok: lee variables/componentes/estilos, lleva UI a Figma (`search_design_system` → `use_figma` + `generate_figma_design`), review. Fallback sin Figma documentado. Output: Figma link + `figma-sync.md`.

## 5 — USER FEEDBACK
- Pausa. Presenta proposal + Figma. Espera aprobación explícita. Ambiguos → `NEEDS CLARIFICATION`, no implementar. Gate: sin aprobación no avanzar.

## 6 — IMPLEMENT
- Tras aprobación. Orden: TOKEN (`app.css`) → COMPONENT (`components/**`) → PAGE (`routes/**`). ARCHITECTURE → STOP, pide aprobación. Output: diffs etiquetados.

## 7 — VISUAL VALIDATION
- Delega a `visual-validation` skill. Ciclo IMPLEMENT → RUN → CAPTURE/INSPECT → COMPARE WITH FIGMA → IDENTIFY DIFFS → FIX → RECHECK hasta pass. Output: `visual-validation-report.md`.

## 8 — ACCESSIBILITY CHECK + FINAL REVIEW
- WCAG 2.1 AA (contrast sed/sed-dark, keyboard, focus, labels, 320px) + review visual final. Si falla → loop a FIX + RECHECK. Output: `a11y-checklist.md` + `final-review.md`.

## Gates
- No saltar pasos. No PAGE sin TOKEN/COMPONENT. No implementar ambiguo.

## Rules linked
- Todos en `.opencode/rules/` (design-system, svelte, tailwind-daisyui, accessibility, visual-quality, architecture)
