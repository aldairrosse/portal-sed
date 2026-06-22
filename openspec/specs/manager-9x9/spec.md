# manager-9x9 Specification

## Purpose

Define la **matriz 3×3 de desempeño vs potencial** (tiers 1–3 automáticos): ejes de medición derivados de datos fuente (metas, competencias), cuadrantes configurables por RH, y la separación explícita de la evaluación RH. Esta spec es la fuente de verdad para la matriz de potencial que alimenta la pantalla A6 (manager-9x9-ui) y el backend C6.

**Decisiones reflejadas:** #4 (fin de año: el sistema calcula automáticamente los tiers desde avance de metas + competencias; NO sustituye la evaluación RH de competencias).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **NineBoxMatrix** | `id`, `cycleId`, `evaluatorId`, `phaseId` | Instancia de la matriz para un evaluador en un ciclo por etapa. Unique (cycle, evaluator, phase). |
| **NineBoxEntry** | `id`, `matrixId`, `evaluateeId`, `performanceTier` (1–3), `potentialTier` (1–3), `quadrant` (calculado), `comments?` | Ubicación automática de un evaluado. Los tiers se derivan de datos fuente (metas, competencias). |
| **NineBoxQuadrant** | `id`, `label`, `title`, `description`, `colorHex`, `color` (legacy), `actionRecommendation` | Definición de cuadrante. title, description, colorHex editables por RH. |
| **NineBoxScale** | `axis` (`performance` \| `potential`), `level` (1–9), `label`, `description` | Definiciones de los 9 niveles por eje. Se conserva para compatibilidad histórica. Ya no se usa en la matriz activa. |

### Matriz 3×3 — Layout

```
Potencial (tier)
  3 │ 7  │ 8  │ 9
  2 │ 4  │ 5  │ 6
  1 │ 1  │ 2  │ 3
    └────┴────┴────┘
      1    2    3
   Desempeño (tier)
```

Quadrant formula: `quadrant = (potentialTier − 1) × 3 + performanceTier`

### Cuadrantes definidos

| Cuadrante | Tiers (Desempeño × Potencial) | Label | Descripción | Acción recomendada |
|-----------|-------------------------------|-------|-------------|-------------------|
| 1 | Perf Bajo (1) × Pot Bajo (1) | Bajo desempeño, bajo potencial | Requiere acción correctiva inmediata | Plan de mejora o reasignación |
| 2 | Perf Medio (2) × Pot Bajo (1) | Desempeño medio, bajo potencial | Estable pero sin potencial de crecimiento | Mantener, desarrollo de habilidades |
| 3 | Perf Alto (3) × Pot Bajo (1) | Alto desempeño, bajo potencial | Excelente rendimiento, estancado | Recompensar, evitar sobrecarga |
| 4 | Perf Bajo (1) × Pot Medio (2) | Bajo desempeño, potencial medio | Potencial sin materializar | Coaching, asignar mentores |
| 5 | Perf Medio (2) × Pot Medio (2) | Desempeño medio, potencial medio | Promedio, crecimiento gradual | Desarrollo planificado |
| 6 | Perf Alto (3) × Pot Medio (2) | Alto desempeño, potencial medio | Sólido, listo para más responsabilidad | Preparar para rol superior |
| 7 | Perf Bajo (1) × Pot Alto (3) | Bajo desempeño, alto potencial | Talento desaprovechado | Investigar causas, reasignar si necesario |
| 8 | Perf Medio (2) × Pot Alto (3) | Desempeño medio, alto potencial | Listo para aceleración | Reto, proyecto de alto impacto |
| 9 | Perf Alto (3) × Pot Alto (3) | Alto desempeño, alto potencial | Estrella, sucesor natural | Promover, plan de sucesión |

## Requirements

### Requirement: Calificación de desempeño y potencial (decisión #4)

El sistema SHALL calcular automáticamente el performance tier (1–3) desde el avance de metas del evaluatee y el potential tier (1–3) desde el promedio de autoevaluación + evaluación RH de competencias. Los tiers son de solo lectura para el jefe. La calificación ya no es manual con escala 1–9.

(Previously: El jefe calificaba manualmente desempeño y potencial en escala 1–9 con sliders.)

#### Scenario: Tiers calculados automáticamente

- GIVEN evaluatee con avance de metas 60% y promedio competencias 4.0
- WHEN jefe abre la matriz 3×3
- THEN performance tier = 2, potential tier = 3
- AND los tiers NO son editables por el jefe

#### Scenario: Cálculo es independiente de la intervención del jefe

- GIVEN evaluatee con datos de metas y competencias completos
- WHEN se recalcula la matriz al cerrar etapa
- THEN los tiers se derivan exclusivamente de datos existentes (metas, autoevaluación, ev. RH)
- AND el jefe no puede modificar los tiers manualmente

