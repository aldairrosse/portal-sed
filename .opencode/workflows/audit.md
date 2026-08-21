---
name: audit
description: Ejecuta UI Audit a nivel Design System. Solo lectura, no modifica web/api.
---

# Workflow: Audit

Trigger: `/audit` o `ui-audit` skill.

## Steps

1. **Inventario** — lista `web/src/app.css` (theme `sed`/`sed-dark`), `web/src/lib/components/**`, `web/src/routes/**`.
2. **Ejecuta ui-audit skill** — hardcodes, duplicados DaisyUI, inconsistencias TOKEN/COMPONENT.
3. **Reporte** — genera `audit-report.md` con tabla hallazgos + top 5 riesgos DS.
4. **Gate** — si hay hallazgos TOKEN/COMPONENT → siguiente es `design-system` proposal. No pasar a PAGE.

## Outputs
- `audit-report.md`

## Rules linked
- `design-system.md`, `tailwind-daisyui.md`, `visual-quality.md`, `accessibility.md`

## Prohibited
- Modificar `web/` / `api/`
- Crear `tailwind.config.*`
