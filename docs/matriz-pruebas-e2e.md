# Matriz de Pruebas E2E — SED Evaluación de Desempeño

## Sección 1: Acceso y Sesión

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-001 | Cualquier usuario | Acceso | Iniciar sesión con credenciales válidas | El usuario ingresa su correo y contraseña. Si los datos son correctos, el sistema lo redirige a la pantalla principal según su perfil. |
| E2E-002 | Cualquier usuario | Acceso | Rechazo por credenciales incorrectas | Si el usuario ingresa una contraseña incorrecta, el sistema muestra un mensaje de error y no permite el acceso. |
| E2E-003 | Cualquier usuario | Acceso | Cierre de sesión | El usuario presiona "Cerrar sesión". El sistema lo regresa a la pantalla de login y la sesión queda invalidada. |
| E2E-004 | Cualquier usuario | Acceso | Mantener sesión activa | Si el usuario está navegando y la sesión está activa, puede cambiar de pantalla sin necesidad de volver a ingresar sus datos. |
| E2E-005 | Cualquier usuario | Acceso | Redirección por rol | Al iniciar sesión, el sistema redirige al usuario a la pantalla que le corresponde según su perfil (colaborador, jefe, RH, etc.). |

---

## Sección 2: Mis Metas (Objetivos)

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-010 | Colaborador | Mis Metas | Crear una categoría de metas | El colaborador crea una nueva categoría (por ejemplo: "Resultados de negocio"). Puede asignarle un nombre, descripción y un porcentaje de peso. |
| E2E-011 | Colaborador | Mis Metas | Crear una meta dentro de una categoría | Dentro de una categoría, el colaborador crea una meta con nombre, descripción, unidad de medida (porcentaje, moneda o número) y una meta numérica a alcanzar. |
| E2E-012 | Colaborador | Mis Metas | Asignar porcentaje de peso a categorías | El colaborador define qué peso tiene cada categoría. Todas las categorías deben sumar exactamente 100%. Si no suman 100%, el sistema no permite guardar. |
| E2E-013 | Colaborador | Mis Metas | Asignar porcentaje de peso a metas | Dentro de cada categoría, las metas también deben sumar 100%. El sistema valida esto antes de permitir el guardado. |
| E2E-014 | Colaborador | Mis Metas | Editar una meta existente | El colaborador puede cambiar el nombre, descripción, meta numérica o unidad de medida de una meta que ya creó. |
| E2E-015 | Colaborador | Mis Metas | Eliminar una meta | El colaborador puede eliminar una meta que ya no necesita. El sistema pide confirmación antes de borrarla. |
| E2E-016 | Colaborador | Mis Metas | Eliminar una categoría | El colaborador puede eliminar una categoría completa. Si tiene metas, el sistema advierte que también se eliminarán. |
| E2E-017 | Colaborador | Mis Metas | Vincular un KPI a una meta | El colaborador puede conectar un indicador de rendimiento (KPI) existente a una de sus metas. Una meta puede tener varios KPIs. |
| E2E-018 | Colaborador | Mis Metas | Fijar metas al terminar la fase de asignación | Cuando las metas están listas, el colaborador las "fija". A partir de ese momento, las metas quedan en estado listo para seguimiento. |
| E2E-019 | Colaborador | Mis Metas | Validar que las ponderaciones sumen 100% | Antes de fijar las metas, el sistema valida que los pesos de categorías sumen 100% y que dentro de cada categoría las metas también sumen 100%. Si hay un error, muestra qué categoría está incompleta. |
| E2E-020 | Colaborador | Mis Metas | Ver mis metas en modo lectura | El colaborador puede ver todas sus categorías y metas con su estado actual, porcentajes de avance y KPIs vinculados. |
| E2E-021 | Jefe | Mis Metas | Ver metas de un evaluado | El jefe puede ver las categorías y metas de las personas que tiene a su cargo, pero en modo lectura (sin poder editar). |
| E2E-022 | Jefe | Mis Metas | Solicitar cambio en meta de evaluado | El jefe puede enviar una solicitud de cambio a una meta de su evaluado. Se abre un formulario donde escribe sus comentarios y la solicitud queda registrada. |
| E2E-023 | Gerente Tienda / Divisional / Regional | Mis Metas | Ver metas de evaluados | Estos perfiles pueden ver las metas de sus evaluados en modo lectura, igual que el jefe. |
| E2E-024 | Director | Mis Metas | Ver metas de toda su jerarquía | El director puede ver las metas de todas las personas que están debajo de él en la estructura organizacional. |
| E2E-025 | RH | Mis Metas | Ver y editar sus propias metas | RH tiene control total sobre sus propias metas, igual que un colaborador. No puede editar metas de otros. |