### Requirement: Cálculo automático de cuadrante

El sistema SHALL calcular el cuadrante automáticamente como `(potentialTier − 1) × 3 + performanceTier`. El cuadrante se recalcula al cambiar los tiers (por cambio en datos fuente).

(Previously: El cuadrante se calculaba desde scores manuales 1–9 de desempeño y potencial.)

#### Scenario: Cuadrante recalculado al actualizar datos fuente

- GIVEN evaluatee con performance tier 2, potential tier 3 → cuadrante 8
- WHEN se registra nuevo avance de metas que eleva performance tier a 3
- THEN el cuadrante se recalcula a 9

#### Scenario: Cambio en evaluación de competencias

- GIVEN evaluatee con potential tier 2
- WHEN RH completa evaluación de competencias y avgPotential sube a 4.2 → tier 3
- THEN el cuadrante se recalcula inmediatamente al recomputar la matriz

### Requirement: Vista de matriz por evaluador

El sistema SHALL renderizar la matriz como grilla 3×3 con 9 cuadrantes coloreados según configuración de `NineBoxQuadrant`. Cada empleado se posiciona según sus tiers calculados. Los cuadrantes son editables por RH (title, description, colorHex).

(Previously: Grilla 9×9 con scores manuales 1–9 y cuadrantes con colores DaisyUI fijos.)

Cada jefe/director/gerente/director-general SHALL ver su propia matriz. El scope de evaluatees varía por perfil: jefe ve reportes directos; director ve todos los managers bajo su jerarquía; director-general ve toda la organización. La matriz de un evaluador no es visible para otros evaluadores del mismo nivel.

#### Scenario: Grilla 3×3 con empleados automáticos

- GIVEN matriz de Carlos con 5 evaluatees en fase `cierre`
- WHEN se renderiza
- THEN 9 celdas coloreadas con colorHex configurado por RH
- AND cada evaluatee aparece como dot en su celda correspondiente
- AND clic en dot abre card de solo lectura con tiers y cuadrante

#### Scenario: Jefe ve su matriz (sin cambios)

- GIVEN jefe "Carlos" con 5 evaluatees directos
- WHEN accede a la matriz
- THEN ve los 5 evaluatees posicionados en la matriz
- AND puede hacer clic en cada punto para ver comentarios

#### Scenario: Director ve todos bajo su jerarquía (modificado)

- GIVEN director con 2 jefes y sus colaboradores (7 personas total)
- WHEN accede a la matriz
- THEN ve los 7 evaluatees posicionados en la matriz
- AND puede drill-down a competencias de cualquier evaluatee

#### Scenario: Director-general ve toda la organización (nuevo)

- GIVEN director-general activo
- WHEN accede a la matriz
- THEN ve todos los empleados del árbol corporativo bajo su mando
- AND puede filtrar por nivel (director/jefe/colaborador)

#### Scenario: Colaborador sin matriz (sin cambios)

- GIVEN perfil `colaborador`
- WHEN navega a la aplicación
- THEN NO ve el ítem "Matriz 9×9" en el menú
- AND si accede por URL directa, ve mensaje "No tienes acceso a esta función"

### Requirement: Separación de evaluación RH (decisión #4)

La matriz 3×3 es **exclusiva del jefe** y se enfoca en **potencial** para la matriz de sucesión. La evaluación formal de competencias la realiza **RH** de forma independiente. La matriz NO sustituye ni reemplaza la evaluación RH.

#### Scenario: RH evalúa competencias por separado

- GIVEN empleado en fase `cierre`
- WHEN RH completa su evaluación formal
- THEN registra calificación de competencias (escala 1–5) y cierre
- AND esta evaluación es la definitiva para el empleado
- AND es independiente de la calificación del jefe en la matriz

#### Scenario: Jefe no evalúa competenciasformalmente

- GIVEN jefe en fase `cierre`
- WHEN revisa la matriz con los tiers calculados de sus evaluatees
- THEN SOLO visualiza desempeño y potencial (tiers 1–3)
- AND NO califica competencias en escala 1–5 (eso es de RH)
- AND NO realiza cierre formal de evaluación (eso es de RH)

### Requirement: Comentarios opcionales por evaluatee

El sistema SHALL permitir al jefe agregar comentarios opcionales por evaluatee al calificar en la matriz.

#### Scenario: Agregar comentario

- GIVEN evaluatee "María" calificada con desempeño 7, potencial 8 (cuadrante 6)
- WHEN jefe agrega comentario "Listo para liderar proyecto transversal"
- THEN el comentario se guarda con la entrada de la matriz
- AND es visible al hacer clic en el punto de María en la matriz

#### Scenario: Sin comentario

- GIVEN evaluatee calificado sin comentario
- WHEN se guarda la calificación
- THEN la entrada se crea sin comentario
- AND no se muestra indicador de "comentario pendiente"

