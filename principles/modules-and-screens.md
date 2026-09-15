# Módulos y pantallas

## Mapa (requerimientos SED)

| Módulo | Pantallas principales (borrador) |
|--------|----------------------------------|
| Catálogo de objetivos | Lista, alta/edición, filtros por categoría |
| Mis evaluados | Lista empleados, detalle desempeño/puesto, fijación, evaluación por tipo, envío notificación |
| Mi evaluación | Vista lectura (y edición si ciclo lo permite) |
| Objetivos | Wizard o formulario fijación; formulario evaluación; metas por desempeño |

## Navegación

- Layout autenticado con menú lateral por rol.
- Rutas lazy-loaded por módulo.

## Estados de UI

- Loading skeleton (DaisyUI `skeleton`).
- Vacío, error de red, sin permiso (403).

## Por spec OpenSpec (uno por módulo)

1. `catalog-objectives`
2. `my-evaluatees`
3. `my-evaluation`
4. `objectives-fixation-evaluation`

Cada spec: actores, precondiciones, flujo feliz, errores, notificaciones email.
