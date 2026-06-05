# manager-9x9 Specification

## Purpose

Define la **matriz 9×9 de desempeño vs potencial**: ejes de medición, cuadrantes, quién califica (el jefe/director/gerente) y la separación explícita de la evaluación RH. Esta spec es la fuente de verdad para la matriz de potencial que alimenta la pantalla A6 (manager-9x9-ui) y el backend C6.

**Decisiones reflejadas:** #4 (fin de año: el jefe evalúa potenciales para matriz 9×9; NO sustituye la evaluación RH de competencias).

## Data Model

| Entity | Fields | Notes |
|--------|--------|-------|
| **NineBoxMatrix** | `id`, `cycleId`, `evaluatorId` | Instancia de la matriz para un evaluador en un ciclo. Un jefe tiene su propia matriz. |
| **NineBoxEntry** | `id`, `matrixId`, `evaluateeId`, `performanceScore` (1–9), `potentialScore` (1–9), `quadrant` (calculado), `comments?` | Calificación de un evaluado. Ambos scores determinan el cuadrante. |
| **NineBoxQuadrant** | `id`, `label`, `description`, `color`, `actionRecommendation` | Definición de cuadrante derivado de la posición en la matriz. |
| **NineBoxScale** | `axis` (`performance` \| `potential`), `level` (1–9), `label`, `description` | Definiciones de los 9 niveles por eje. |

### Matriz 9×9 — Layout

```
Potencial
  9 │ 3  │ 6  │ 9
  8 │ 3  │ 6  │ 9
  7 │ 3  │ 6  │ 9
  6 │ 2  │ 5  │ 8
  5 │ 2  │ 5  │ 8
  4 │ 2  │ 5  │ 8
  3 │ 1  │ 4  │ 7
  2 │ 1  │ 4  │ 7
  1 │ 1  │ 4  │ 7
    └────┴────┴────┘
      1-3   4-6   7-9
          Desempeño
```

### Cuadrantes definidos

| Cuadrant | Rango (Desempeño × Potencial) | Label | Descripción | Acción recomendada |
|----------|------------------------------|-------|-------------|-------------------|
| 1 | Desempeño bajo (1–3) × Potencial bajo (1–3) | Bajo desempeño, bajo potencial | Requiere acción correctiva inmediata | Plan de mejora o reasignación |
| 2 | Desempeño medio (4–6) × Potencial bajo (1–3) | Desempeño medio, bajo potencial | Estable pero sin potencial de crecimiento | Mantener, desarrollo de habilidades |
| 3 | Desempeño alto (7–9) × Potencial bajo (1–3) | Alto desempeño, bajo potencial | Excelente rendimiento, estancado | Recompensar, evitar sobrecarga |
| 4 | Desempeño bajo (1–3) × Potencial medio (4–6) | Bajo desempeño, potencial medio | Potencial sin materializar | Coaching, asignar mentores |
| 5 | Desempeño medio (4–6) × Potencial medio (4-6) | Desempeño medio, potencial medio | Promedio, crecimiento gradual | Desarrollo planificado |
| 6 | Desempeño alto (7–9) × Potencial medio (4–6) | Alto desempeño, potencial medio | Sólido,listo para más responsabilidad | Preparar para rol superior |
| 7 | Desempeño bajo (1–3) × Potencial alto (7–9) | Bajo desempeño, alto potencial | Talento desaprovechado | Investigar causas, reasignar si necesario |
| 8 | Desempeño medio (4–6) × Potencial alto (7–9) | Desempeño medio, alto potencial | Listo para aceleración | Reto, proyecto de alto impacto |
| 9 | Desempeño alto (7–9) × Potencial alto (7–9) | Alto desempeño, alto potencial | Estrella, sucesor natural | Promover, plan de sucesión |

## Requirements

### Requirement: Calificación de desempeño y potencial (decisión #4)

El sistema SHALL permitir al jefe/director/gerente calificar a cada evaluatee en dos ejes: desempeño (1–9) y potencial (1–9). La calificación es **independiente** de la evaluación RH de competencias.

#### Scenario: Jefe califica desempeño

- GIVEN jefe con 3 evaluatees en fase `cierre`
- WHEN abre la matriz 9×9
- THEN ve lista de sus evaluatees con slider o select para desempeño (1–9)
- AND puede asignar un valor de desempeño por evaluatee

