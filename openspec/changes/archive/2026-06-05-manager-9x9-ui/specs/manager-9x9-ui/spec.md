# Delta for manager-9x9

> Modifies: `openspec/specs/manager-9x9/spec.md`

## ADDED Requirements

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

## MODIFIED Requirements

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

## REMOVED Requirements

(Ninguno)