---

## Sección 3: Seguimiento de Metas (Medio de Año)

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-030 | Colaborador | Seguimiento | Editar meta en fase de avance | Cuando el ciclo está en fase de medio año, el colaborador puede editar nombre, descripción y meta numérica, pero NO puede eliminar ni crear metas nuevas. |
| E2E-031 | Colaborador | Seguimiento | Registrar avance en una meta | El colaborador indica cuánto ha avanzado en una meta (porcentaje o monto). El sistema actualiza el indicador de progreso. |
| E2E-032 | Colaborador | Seguimiento | Bloqueo de eliminar meta en avance | Si el colaborador intenta eliminar una meta durante la fase de avance, el sistema no lo permite. El botón de eliminar está deshabilitado. |
| E2E-033 | Colaborador | Seguimiento | Bloqueo de crear meta nueva en avance | Si el colaborador intenta crear una meta nueva durante la fase de avance, el sistema no lo permite. El botón de crear no está disponible. |
| E2E-034 | Colaborador | Seguimiento | Ver semáforo de avance | El sistema muestra un indicador visual (semáforo) que refleja el porcentaje de avance de cada meta según el valor registrado. |

---

## Sección 4: Autoevaluación (Fin de Año)

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-040 | Colaborador | Autoevaluación | Calificar competencias | El colaborador califica cada una de sus competencias en una escala del 1 al 5, donde 1 es "necesita mejorar" y 5 es "excepcional". |
| E2E-041 | Colaborador | Autoevaluación | Agregar comentarios de cierre | Después de calificar, el colaborador puede escribir comentarios finales sobre su desempeño del año. |
| E2E-042 | Colaborador | Autoevaluación | Enviar autoevaluación | El colaborador envía su autoevaluación. A partir de ese momento, su evaluación queda en estado "completada". |
| E2E-043 | Colaborador | Autoevaluación | Ver estado de mi evaluación | El colaborador puede ver si su autoevaluación está pendiente, en proceso o completada. |
| E2E-044 | Colaborador | Autoevaluación | Bloqueo de editar metas en cierre | Durante la fase de cierre, el colaborador NO puede editar sus metas ni registrar avances. Solo puede realizar su autoevaluación. |

---

