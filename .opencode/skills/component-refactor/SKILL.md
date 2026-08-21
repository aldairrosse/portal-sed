---
name: component-refactor
description: Refactoriza componente Svelte existente hacia DS tokens y DaisyUI, preservando API y comportamiento.
license: MIT
compatibility: Svelte 5 runes, Tailwind 4, DaisyUI
metadata:
  author: portal-sed
  version: "1.0"
---

# Component Refactor — Skill

Trigger: `refactor componente`, `migrar a ds`, `daisyui`, componente inconsistente.

## Input
- Path componente: `web/src/lib/components/**.svelte`

## Steps
1. **Leer**: componente + `app.css` tokens + DaisyUI equivalente.
2. **Mapear**: props/API actuales → mantener 100%. No romper consumidores (`grep` referencias).
3. **Reemplazar**: hardcodes → tokens (`var(--color-*)` / clase Tailwind / `@utility`). DaisyUI class antes de custom.
4. **Verificar**: responsive + a11y (focus, label, contrast) + no `box-shadow`/`border` decorativo fuera de token.
5. **Output**: diff mínimo, etiquetado COMPONENT (o TOKEN si tocó `app.css`).

## Checks
- [ ] Props/API sin breaking change
- [ ] No hardcode restante
- [ ] Usa DaisyUI si existe
- [ ] Referencias verificadas

## Prohibited
- Cambiar lógica/estado/ruta
- Crear nuevo componente si reusar alcanza
- Añadir React/lib nueva
