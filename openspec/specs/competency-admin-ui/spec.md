# competency-admin-ui Specification

## Purpose

Pantallas del módulo RH para administrar el marco único de competencias: pilares, competencias por pilar, criterios de escala 1–5 y niveles de aceptación por competencia × perfil de evaluación. UI-first con fixtures JSON, sin API ni persistencia. Refleja las decisiones #1 (categorías sin ponderación) y #2 (catálogo único para todos los perfiles).

## Screens

| Screen | Route | Description |
|--------|-------|-------------|
| Pilares | `/rh/pilares` | Lista, creación, edición y eliminación de pilares |
| Competencias | `/rh/pilares/:id/competencias` | CRUD de competencias agrupadas por pilar |
| Criterios escala | `/rh/criterios-escala` | Tablas por pilar: competencias × niveles con título, múltiples criterios por celda |
| Niveles aceptación | `/rh/niveles-aceptacion` | Definiciones globales de nivel + asignación por competencia × perfil + vista resumen |

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **Pillar** | `id`, `name`, `description` | Agrupa competencias relacionadas |
| **Competency** | `id`, `name`, `description`, `pillarId` | Pertenece a un único pilar |
| **ScaleCriterion** | `id`, `competencyId`, `pillarId`, `level` (1–5), `description` | Admite múltiples criterios por combinación competencyId × pillarId × level |
| **LevelDefinition** | `level` (1–5), `label`, `description` | Global — misma etiqueta y descripción en todos los perfiles |
| **CompetencyAcceptanceLevel** | `competencyId`, `profileId`, `level` (1–5) | Nivel de aceptación por competencia × perfil |
| **AcceptanceLevel** | `profileId`, `level`, `label`, `description` | **Deprecado** — reemplazado por `LevelDefinition` + `CompetencyAcceptanceLevel` |