## Sección 5: Evaluación 9×9 (Jefe)

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-050 | Jefe | Matriz 9×9 | Ver matriz con evaluados | El jefe accede a la matriz de 3×3 y ve a sus evaluados posicionados según su desempeño y potencial calculados automáticamente. |
| E2E-051 | Jefe | Matriz 9×9 | Ver detalle de un evaluado | Al hacer clic en el punto de un evaluado, el jefe ve una tarjeta con su nombre, desempeño, potencial, cuadrante y comentarios. |
| E2E-052 | Jefe | Matriz 9×9 | Agregar comentario a un evaluado | El jefe puede escribir un comentario opcional sobre un evaluado. Este comentario se guarda con la entrada de la matriz. |
| E2E-053 | Jefe | Matriz 9×9 | Ver red de competencias | Desde la matriz, el jefe puede ver una tabla comparativa de autoevaluación vs evaluación RH de cada competencia del evaluado. |
| E2E-054 | Jefe | Matriz 9×9 | Ver brecha entre autoevaluación y evaluación RH | Si la diferencia entre la autoevaluación y la evaluación RH es mayor a 1 punto, el sistema resalta esa competencia comoAlerta. |
| E2E-055 | Gerente Tienda | Matriz 9×9 | Ver matriz con sus evaluados | El gerente de tienda accede a su propia matriz y ve a los colaboradores de su tienda posicionados. |
| E2E-056 | Divisional | Matriz 9×9 | Ver matriz con managers | El divisional ve a los gerentes de tienda y colaboradores bajo su mando en la matriz. |
| E2E-057 | Regional | Matriz 9×9 | Ver matriz regional | El regional ve a todos los evaluados de su región en la matriz. |
| E2E-058 | Director | Matriz 9×9 | Ver matriz completa de su tramo | El director ve a todos los jefes, gerentes y colaboradores bajo su jerarquía en la matriz. |
| E2E-059 | Director General | Matriz 9×9 | Ver toda la organización | El director general ve a todos los empleados de la empresa en la matriz. Puede filtrar por nivel jerárquico. |
| E2E-060 | Colaborador | Matriz 9×9 | Sin acceso a matriz | El colaborador NO ve la opción "Matriz 9×9" en el menú. Si intenta acceder por otra vía, ve un mensaje de "No tienes acceso". |
| E2E-061 | RH | Matriz 9×9 | Sin acceso directo a matriz de jefes | RH NO realiza la calificación 9×9. Su evaluación es por separado (evaluación formal). |

---

## Sección 6: Evaluación Formal RH

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-070 | RH | Evaluación RH | Ver lista de empleados a evaluar | RH ve una lista de todos los empleados que necesita evaluar en el ciclo actual. |
| E2E-071 | RH | Evaluación RH | Evaluar competencias de un empleado | RH califica cada competencia del empleado en escala 1–5. Esta es la evaluación oficial y definitiva. |
| E2E-072 | RH | Evaluación RH | Agregar comentarios de evaluación | RH puede escribir comentarios sobre el desempeño del empleado durante la evaluación. |
| E2E-073 | RH | Evaluación RH | Enviar evaluación | RH envía la evaluación. El estado del empleado cambia a "evaluación completada". |
| E2E-074 | RH | Evaluación RH | Ver brecha entre autoevaluación y evaluación RH | RH puede ver la diferencia entre lo que el empleado se autoevaluó y lo que RH calificó. |
| E2E-075 | RH | Evaluación RH | Ver estado de evaluaciones pendientes | RH ve un resumen de cuántas evaluaciones tiene pendientes, completadas y en proceso. |
| E2E-076 | Director General | Evaluación RH | Ver resultados de evaluaciones | El director general puede ver los resultados de las evaluaciones de toda la organización. |

---

## Sección 7: Transición de Ciclos

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-080 | Director | Ciclos | Transición de inicio a medio año | El director puede cambiar el ciclo de la fase "Inicio de año" a "Medio de año". A partir de ese momento, cambian las acciones disponibles para todos. |
| E2E-081 | Director General | Ciclos | Transición de inicio a medio año | El director general también puede realizar esta transición. |
| E2E-082 | RH | Ciclos | Transición de inicio a medio año | RH también puede cambiar de fase. |
| E2E-083 | Director | Ciclos | Transición de medio año a cierre | El director puede cambiar el ciclo a la fase de cierre, habilitando las evaluaciones. |
| E2E-084 | Director General | Ciclos | Transición de medio año a cierre | El director general puede realizar esta transición. |
| E2E-085 | RH | Ciclos | Transición de medio año a cierre | RH puede cambiar a la fase de cierre. |
| E2E-086 | Jefe | Ciclos | Sin acceso a transición | El jefe NO puede cambiar de fase. Solo los roles superiores pueden hacerlo. |
| E2E-087 | Colaborador | Ciclos | Sin acceso a transición | El colaborador NO puede cambiar de fase. |
| E2E-088 | Cualquier usuario | Ciclos | Ver fase actual del ciclo | Todos los usuarios pueden ver en qué fase está el ciclo actual (inicio, medio año o cierre). |