#### Scenario: Jefe califica potencial

- GIVEN jefe con 3 evaluatees
- WHEN asigna potencial a cada uno
- THEN el sistema calcula el cuadrante automáticamente
- AND muestra el cuadrante con color y label correspondiente

#### Scenario: Calificación es independiente de evaluación RH

- GIVEN empleado con autoevaluación completada y evaluación RH pendiente
- WHEN jefe lo califica en 9×9
- THEN las calificaciones 9×9 no dependen de la autoevaluación ni de la evaluación RH
- AND las tres vías son paralelas e independientes

### Requirement: Cálculo automático de cuadrante

El sistema SHALL calcular el cuadrante automáticamente al asignar desempeño y potencial. El cuadrante se deriva de la posición en la matriz.

#### Scenario: Cuadrante 9 — estrella

- GIVEN evaluatee con desempeño 8 y potencial 9
- WHEN jefe guarda la calificación
- THEN el cuadrante calculado es 9 ("Alto desempeño, alto potencial")
- AND se muestra con color distinctivo y acción "Promover, plan de sucesión"

#### Scenario: Cuadrante 1 — bajo desempeño y potencial

- GIVEN evaluatee con desempeño 2 y potencial 1
- WHEN jefe guarda la calificación
- THEN el cuadrante calculado es 1 ("Bajo desempeño, bajo potencial")
- AND se muestra con color de alerta y acción "Plan de mejora o reasignación"

#### Scenario: Cambio de calificación recalcula cuadrante

- GIVEN evaluatee con desempeño 5 y potencial 5 (cuadrante 5)
- WHEN jefe cambia desempeño a 8
- THEN el cuadrante se recalcula a 6 ("Alto desempeño, potencial medio")

### Requirement: Vista de matriz por evaluador

Cada jefe/director/gerente/director-general SHALL ver su propia matriz 9×9. El scope de evaluatees varía por perfil: jefe ve reportes directos; director ve todos los managers bajo su jerarquía; director-general ve toda la organización. La matriz de un evaluador no es visible para otros evaluadores del mismo nivel.

(Previously: Solo jefe y director veían reportes directos; sin director-general ni scope multi-nivel.)

#### Scenario: Jefe ve su matriz (sin cambios)

- GIVEN jefe "Carlos" con 5 evaluatees directos
- WHEN accede a la matriz 9×9
- THEN ve los 5 evaluatees posicionados en la matriz
- AND puede hacer clic en cada punto para ver comentarios

#### Scenario: Director ve todos bajo su jerarquía (modificado)

- GIVEN director con 2 jefes y sus colaboradores (7 personas total)
- WHEN accede a la matriz 9×9
- THEN ve los 7 evaluatees posicionados en la matriz
- AND puede drill-down a competencias de cualquier evaluatee

#### Scenario: Director-general ve toda la organización (nuevo)

- GIVEN director-general activo
- WHEN accede a la matriz 9×9
- THEN ve todos los empleados del árbol corporativo bajo su mando
- AND puede filtrar por nivel (director/jefe/colaborador)

#### Scenario: Colaborador sin matriz (sin cambios)

- GIVEN perfil `colaborador`
- WHEN navega a la aplicación
- THEN NO ve el ítem "Matriz 9×9" en el menú
- AND si accede por URL directa, ve mensaje "No tienes acceso a esta función"

### Requirement: Separación de evaluación RH (decisión #4)

La matriz 9×9 es **exclusiva del jefe** y se enfoca en **potencial** para la matriz de sucesión. La evaluación formal de competencias la realiza **RH** de forma independiente. La 9×9 NO sustituye ni reemplaza la evaluación RH.

#### Scenario: RH evalúa competencias por separado

- GIVEN empleado en fase `cierre`
- WHEN RH completa su evaluación formal
- THEN registra calificación de competencias (escala 1–5) y cierre
- AND esta evaluación es la definitiva para el empleado
- AND es independiente de la calificación 9×9 del jefe

#### Scenario: Jefe no evalúa competenciasformalmente

- GIVEN jefe en fase `cierre`
- WHEN completa la calificación 9×9 de sus evaluatees
- THEN SOLO califica desempeño y potencial (1–9)
- AND NO califica competencias en escala 1–5 (eso es de RH)
- AND NO realiza cierre formal de evaluación (eso es de RH)