Perfiles: `colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `rh` (8 perfiles, NO incluye `director-general`).

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
- THEN el pilar se agrega a la tabla con confirmación visual (`alert-success`)

#### Scenario: Eliminar pilar con competencias hijas

- GIVEN pilar con competencias asociadas
- WHEN confirma eliminación tras advertencia
- THEN el pilar, sus competencias y sus criterios de escala se eliminan del store local (cascada)

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
- AND se puede asignar criterios de escala desde la pantalla de criterios

### Requirement: Tablas de criterios de escala por pilar

El sistema SHALL mostrar una tabla independiente por cada pilar. Cada tabla tiene filas = competencias del pilar, columnas = niveles 1–5 con título desde `LevelDefinition` (ej. "N1 - No aceptable"). El encabezado de cada nivel SHALL incluir un icono de estrella (`lucide Star`) y mostrar "NX - Título".

#### Scenario: Visualizar tabla por pilar

- GIVEN usuario en `/rh/criterios-escala`
- WHEN la página carga
- THEN se renderiza una tabla por cada pilar que tenga competencias
- AND cada tabla tiene un badge con el nombre del pilar en color `primary` (o color rotativo: primary, secondary, accent, etc.)
- AND las columnas muestran "N1 - No aceptable", "N2 - En desarrollo", etc. con icono de estrella
- AND cada celda muestra todas las descripciones de criterio para esa competencia × pilar × nivel (múltiples criterios soportados)

#### Scenario: Editar criterios de escala

- GIVEN celda seleccionada en una tabla
- WHEN hace clic en la celda
- THEN se abre `ScaleCriterionModal` mostrando todos los niveles 1–5
- AND puede agregar múltiples criterios por nivel (botón "Agregar criterio")
- AND puede eliminar criterios individuales (botón trash por criterio)
- AND puede editar la descripción de cada criterio

#### Scenario: Editar definiciones de nivel

- GIVEN usuario en `/rh/criterios-escala`
- WHEN hace clic en "Editar definiciones de nivel"
- THEN se abre `LevelDefinitionModal` mostrando los 5 niveles con label + description editables
- AND los cambios se reflejan inmediatamente en los encabezados de columna

### Requirement: Niveles de aceptación por competencia y perfil

El sistema SHALL permitir:
1. Definir etiquetas y descripciones globales para niveles 1–5 (compartidas por todos los perfiles)
2. Asignar nivel de aceptación por competencia × perfil mediante un `CustomSelect`
3. Visualizar resumen completo de niveles por competencia × perfil

#### Scenario: Editar definiciones globales de nivel

- GIVEN usuario en `/rh/niveles-aceptacion`
- WHEN hace clic en "Editar definiciones de nivel"
- THEN se abre `LevelDefinitionModal` con label + description para cada nivel 1–5
- AND los cambios son globales (afectan a todos los perfiles)

#### Scenario: Asignar nivel por competencia y perfil

- GIVEN usuario en `/rh/niveles-aceptacion`
- WHEN selecciona perfil mediante tabs (`tabs-lift`, uno por perfil)
- THEN se muestran todas las competencias agrupadas por pilar
- AND cada competencia tiene un `CustomSelect` para seleccionar nivel (1–5)
- AND las opciones muestran "N1 - No aceptable", etc. desde `LevelDefinition`
- WHEN cambia un nivel usando el `CustomSelect`
- THEN el cambio se marca como pendiente de guardar
- WHEN hace clic en "Guardar cambios"
- THEN los cambios persisten en el store local durante la sesión
- AND se muestra confirmación visual (`alert-success`)

#### Scenario: Vista resumen

- GIVEN usuario en `/rh/niveles-aceptacion`
- WHEN hace clic en "Vista resumen"
- THEN se abre `AcceptanceLevelSummaryModal`
- AND muestra una tabla con competencias como filas y perfiles como columnas
- AND cada celda muestra el nivel asignado en un círculo con número
- AND las competencias se agrupan visualmente por pilar

## UI Components

| Component | Description |
|-----------|-------------|
| **CustomSelect** | Dropdown reutilizable usando `ul > li` con clases `menu-xs`, `menu-active`. Popover API para el menú desplegable. Soporte de teclado (ArrowUp/Down, Enter, Escape, Home/End, role combobox/listbox). |
| **LevelDefinitionModal** | Modal `<dialog>` con inputs para label y textarea para description de cada nivel 1–5. Botones "Guardar cambios" y "Cancelar". |
| **AcceptanceLevelSummaryModal** | Modal `<dialog>` con tabla resumen: filas = competencias agrupadas por pilar, columnas = perfiles (abreviados COL, JEF, VEN, GTE, DIV, REG, DIR, RH), celdas = nivel en círculo. Botón "Cerrar". |
| **ScaleCriteriaMatrix** | Renderiza una tabla por pilar. Badge de color por pilar. Columnas: competencia + niveles con "NX - Título" y estrella. Celdas cliqueables que muestran descripciones múltiples. |
| **ScaleCriterionModal** | Modal con un fieldset por nivel. Soporta múltiples criterios: agregar (Plus), eliminar (Trash2), editar descripción. Submit calcula diff (crear/actualizar/eliminar) sobre el store. |
| **AcceptanceLevelEditor** | Tabs-lift por perfil. Competencias agrupadas por pilar con `CustomSelect` por competencia. Botones: "Editar definiciones de nivel", "Vista resumen", "Guardar cambios". |
| **PillarTable** | Tabla de pilares con acciones editar/eliminar. Modal de confirmación (`ConfirmDeleteModal`) al eliminar. |
| **CompetencyTable** | Tabla de competencias con acciones editar/eliminar. Filtrada por pillarId. |

### State Classes

| State | DaisyUI class | Usage |
|-------|--------------|-------|
| Loading | `skeleton` via `PageSkeleton` | Tablas en carga inicial (timeout 300ms simulado) |
| Empty | `EmptyState` component | Sin datos, con mensaje descriptivo |
| Error | `alert-error` | Fallo al cargar fixtures (no aplica — carga síncrona) |
| Success | `alert-success` | Confirmación CRUD |
| Forbidden | `ForbiddenState` | Perfil sin acceso a ruta RH |

Formularios: `input-bordered`, `textarea`, `select`. Botones: `btn-primary` (crear/guardar), `btn-ghost` (cancelar), `btn-error` (eliminar). Tablas: `table-zebra`. Modales: `dialog` nativo con `modal`, `modal-box`, `modal-action`. Sin estilos inline ni CSS raw. Sentence case en todos los textos visibles.

## Fixtures

Archivos JSON en `web/src/lib/fixtures/competency/`:

| File | Content |
|------|---------|
| `pillars.json` | 3 pilares seed ("Liderazgo", "Técnico", "Comportamental") |
| `competencies.json` | 8 competencias distribuidas en 3 pilares (3 liderazgo, 2 técnico, 3 comportamental) |
| `scale-criteria.json` | 40 entradas con IDs únicos (8 competencias × 5 niveles) |
| `acceptance-levels.json` | 5 definiciones globales de nivel (label + description) |
| `competency-acceptance-levels.json` | 64 entradas (8 competencias × 8 perfiles, nivel default 3) |

Carga síncrona via `import`. Sin lazy loading ni fetch. El store usa `structuredClone` para evitar mutaciones del fixture original.

## Store

`web/src/lib/stores/competencyStore.svelte.ts` — store reactivo Svelte 5 (runas `$state`).

### Getters

| Function | Returns |
|----------|---------|
| `getPillars()` | `Pillar[]` |
| `getCompetencies()` | `Competency[]` |
| `getCompetenciesByPillar(pillarId)` | `Competency[]` filtrado |
| `getScaleCriteria()` | `ScaleCriterion[]` |
| `getScaleCriteriaForCell(competencyId, pillarId)` | `ScaleCriterion[]` filtrado |
| `getLevelDefinitions()` | `LevelDefinition[]` |
| `getLevelDefinition(level)` | `LevelDefinition \| undefined` |
| `getCompetencyAcceptanceLevels()` | `CompetencyAcceptanceLevel[]` |
| `getCompetencyAcceptanceLevelsByProfile(profileId)` | `CompetencyAcceptanceLevel[]` filtrado |
| `getCompetencyAcceptanceLevel(competencyId, profileId)` | `CompetencyAcceptanceLevel \| undefined` |

### Mutations

| Function | Side effects |
|----------|-------------|
| `addPillar(pillar)` | Agrega pilar |
| `updatePillar(id, updates)` | Edita pilar |
| `deletePillar(id)` | Elimina pilar + competencias hijas + criterios de escala del pilar (cascada) |
| `addCompetency(competency)` | Agrega competencia |
| `updateCompetency(id, updates)` | Edita competencia |
| `deleteCompetency(id)` | Elimina competencia + criterios de escala asociados (cascada) |
| `updateScaleCriterion(id, description)` | Edita descripción de criterio |
| `addScaleCriterion(criterion)` | Agrega criterio con ID generado |
| `removeScaleCriterion(id)` | Elimina criterio por ID |
| `updateLevelDefinition(level, label, description)` | Actualiza definición global de nivel |
| `setCompetencyAcceptanceLevel(competencyId, profileId, level)` | Asigna nivel por competencia × perfil (crea o actualiza) |
| `setCompetencyAcceptanceLevelsForProfile(profileId, assignments)` | Reemplaza todas las asignaciones de un perfil |

## Validations

| Rule | Enforcement |
|------|-------------|
| Nombre de pilar único en catálogo | Al crear/editar |
| Nombre de competencia único por pilar | Al crear/editar |
| Campos nombre y descripción requeridos | `required` en formularios |
| Sin campo de peso en categorías (decisión #1) | No renderizar input de peso |
| Catálogo idéntico desde cualquier perfil (decisión #2) | Fixtures únicos, sin filtro por perfil activo |
| Nivel 1–5 en asignaciones | Tipo estricto `1 \| 2 \| 3 \| 4 \| 5` en TypeScript |

## Non-functional

- **Accesibilidad**: WCAG 2.1 AA — contraste mínimo, foco visible (`focus-visible:ring`), `aria-label` en botones de acción, `<th>` semánticos, `role` combobox/listbox/option en CustomSelect, `role="dialog"` y `aria-modal` en modales.
- **Responsive**: Tablas con scroll horizontal en viewports < `md`; formularios en columna en móvil.
- **Menú**: Rutas visibles solo con perfil `rh` activo (hereda filtro de `ui-shell` A1).
- **Rendimiento**: Lazy routes (code-splitting por pantalla). Sin llamadas de red. Carga síncrona de fixtures.
- **Estilo**: Sin box-shadow decorativo, sin border-left/right como acento. Respeto a `prefers-reduced-motion`.
- **Store**: Svelte 5 runas (`$state`) con `structuredClone` en inicialización. Mutaciones inmutables (nuevo array en cada cambio).
