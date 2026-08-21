---
name: redesign
description: Orquesta el workflow completo de rediseño UI/UX en 8 pasos. Coordina audit → DS → Figma → implement → validate → a11y → review.
license: MIT
compatibility: Requiere skills ui-audit, design-system, figma, visual-validation
metadata:
  author: portal-sed
  version: "1.0"
---

# Redesign — Skill (orchestrator)

Trigger: `rediseño`, `redesign`, `rediseñar página`, o workflow `redesign.md`.

## Workflow (8 pasos — orden estricto, no saltar)
1. **ANALYZE** — inventario `web/src/routes`, `components`, `stores`, `app.css`. Preservar lógica/APIs/rutas. Reutilizar componentes.
2. **UI AUDIT** — delega a `ui-audit` skill. Inconsistencias DS primero.
3. **DESIGN SYSTEM PROPOSAL** — delega a `design-system` skill. Tokens Tailwind4+DaisyUI.
4. **FIGMA** — delega a `figma` skill. Llevar propuesta/UI a Figma, leer variables/componentes/estilos.
5. **USER FEEDBACK** — pausar, presentar Figma/proposal, esperar aprobación. No ambiguos → no implementar.
6. **IMPLEMENT** — solo tras aprobación. TOKEN→`app.css`, COMPONENT→`components/**`, PAGE→`routes/**`. ARCHITECTURE requiere aprobación extra.
7. **VISUAL VALIDATION** — delega a `visual-validation` skill (IMPLEMENT→RUN→CAPTURE→COMPARE→FIX→RECHECK).
8. **ACCESSIBILITY CHECK → FINAL REVIEW** — WCAG 2.1 AA + review visual final. Si falla, loop a fix.

## Gates
- No IMPLEMENT sin Figma + aprobación.
- No PAGE sin TOKEN/COMPONENT resuelto.
- ARCHITECTURE → STOP, pedir aprobación.

## Outputs
- Proposal, Figma link, implementación, reporte validación, checklist a11y, review final.

## Prohibited
- Saltar pasos
- Implementar ambiguo
- Modificar `api/` sin aprobación
