# Estilos y UI

## Stack UI

- **DaisyUI** sobre Tailwind (tema `winter` con colores institucionales SED).
- Tipografía y espaciado consistentes; sentence case en títulos y botones.

## Reglas de componentes (MANDATORIO)

- **Todos los componentes deben usar clases de DaisyUI.** No se permite usar elementos HTML nativos sin las clases DaisyUI correspondientes:
  - `button` → `btn` (+ variante: `btn-primary`, `btn-ghost`, `btn-outline`, etc.)
  - `input` → `input` (+ variante: `input-bordered` en v4, nativo en v5)
  - `select` → `select` (+ variante: `select-primary`, `select-ghost`, etc.)
  - `textarea` → `textarea`
  - Las reglas en `app.css` (`@layer components`) aplican fallback automático para elementos nativos sin clase.
- **No se deben usar estilos inline ni clases CSS raw para componentes de formulario o interacción.** Usar siempre las utilidades de DaisyUI.

## Reglas de diseño del proyecto

- Sin box-shadow decorativo.
- Sin borde solo izquierdo/derecho como acento.
- Hovers sutiles, sin elevación fuerte.
- Pocas animaciones; respetar `prefers-reduced-motion`.

## Accesibilidad

- Contraste WCAG AA.
- Labels en inputs; foco visible.
- Tablas con headers semánticos.

## Internacionalización

- Español por defecto; preparar claves i18n si más idiomas en roadmap.

## Skill sugerida

Definir `design-tokens` en OpenSpec: colores primario/secundario, radius, breakpoints antes de maquetar.
