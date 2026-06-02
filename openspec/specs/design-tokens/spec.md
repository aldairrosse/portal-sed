# design-tokens Specification

## Purpose
TBD - created by archiving change ui-shell-and-design-tokens. Update Purpose after archive.
## Requirements
### Requirement: Tema DaisyUI corporativo SED

El sistema SHALL aplicar un tema DaisyUI personalizado con colores primario, secundario y acento definidos como tokens CSS reutilizables.

#### Scenario: Componentes DaisyUI usan tema

- **WHEN** se renderiza un botón `btn-primary` o card DaisyUI
- **THEN** usa los colores del tema SED, no los valores por defecto de DaisyUI

### Requirement: Sin sombras decorativas

Los componentes del shell y páginas placeholder SHALL NOT usar box-shadow decorativo ni bordes únicamente en lado izquierdo o derecho como acento visual.

#### Scenario: Revisión de layout

- **WHEN** se inspecciona AppShell y sidebar
- **THEN** no hay clases Tailwind `shadow-*` decorativas ni `border-l`/`border-r` como acento de diseño

### Requirement: Sentence case en UI

Textos visibles de títulos, subtítulos, botones e inputs SHALL usar sentence case (primera letra mayúscula, resto según español).

#### Scenario: Menú lateral

- **WHEN** se listan ítems de navegación
- **THEN** las etiquetas siguen sentence case (ej. "Mi evaluación", no "MI EVALUACIÓN")

### Requirement: Tipografía y radius consistentes

El sistema SHALL definir tokens de fuente base y radio de borde (`radius-box`) aplicados globalmente en `app.css` o configuración Tailwind.

#### Scenario: Consistencia visual

- **WHEN** se comparan card, input y botón en la misma vista
- **THEN** comparten la misma familia tipográfica base y radio de borde definido en tokens

### Requirement: Accesibilidad WCAG AA en shell

El shell SHALL cumplir contraste mínimo WCAG 2.1 AA en texto primario sobre fondos del tema y SHALL mostrar indicador de foco visible en controles interactivos del menú y toolbar dev.

#### Scenario: Navegación por teclado

- **WHEN** el usuario tabula por ítems del menú lateral
- **THEN** cada ítem enfocado muestra outline o ring visible

### Requirement: Respeto a prefers-reduced-motion

Animaciones del drawer móvil y transiciones SHALL reducirse o desactivarse cuando el usuario tiene `prefers-reduced-motion: reduce`.

#### Scenario: Usuario con movimiento reducido

- **WHEN** el media query `prefers-reduced-motion: reduce` está activo
- **THEN** la apertura del menú móvil no usa animaciones prolongadas

