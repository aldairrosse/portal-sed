# competency-framework Specification

## Purpose

Define el **marco de competencias** de la empresa: pilares únicos para todos los perfiles, competencias por pilar, escala 1–5 con criterios detallados por competencia × pilar, y niveles de aceptación por competencia × perfil de evaluación. Esta spec es la fuente de verdad para el catálogo de competencias que alimenta las pantallas A2 (RH admin) y A3 (asignación anual), y el backend C3.

**Decisiones reflejadas:** #1 (categorías de metas sin ponderación — las categorías de competencias/pilares no ponderan), #2 (catálogo único de pilares y competencias para todos los perfiles).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **Pillar** | `id`, `name`, `description` | Agrupa competencias relacionadas. Catálogo único para toda la empresa (decisión #2). Sin campo de ponderación (decisión #1). |
| **Competency** | `id`, `name`, `description`, `pillarId` | Competencia perteneciente a un único pilar. |
| **ScaleCriterion** | `id`, `competencyId`, `pillarId`, `level` (1–5), `description` | Criterio descriptivo para una competencia en un nivel dado. Admite múltiples criterios por combinación (competencyId × pillarId × level). |
| **LevelDefinition** | `level` (1–5), `label`, `description` | Definiciones globales de nivel. Etiquetas como "No aceptable", "En desarrollo", etc. Compartidas por todos los perfiles y competencias. |
| **CompetencyAcceptanceLevel** | `competencyId`, `profileId`, `level` (1–5) | Nivel mínimo de aceptación por competencia × perfil de evaluación. Define el umbral para considerar "aceptable" un empleado en esa competencia. |
| **EvaluationProfile** | `id`, `name`, `description` | Perfil de evaluación: `colaborador`, `jefe`, `vendedor`, `gerente-tienda`, `divisional`, `regional`, `director`, `rh` (8 perfiles, no incluye `director-general`). |

### Relaciones

```
Pillar 1:N Competency
Competency 1:N ScaleCriterion
LevelDefinition (catálogo plano, 5 registros)
Competency × EvaluationProfile → CompetencyAcceptanceLevel
```

## Requirements

### Requirement: Catálogo único de pilares (decisión #2)

El sistema SHALL mantener un único conjunto de pilares y competencias visible desde cualquier perfil de evaluación. No SHALL existir pilares duplicados o filtrados por rol. RH administra el catálogo; los demás perfiles solo lo consultan.

#### Scenario: Mismo catálogo desde cualquier perfil

- GIVEN pilares "Liderazgo", "Técnico", "Comportamental" en el sistema
- WHEN perfil `colaborador` consulta pilares
- THEN ve los mismos 3 pilares que vería `rh`
- AND las competencias dentro de cada pilar son idénticas

#### Scenario: RH administra pilares

- GIVEN perfil `rh` activo
- WHEN crea, edita o elimina un pilar
- THEN el cambio es visible para todos los perfiles
- AND no existe mecanismo de "pilar privado" o "pilar por rol"

### Requirement: Competencias por pilar

Cada pilar SHALL contener 1..N competencias. Las competencias son unitarias (no se comparten entre pilares).

#### Scenario: Crear competencia en pilar

- GIVEN pilar "Liderazgo" con 2 competencias existentes
- WHEN se agrega "Comunicación efectiva" al pilar
- THEN la competencia aparece en el pilar con su descripción
- AND queda disponible para asignación de criterios de escala y niveles de aceptación

#### Scenario: Eliminar competencia con criterios

- GIVEN competencia "Comunicación efectiva" con 5 ScaleCriterion asociados
- WHEN se elimina la competencia
- THEN los ScaleCriterion se eliminan en cascada
- AND los CompetencyAcceptanceLevel asociados también se eliminan

### Requirement: Escala 1–5 con criterios detallados

El sistema SHALL definir una escala numérica del 1 al 5 para evaluar competencias. Cada nivel SHALL tener una definición global (label + description) y cada competencia SHALL poder tener múltiples criterios descriptivos por nivel.

#### Scenario: Definiciones de nivel globales

- GIVEN 5 niveles definidos: N1 "No aceptable", N2 "En desarrollo", N3 "Cumple", N4 "Supera", N5 "Excepcional"
- WHEN cualquier usuario consulta la escala
- THEN ve las mismas 5 definiciones con sus descripciones
- AND las definiciones son editables solo por RH

#### Scenario: Múltiples criterios por celda

- GIVEN competencia "Liderazgo de equipo" en pilar "Liderazgo"
- WHEN RH define criterios para nivel 3
- THEN puede agregar 1..N criterios descriptivos para esa celda (competencia × nivel)
- AND cada criterio tiene su propia descripción

### Requirement: Niveles de aceptación por competencia × perfil

El sistema SHALL definir un nivel de aceptación (1–5) para cada combinación de competencia × perfil de evaluación. El nivel de aceptación representa el mínimo esperado para un empleado con ese perfil en esa competencia.

#### Scenario: Asignar nivel de aceptación

- GIVEN competencia "Comunicación efectiva" y perfil `vendedor`
- WHEN RH asigna nivel de aceptación 3 para esa combinación
- THEN un empleado con perfil `vendedor` evaluado en "Comunicación efectiva" necesita alcanzar al menos nivel 3 para estar "aceptado"

#### Scenario: Nivel de aceptación varía por perfil

- GIVEN competencia "Liderazgo de equipo"
- WHEN se comparan niveles de aceptación entre perfiles
- THEN `colaborador` puede tener nivel 2, `jefe` nivel 4, y `director` nivel 5
- AND los niveles son independientes entre perfiles

#### Scenario: Nivel de aceptación varía por competencia

- GIVEN perfil `rh`
- WHEN se comparan niveles de aceptación entre competencias
- THEN cada competencia puede tener un nivel distinto para el mismo perfil
- AND los niveles no se heredan entre competencias

### Requirement: Separación categorías de metas vs categorías de competencias

Las **categorías de metas** (definidas por cada empleado, decisión #5) son **independientes** de los **pilares/categorías de competencias** (definidos por la empresa, decisión #2). No SHALL existir vínculo directo entre una categoría de metas y un pilar de competencias.

#### Scenario: Categorías de metas no afectan competencias

- GIVEN empleado con 3 categorías de metas custom
- WHEN se consulta el marco de competencias
- THEN las categorías de metas no aparecen en la vista de competencias
- AND los pilares de competencias no se modifican por las categorías de metas

## Non-goals

- **Asignación masiva de competencias a empleados**: esta spec define el catálogo; la asignación a empleados específicos es scope de A3 y C3.
- **Importación desde Excel**: no se soporta carga masiva de competencias o pilares.
- **Ponderación de pilares**: los pilares no ponderan (decisión #1); no SHALL existir campo de peso en pilares.
- **Escala fuera de 1–5**: la escala es fija; no se soportan escalas personalizadas por empresa.
- **Competencias transversales**: no se implementa mecanismo de competencias que apliquen a múltiples pilares simultáneamente.
- **Historial de cambios**: no se registra audit log de cambios al catálogo de competencias (scope de C3/C7).