### Requirement: Definiciones de escala por eje

El sistema SHALL mantener definiciones de los 9 niveles para cada eje (desempeño y potencial). Las definiciones ayudan al jefe a calificar consistentemente.

#### Scenario: Definiciones de desempeño

- GIVEN jefe calificando desempeño
- WHEN selecciona un nivel
- THEN ve la definición del nivel (ej. nivel 7: "Supera consistentemente las expectativas")
- AND la definición es la misma para todos los evaluadores

#### Scenario: Definiciones de potencial

- GIVEN jefe calificando potencial
- WHEN selecciona un nivel
- THEN ve la definición del nivel (ej. nivel 8: "Listo para asumir roles de mayor complejidad en 1–2 años")
- AND la definición es la misma para todos los evaluadores

### Requirement: Comentarios opcionales por evaluatee

El sistema SHALL permitir al jefe agregar comentarios opcionales por evaluatee al calificar en la matriz 9×9.

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

El sistema SHALL soportar el perfil `director-general` con visibilidad de toda la jerarquía bajo su mando en la matriz 9×9. El scope de la matriz varía por perfil: jefe ve reportes directos, director ve todos los managers en su tramo, director-general ve toda la organización.

#### Scenario: Director-general ve toda la organización

- GIVEN director-general activo con árbol de 4 niveles (DG → 2 directores → 3 jefes → 6 colaboradores)
- WHEN accede a la matriz 9×9
- THEN ve 11 evaluatees posicionados en la matriz
- AND puede filtrar por nivel jerárquico

#### Scenario: Director ve managers en su tramo

- GIVEN director "A" con 2 jefes y 5 colaboradores indirectos
- WHEN accede a la matriz 9×9
- THEN ve los 2 jefes y los 5 colaboradores (todos bajo su jerarquía)
- AND no ve empleados de otro director

#### Scenario: Jefe ve solo reportes directos

- GIVEN jefe con 3 colaboradores
- WHEN accede a la matriz 9×9
- THEN ve solo sus 3 reportes directos
- AND el scope no cambia respecto al comportamiento anterior

### Requirement: Vista de matriz 9×9 visual

El sistema SHALL renderizar la matriz 9×9 como grilla visual con ejes Desempeño (X, 1–9) y Potencial (Y, 1–9). Cada empleado se posiciona como punto según sus scores. Los 9 cuadrantes se colorean según la definición de `NineBoxQuadrant`.

#### Scenario: Grilla con empleados posicionados

- GIVEN matriz con 5 evaluatees con scores variados
- WHEN se renderiza la grilla
- THEN cada empleado aparece como punto en su celda correspondiente
- AND los cuadrantes muestran color de fondo según definición (ej. cuadrante 9 verde, cuadrante 1 rojo)

#### Scenario: Clic en punto muestra detalle

- GIVEN empleado "María" en celda desempeño=7, potencial=8
- WHEN se hace clic en su punto
- THEN se muestra card con: nombre, scores, cuadrante, comentarios del jefe
- AND botón "Ver competencias" navega a `/evaluacion/9x9/competencias/[employeeId]`

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

### Requirement: Sliders de desempeño y potencial

El sistema SHALL proveer controles duales (sliders) para ajustar scores de desempeño (1–9) y potencial (1–9) por evaluatee. Al modificar, el cuadrante se recalcula en tiempo real.

#### Scenario: Ajuste de scores recalcula cuadrante

- GIVEN evaluatee con desempeño=5, potencial=5 (cuadrante 5)
- WHEN jefe mueve slider de desempeño a 8
- THEN el cuadrante se recalcula a 6 (Alto desempeño, potencial medio)
- AND el punto se reposiciona en la grilla sin recargar

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
- **API real**: en fase UI-first, la matriz se alimenta de fixtures JSON. La API real es C6.
- **Autenticación y RBAC**: esta spec define el behavior; la implementación de permisos es scope de C7.
- **Exportación**: no se soporta exportar la matriz a PDF, Excel u otro formato.
- **Comparación entre ciclos**: no se soporta comparar la matriz de un año con la del anterior.
- **Notificaciones**: no se envía email al jefe para calificar (scope de C7).