### Requirement: Perfil director-general y scope multi-nivel

El sistema SHALL soportar el perfil `director-general` con visibilidad de toda la jerarquía bajo su mando en la matriz. El scope de la matriz varía por perfil: jefe ve reportes directos, director ve todos los managers en su tramo, director-general ve toda la organización.

#### Scenario: Director-general ve toda la organización

- GIVEN director-general activo con árbol de 4 niveles (DG → 2 directores → 3 jefes → 6 colaboradores)
- WHEN accede a la matriz
- THEN ve 11 evaluatees posicionados en la matriz
- AND puede filtrar por nivel jerárquico

#### Scenario: Director ve managers en su tramo

- GIVEN director "A" con 2 jefes y 5 colaboradores indirectos
- WHEN accede a la matriz
- THEN ve los 2 jefes y los 5 colaboradores (todos bajo su jerarquía)
- AND no ve empleados de otro director

#### Scenario: Jefe ve solo reportes directos

- GIVEN jefe con 3 colaboradores
- WHEN accede a la matriz
- THEN ve solo sus 3 reportes directos
- AND el scope no cambia respecto al comportamiento anterior

### Requirement: Vista de matriz 3×3 visual

El sistema SHALL renderizar la matriz 3×3 como grilla visual con ejes Desempeño (tier 1–3) y Potencial (tier 1–3). Cada empleado se posiciona como punto según sus tiers calculados. Los 9 cuadrantes se colorean según el `colorHex` configurable por RH.

#### Scenario: Grilla con empleados posicionados

- GIVEN matriz con 5 evaluatees con tiers variados
- WHEN se renderiza la grilla
- THEN cada empleado aparece como punto en su celda correspondiente
- AND los cuadrantes muestran color de fondo según colorHex (ej. cuadrante 9 verde, cuadrante 1 rojo)

#### Scenario: Clic en punto muestra detalle

- GIVEN empleado "María" en performance tier=2, potential tier=3
- WHEN se hace clic en su punto
- THEN se muestra card con: nombre, performanceTier, potentialTier, cuadrante, comentarios
- AND botón "Ver competencias" navega a la vista de competencias del empleado

#### Scenario: Celda vacía sin empleados

- GIVEN cuadrante sin evaluatees asignados
- WHEN se renderiza
- THEN la celda aparece con color de cuadrante pero sin puntos
- AND no muestra placeholder ni mensaje de error

### Requirement: Sub-vista de red de competencias

El sistema SHALL proveer una tabla de competencias para un empleado individual mostrando autoevaluación vs evaluación RH por competencia. Accesible desde la matriz 9×9 o vía ruta directa `/evaluacion/9x9/competencias/[employeeId]`.

#### Scenario: Tabla de competencias

- GIVEN empleado evaluado en 5 competencias
- WHEN se accede a la vista de competencias
- THEN se muestra tabla con columnas: Competencia, Autoevaluación (1–5), Evaluación RH (1–5)
- AND cada fila resalta si hay brecha >1 entre auto y RH

#### Scenario: Sin evaluación RH aún

- GIVEN empleado con autoevaluación completada pero RH pendiente
- WHEN se accede a la vista de competencias
- THEN columna RH muestra "Pendiente"
- AND no hay comparación ni resalte de brecha

### Requirement: Rutas de jerarquía desde 9×9

El sistema SHALL exponer `/evaluacion/9x9/jerarquia` para drill-down jerárquico desde contexto 9×9. El menú lateral SHALL mostrar "Jerarquía" y "Competencias" como sub-ítems de "Matriz 9×9" para perfiles con acceso.

#### Scenario: Navegación a jerarquía

- GIVEN director-general en matriz 9×9
- WHEN navega a "Jerarquía" desde menú
- THEN ve árbol expandible DG → Director → Jefe → Colaborador
- AND cada nodo muestra nombre, puesto y cantidad de subordinados

## Non-goals

- **Evaluación RH**: la evaluación formal de competencias por RH es scope de A5 y C6 (evaluations-and-9x9-api).
- **Agregados de empresa**: no se soporta consolidar matrices de múltiples jefes en una vista global de empresa.
- **Historial**: no se guardan versiones anteriores de la matriz (solo la última calificación del ciclo actual).
- **Pendientes de ubicación**: empleados sin datos suficientes reciben tier 2 por defecto. El estado "pendiente" se difiere a cambio posterior.
- **API real**: en fase UI-first, la matriz se alimenta de fixtures JSON. La API real es C6.
- **Autenticación y RBAC**: esta spec define el behavior; la implementación de permisos es scope de C7.
- **Exportación**: no se soporta exportar la matriz a PDF, Excel u otro formato.
- **Comparación entre ciclos**: no se soporta comparar la matriz de un año con la del anterior.
- **Notificaciones**: no se envía email al jefe para calificar (scope de C7).
