---
name: visual-validation
description: Valida visualmente implementación contra Figma. Ciclo capture → compare → fix → recheck.
license: MIT
compatibility: Figma MCP si disponible, fallback screenshot/inspect
metadata:
  author: portal-sed
  version: "1.0"
---

# Visual Validation — Skill

Trigger: `validar visual`, `visual check`, o paso VISUAL VALIDATION.

## Cycle (loop hasta pass)
1. **IMPLEMENT** — código ya aplicado (TOKEN/COMPONENT/PAGE).
2. **RUN** — `pnpm dev` (web) + `pnpm build` si aplica.
3. **CAPTURE/INSPECT** — screenshot (`maestro_take_screenshot` / `figma_get_screenshot` / browser) + `inspect_screen` si mobile. Si Figma MCP disponible: `figma_get_design_context` + `figma_get_screenshot` del nodo aprobado.
4. **COMPARE WITH FIGMA** — overlay/diff: spacing, color, radius, shadow, typography, breakpoints. Tolerancia ≤2px.
5. **IDENTIFY DIFFS** — lista diffs: archivo:línea | esperado Figma | actual | tipo TOKEN/COMPONENT.
6. **FIX** — patch mínimo (token en `app.css` primero).
7. **RECHECK** — re-captura, comparar. Repetir hasta 0 diffs críticos.

## Outputs
- `visual-validation-report.md`: pass/fail, diffs, screenshots, fix aplicado.

## Fallback (sin Figma MCP)
- Usa screenshots locales + checklist DS proposal como fuente verdad.

## Checks
- [ ] Diffs documentados
- [ ] Fix mínimo, sin refactor extra
- [ ] Recheck confirma pass

## Prohibited
- Aprobar con diffs críticos
- Implementar sin recheck
