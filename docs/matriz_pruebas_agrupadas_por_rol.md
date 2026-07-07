# Matriz de Pruebas E2E Agrupadas por Rol

## 1. RRHH
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-001 | Crear, editar o eliminar pilares | El sistema puede crear un pilar, editar o eliminar. | Revisar edición. |
| E2E-002 | Crear, editar o eliminar competencias | El sistema puede crear una competencia, editar o eliminar. | Función de editar inactiva. |
| E2E-003 | Editar definicion de niveles | El sistema puede editar la definición de niveles del 1-5. | OK |
| E2E-004 | Crear, editar o eliminar criterios | El sistema puede crear un criterio por competencia y nivel, editar o eliminar. | OK |
| E2E-005 | Editar niveles de aceptación | El sistema permite guardar las definiciones de nivel | OK |
| E2E-006 | Gestión de fases y ciclos | El sistema puede crear un ciclo si no hay activos, y avanzar en fase. | Error al avanzar fase |
| E2E-007 | Lista de evaluaciones RRHH | El sistema puede buscar personal. Ver evaluaciones, y evaluar. | OK |
| E2E-008 | Jerarquía de departamentos | El sistema puede mostrar todos los departamentos, ver empleados a evaluar y métricas | OK |

---

## 2. Director General
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-013 | Ver toda la organización y resultados obtenidos | El director general visualiza la estructura completa de la empresa con acceso a todos los niveles jerárquicos y resultados consolidados de evaluaciones. | Ver personal de gerencias o jefes de manera horizontal |
| E2E-014 | Visibilidad de metas | El director general puede ver las metas de todos los empleados de la organización, incluyendo avance y estado de cada meta. | OK |
| E2E-015 | Visibilidad de competencias | El director general puede ver las competencias evaluadas de todos los empleados, con las calificaciones asignadas por RRHH y la brecha con la autoevaluación. Se visualiza tabla y gráfico de radar. | OK |
| E2E-016 | Visibilidad de nine box | El director general puede visualizar la matriz 9×9 de toda la organización, con el posicionamiento de cada empleado según su desempeño y potencial obtenido. | OK |

---

## 3. Director
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-084 | Transición de medio año a cierre | Cambia el ciclo a la fase de cierre. | --- |
| E2E-058 | Ver matriz completa de su tramo | Ver todos los jefes y colaboradores de su jerarquía. | --- |
| E2E-091 | Ver su tramo organizacional | Ve las personas bajo su mando. | --- |
| E2E-141 | Ver actividad de su área | Puede ver las acciones de las personas de su área. | --- |

---

## 4. Jefe
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-050 | Ver tabla con evaluados de su equipo directo | Accede a la tabla de empleados, con resultados y avances obtenidos | --- |
| E2E-051 | Ver detalle de un evaluado | Ver una tarjeta con el evaluado seleccionado. | --- |
| E2E-052 | Agregar comentario a un evaluado | Comentar sobre el evaluado. | --- |
| E2E-054 | Ver brecha entre autoevaluación y evaluación RH | Ve la diferencia de puntuaciones. | --- |
| E2E-057 | Ver matriz completa de su tramo | Ve los colaboradores bajo su mando. | --- |

---

## 5. Colaborador
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-010 | Crear una categoría de metas | Crear categoría para sus metas. | --- |
| E2E-011 | Crear una meta dentro de una categoría | Crear metas en la categoría. | --- |
| E2E-012 | Asignar porcentaje de peso a categorías | Define el peso de cada categoría. | --- |
| E2E-019 | Validar que las ponderaciones sumen 100% | Sistema valida que los pesos sumen 100%. | --- |
| E2E-030 | Editar meta en fase de avance | Edita meta durante el ciclo de avance. | --- |
| E2E-031 | Registrar avance en una meta | Registra cuánto ha avanzado en una meta. | --- |
| E2E-040 | Calificar competencias | Califica competencias en autoevaluación. | --- |
| E2E-041 | Agregar comentarios de cierre | Escribir comentarios finales sobre el desempeño. | --- |
| E2E-042 | Enviar autoevaluación | Envía la autoevaluación. | --- |
