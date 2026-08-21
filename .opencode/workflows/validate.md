---
name: validate
description: Validación visual y accesibilidad independiente — ciclo IMPLEMENT→RECHECK + a11y. Extrae pasos 7-8 de redesign.
---

# Workflow: Validate (ciclo IMPLEMENT → RECHECK + a11y)

> Secuencia exacta: IMPLEMENT → RUN APPLICATION → CAPTURE/INSPECT VISUAL → COMPARE WITH FIGMA → IDENTIFY DIFFERENCES → FIX → RECHECK → ACCESSIBILITY CHECK → FINAL REVIEW

Trigger: `/validate`, `validar`, `visual check`, o tras IMPLEMENT de cualquier cambio UI.

Precondición: código ya implementado (TOKEN en `app.css` / COMPONENT en `components/**` / PAGE en `routes/**`). Figma aprobado es fuente de verdad si existe; sin Figma usa proposal DS + screenshots locales.

## 1 — IMPLEMENT (precondición)
- Verifica que el diff está aplicado y commiteable (no requiere Figma en este paso).
- Orden ya respetado: TOKEN → COMPONENT → PAGE. No validar PAGE si TOKEN/COMPONENT tiene diffs críticos.

## 2 — RUN APPLICATION
- `pnpm --dir web dev` (o `pnpm dev` dentro de `web/`).
- Si hay build check: `pnpm --dir web build`.
- Confirma que la ruta/página objetivo carga sin errores de consola.

## 3 — CAPTURE / INSPECT VISUAL
- Delega a `visual-validation` skill.
- Con Figma MCP: `figma_get_design_context` + `figma_get_screenshot` del nodo aprobado + screenshot app (`figma_get_screenshot` / `maestro_take_screenshot` / browser capture).
- Sin Figma MCP: captura local + `inspect_screen` si aplica. Usa DS proposal como referencia.
- Guarda capturas para el reporte.

## 4 — COMPARE WITH FIGMA
- Overlay/diff contra Figma (o proposal). Revisa: spacing, color (sed/sed-dark), radius, shadow, typography (Geist/Binjay), breakpoints. Tolerancia ≤ 2px.
- Aplica reglas `visual-quality` (flat by default, shadows solo via `soft-shadow`/`card-shadow`/`card-elevated`) y `tailwind-daisyui` (tokens en `app.css`, sin hardcode hex).

## 5 — IDENTIFY DIFFERENCES
- Lista cada diff: `archivo:línea | esperado (Figma) | actual | tipo TOKEN/COMPONENT | severidad`.
- Agrupa por TOKEN primero, luego COMPONENT. PAGE solo si los anteriores están clean.

## 6 — FIX
- Patch mínimo. TOKEN primero (`web/src/app.css` theme vars / `@utility`).
- Luego COMPONENT (`web/src/lib/components/**`). No refactor extra, no abstracciones prematuras.
- Si el fix toca ARCHITECTURE → STOP, pide aprobación antes de recheck.

## 7 — RECHECK
- Re-ejecuta 2 → 3 → 4. Repite 5 → 6 → 7 hasta 0 diffs críticos.
- Loop determinista: cada iteración documenta captura + diffs restantes.

## 8 — ACCESSIBILITY CHECK
- Delega checks de `accessibility` rule (WCAG 2.1 AA):
  - Contraste ≥ 4.5:1 (texto) / 3:1 (large) en `sed` y `sed-dark`.
  - Keyboard: Tab, Enter/Space, `:focus-visible` visible (prohibido `outline:none` sin reemplazo).
  - Semántica: `button`/`a`/`label↔input`, `alt`/`aria-hidden`, `aria-invalid`+`aria-describedby`.
  - Responsive: 320px sin scroll horizontal, touch targets ≥ 44×44px, verifica 320/768/1280.
- Si falla a11y → vuelve a 6 (FIX) → 7 (RECHECK) → 8.

## 9 — FINAL REVIEW
- Review visual final + a11y checklist firmado. Si pasa → marca workflow como pass.
- Outputs: `visual-validation-report.md` (pass/fail, diffs, screenshots, fix) + `a11y-checklist.md` + `final-review.md` si aplica.

## Outputs
- `visual-validation-report.md` — capturas, diffs, fix aplicado, estado RECHECK
- `a11y-checklist.md` — WCAG 2.1 AA (ambos themes)
- `final-review.md` — aprobación final (opcional si solo validate)

## Gates
- No aprobar con diffs críticos. No saltar RECHECK. Sin RECHECK no hay pass.
- Sin Figma, no bloquear: usa proposal DS como fuente de verdad (documenta fallback).

## Rules linked
- `.opencode/skills/visual-validation/SKILL.md` — ciclo capture→compare→fix→recheck
- `.opencode/rules/accessibility.md` — WCAG 2.1 AA, keyboard, contrast, responsive
- `.opencode/rules/visual-quality.md` — flat default, shadows/borders, typography, diff ≤2px
- `.opencode/rules/tailwind-daisyui.md` — Tailwind 4 via @tailwindcss/vite, theme sed/sed-dark en app.css

## Prohibited
- Aprobar con diffs críticos o a11y fail
- Fix sin RECHECK posterior
- Crear `tailwind.config.*` o hardcodear hex/box-shadow fuera de `app.css`