---

## Sección 8: Organización y Jerarquía

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-090 | RH | Organización | Ver estructura organizacional | RH puede ver el árbol completo de la organización, con todos los niveles jerárquicos. |
| E2E-091 | Director | Organización | Ver su tramo jerárquico | El director puede ver las personas que están bajo su mando en la estructura. |
| E2E-092 | Jefe | Organización | Ver sus reportes directos | El jefe puede ver únicamente a las personas que reportan directamente a él. |
| E2E-093 | Colaborador | Organización | Ver su información | El colaborador solo puede ver su propia información y la de su jefe directo. |
| E2E-094 | Director General | Organización | Ver toda la organización | El director general puede ver la estructura completa de la empresa. |
| E2E-095 | Director | Organización | Ver métricas de área | Al seleccionar un nodo de la organización, el director puede ver métricas como cantidad de empleados, avance de evaluaciones, etc. |
| E2E-096 | RH | Organización | Ver métricas de área | RH puede ver métricas de cualquier área de la organización. |

---

## Sección 9: Competencias

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-100 | RH | Competencias | Ver catálogo de competencias | RH puede ver la lista completa de competencias disponibles en el sistema. |
| E2E-101 | RH | Competencias | Crear una competencia nueva | RH puede agregar una competencia nueva al catálogo con nombre, descripción y nivel de aceptación. |
| E2E-102 | RH | Competencias | Editar una competencia | RH puede modificar el nombre, descripción o niveles de una competencia existente. |
| E2E-103 | RH | Competencias | Eliminar una competencia | RH puede eliminar una competencia que ya no se utiliza. El sistema pide confirmación. |
| E2E-104 | Colaborador | Competencias | Ver competencias asignadas | El colaborador puede ver las competencias que le han sido asignadas para su evaluación. |
| E2E-105 | Jefe | Competencias | Ver competencias de evaluados | El jefe puede ver las competencias asignadas a sus evaluados. |

---

## Sección 10: Panel de Mis Evaluados

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-110 | Jefe | Mis Evaluados | Ver lista de evaluados | El jefe ve una tabla con todos sus evaluados, incluyendo nombre, puesto, estado de evaluación y avance de metas. |
| E2E-111 | Jefe | Mis Evaluados | Ver detalle de un evaluado | Al seleccionar un evaluado, el jefe puede ver sus metas, competencias, estado de evaluación y progreso. |
| E2E-112 | Jefe | Mis Evaluados | Filtrar evaluados por estado | El jefe puede filtrar la lista para ver solo evaluados pendientes, en proceso o completados. |
| E2E-113 | Gerente Tienda | Mis Evaluados | Ver evaluados de su tienda | El gerente ve a los colaboradores de su tienda en la lista de evaluados. |
| E2E-114 | Divisional | Mis Evaluados | Ver evaluados de su división | El divisional ve a los gerentes y colaboradores de su división. |
| E2E-115 | Regional | Mis Evaluados | Ver evaluados de su región | El regional ve a todos los evaluados de su región. |
| E2E-116 | Director | Mis Evaluados | Ver evaluados de toda su jerarquía | El director ve a todas las personas bajo su mando en la lista de evaluados. |

---

## Sección 11: Mi Evaluación (Vista del Empleado)

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-120 | Colaborador | Mi Evaluación | Ver resumen de mi evaluación | El colaborador puede ver un resumen de su evaluación: metas, competencias, calificaciones y estado. |
| E2E-121 | Colaborador | Mi Evaluación | Ver resultado de competencias | El colaborador ve las calificaciones de sus competencias (autoevaluación y, si está disponible, la evaluación de RH). |
| E2E-122 | Colaborador | Mi Evaluación | Ver avance de metas | El colaborador ve el porcentaje de avance de cada una de sus metas. |
| E2E-123 | Colaborador | Mi Evaluación | Ver estado de la evaluación | El colaborador puede ver si su evaluación está pendiente, en proceso o completada. |

