---
name: ui-audit
description: Audita UI/UX existente a nivel Design System antes de tocar páginas. Detecta inconsistencias de tokens y componentes.
license: MIT
compatibility: Svelte 5 + Tailwind 4 + DaisyUI
metadata:
  author: portal-sed
  version: "1.0"
---

# UI Audit — Skill

Trigger: `/audit`, `auditar ui`, `ui audit`, o paso UI AUDIT del workflow redesign.

## Input
- Scope: ruta o `web/src/**` completo. Default completo.

## Steps (determinista, solo lectura)
1. **Inventario**: lista `web/src/app.css` theme vars + `web/src/lib/components/**` + `web/src/routes/**` (glob, no edites).
2. **Tokens**: busca hardcodes fuera de `app.css` → `grep` hex/rgb/`box-shadow`/`border-.*px`/`style=` en `.svelte`.
3. **Componentes**: compara cada componente vs DaisyUI. Marca duplicados/inconsistentes.
4. **Inconsistencias DS primero**: agrupa por TOKEN / COMPONENT (no page).
5. **A11y quick**: contrasta `sed`/`sed-dark`, revisa focus/keyboard/labels básicos.
6. **Output**: reporte markdown (tabla: archivo | hallazgo | tipo TOKEN/COMPONENT | severidad).

## Outputs
- `audit-report.md` (o stdout) con: resumen, tabla hallazgos, top 5 riesgos DS, next: Design System Proposal.

## Checks
- [ ] No se modificó `web/` ni `api/`
- [ ] Hallazgos etiquetados TOKEN/COMPONENT/PAGE
- [ ] Hardcodes listados con archivo:línea

## Prohibited
- Proponer páginas antes de DS.
- Crear `tailwind.config.*` o tocar `app.css` en este paso.
