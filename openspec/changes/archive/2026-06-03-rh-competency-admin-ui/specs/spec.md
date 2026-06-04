# competency-admin-ui Specification

## Purpose

Pantallas del módulo RH para administrar el marco único de competencias: pilares, competencias por pilar, criterios de escala 1-5 y niveles de aceptación por perfil de evaluación. UI-first con fixtures JSON, sin API ni persistencia. Refleja las decisiones #1 (categorías sin ponderación) y #2 (catálogo único para todos los perfiles).

## Screens

| Screen | Route | Description |
|--------|-------|-------------|
| Pilares | `/rh/pilares` | Lista, creación, edición y eliminación de pilares/categorías |
| Competencias | `/rh/pilares/:id/competencias` | CRUD de competencias agrupadas por pilar |
| Criterios escala | `/rh/criterios-escala` | Matriz editable: criterios 1-5 por competencia y pilar |
| Niveles aceptación | `/rh/niveles-aceptacion` | Definición de niveles de aceptación por perfil (8 perfiles) |

## Data Model

| Entity | Fields |
|--------|--------|
| **Pillar** | `id`, `name`, `description` |
| **Competency** | `id`, `name`, `description`, `pillarId` |
| **ScaleCriterion** | `competencyId`, `pillarId`, `level` (1-5), `description` |
| **AcceptanceLevel** | `profileId`, `level` (1-5), `label`, `description` |

Perfiles: `colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `director-general`, `rh`.

## Requirements

### Requirement: Gestión de pilares

El sistema SHALL permitir listar, crear, editar y eliminar pilares. No se debe mostrar campo de ponderación (decisión #1).

#### Scenario: Lista de pilares

- GIVEN perfil RH activo
- WHEN navega a `/rh/pilares`
- THEN ve tabla con nombre, descripción y acciones (editar, eliminar)
- AND no existe columna ni campo de peso/ponderación

#### Scenario: Crear pilar

- GIVEN lista de pilares visible
- WHEN hace clic en "Nuevo pilar", completa nombre y descripción, y confirma
- THEN el pilar se agrega a la tabla con confirmación visual (alert-success)

#### Scenario: Eliminar pilar con competencias hijas

- GIVEN pilar con competencias asociadas
- WHEN confirma eliminación tras advertencia
- THEN el pilar y sus competencias se eliminan del store local

### Requirement: Gestión de competencias por pilar

El sistema SHALL permitir CRUD de competencias dentro de cada pilar. El catálogo SHALL ser idéntico desde cualquier perfil activo (decisión #2).

#### Scenario: Lista de competencias

- GIVEN usuario en `/rh/pilares/:id/competencias`
- WHEN la página carga
- THEN muestra tabla de competencias del pilar seleccionado

#### Scenario: Crear competencia

- GIVEN vista de competencias de un pilar
- WHEN completa nombre y descripción y confirma
- THEN aparece en la tabla agrupada bajo ese pilar

### Requirement: Matriz de criterios de escala

El sistema SHALL mostrar una matriz donde filas = competencias, columnas = pilares, celdas = descripciones para niveles 1 al 5.

#### Scenario: Visualizar matriz

- GIVEN usuario en `/rh/criterios-escala`
- WHEN la página carga
- THEN cada celda muestra resumen de criterios y permite editar modal con los 5 niveles

#### Scenario: Editar criterio

- GIVEN celda seleccionada en la matriz
- WHEN edita descripción del nivel 3
- THEN el cambio se refleja inmediatamente en la celda

### Requirement: Niveles de aceptación por perfil

El sistema SHALL permitir definir etiquetas y descripciones para cada nivel 1-5 por perfil de evaluación.

#### Scenario: Editar niveles por perfil

- GIVEN usuario en `/rh/niveles-aceptacion`
- WHEN selecciona perfil y edita etiqueta del nivel 2
- THEN los cambios persisten en el store local durante la sesión

## UI Components

| State | DaisyUI class | Usage |
|-------|--------------|-------|
| Loading | `skeleton` | Tablas en carga inicial |
| Empty | `alert` + `btn-primary` | Sin datos, con CTA para crear |
| Error | `alert-error` | Fallo al cargar fixtures |
| Success | `alert-success` | Confirmación CRUD |

Formularios: `input-bordered`, `textarea`, `select`. Botones: `btn-primary` (crear/guardar), `btn-ghost` (cancelar), `btn-error` (eliminar). Tablas: `table-zebra`. Sin estilos inline ni CSS raw. Sentence case en todos los textos visibles.

## Fixtures

Archivos JSON en `web/src/lib/fixtures/competency/`:

| File | Content |
|------|---------|
| `pillars.json` | 3-4 pilares (ej. "Liderazgo", "Técnico", "Comportamental") |
| `competencies.json` | 2-3 competencias por pilar |
| `scale-criteria.json` | Matriz de criterios nivel 1-5 por competencia × pilar |
| `acceptance-levels.json` | Niveles de aceptación por cada uno de los 8 perfiles |

Carga síncrona via `import`. Sin lazy loading ni fetch.

## Validations

| Rule | Enforcement |
|------|-------------|
| Nombre de pilar único en catálogo | Al crear/editar |
| Nombre de competencia único por pilar | Al crear/editar |
| Campos nombre y descripción requeridos | `required` en formularios |
| Sin campo de peso en categorías (decisión #1) | No renderizar input de peso |
| Catálogo idéntico desde cualquier perfil (decisión #2) | Fixtures únicos, sin filtro por perfil activo |

## Non-functional

- **Accesibilidad**: WCAG 2.1 AA — contraste mínimo, foco visible (`focus:ring`), `aria-label` en botones de acción, `<th>` semánticos.
- **Responsive**: Tablas con scroll horizontal en viewports < `md`; formularios en columna en móvil.
- **Menú**: Rutas visibles solo con perfil `rh` activo (hereda filtro de `ui-shell` A1).
- **Rendimiento**: Lazy routes (code-splitting por pantalla). Sin llamadas de red.
- **Estilo**: Sin box-shadow decorativo, sin border-left/right como acento. Respeto a `prefers-reduced-motion`.