---

## Sección 12: Seguridad y Permisos

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-130 | Colaborador | Seguridad | Sin acceso a evaluación de otros | Un colaborador NO puede ver las metas ni evaluaciones de otro colaborador. |
| E2E-131 | Colaborador | Seguridad | Sin acceso a administrar competencias | Un colaborador NO puede crear, editar o eliminar competencias. |
| E2E-132 | Jefe | Seguridad | Sin acceso a transición de ciclo | Un jefe NO puede cambiar la fase del ciclo. |
| E2E-133 | Jefe | Seguridad | Sin acceso a evaluación formal RH | Un jefe NO puede realizar la evaluación formal de competencias (eso es de RH). |
| E2E-134 | RH | Seguridad | Sin acceso a matriz 9×9 como evaluador | RH NO califica en la matriz 9×9. Su evaluación es por separado. |
| E2E-135 | Director General | Seguridad | Solo lectura en metas | El director general solo puede ver las metas, no puede crear ni editar. Tiene acceso total de administrador en otras áreas. |
| E2E-136 | Cualquier usuario | Seguridad | Bloqueo por sesión expirada | Si la sesión del usuario expira, el sistema lo redirige al login y muestra un mensaje. |
| E2E-137 | Cualquier usuario | Seguridad | Protección de datos personales | Un usuario no puede ver información salarial o personal sensible de otro usuario que no esté en su jerarquía. |

---

## Sección 13: Actividad y Auditoría

| Id | Usuario | Sección | Funcionalidad | Descripción |
|----|---------|---------|---------------|-------------|
| E2E-140 | RH | Auditoría | Ver registro de actividad | RH puede ver un historial de acciones realizadas por los usuarios (quién creó, editó o eliminó algo). |
| E2E-141 | Director | Auditoría | Ver actividad de su área | El director puede ver las acciones de las personas de su área. |
| E2E-142 | Colaborador | Auditoría | Ver mi propia actividad | El colaborador solo puede ver sus propias acciones. |

---

## Resumen de Cobertura por Rol

| Rol | Pruebas E2E | Acciones Principales |
|-----|-------------|---------------------|
| Colaborador | ~25 | Crear/editar metas, autoevaluación, ver resultados |
| Jefe | ~20 | Ver evaluados, matriz 9×9, solicitar cambios |
| Vendedor | ~15 | Similar a colaborador (metas propias, autoevaluación) |
| Gerente Tienda | ~18 | Ver evaluados de tienda, matriz 9×9 |
| Divisional | ~18 | Ver evaluados de división, matriz 9×9 |
| Regional | ~18 | Ver evaluados de región, matriz 9×9 |
| Director | ~25 | Transición de ciclo, matriz 9×9 completa, métricas |
| Director General | ~20 | Toda la organización, transición, solo lectura en metas |
| RH | ~22 | Evaluar competencias, administrar catálogo, auditoría |

---

## Flujo Completo de un Ciclo Anual

1. **Inicio de año (Fase: Asignación)**
   - RH crea el ciclo y define las competencias.
   - Cada colaborador crea sus categorías y metas.
   - El jefe revisa las metas y puede solicitar cambios.
   - Se fijan las metas al finalizar la fase.

2. **Medio de año (Fase: Avance)**
   - Los colaboradores registran avances en sus metas.
   - Se pueden editar metas pero NO eliminar ni crear nuevas.
   - Se muestra el semáforo de progreso.

3. **Fin de año (Fase: Cierre)**
   - Los colaboradores hacen su autoevaluación.
   - Los jefes califican en la matriz 9×9 (desempeño y potencial).
   - RH realiza la evaluación formal de competencias.
   - Se cierra el ciclo y se generan los resultados.

4. **Resultados**
   - Cada empleado ve su evaluación completa.
   - RH y directores ven los resultados consolidados.
   - Se prepara el siguiente ciclo.
