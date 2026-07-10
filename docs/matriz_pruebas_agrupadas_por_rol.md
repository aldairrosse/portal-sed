# Matriz de Pruebas E2E Agrupadas por Rol

## 1. RRHH
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-001 | Crear, editar o eliminar pilares | El sistema puede crear un pilar, editar o eliminar. | Revisar edición. OK |
| E2E-002 | Crear, editar o eliminar competencias | El sistema puede crear una competencia, editar o eliminar. | Función de editar inactiva. OK |
| E2E-003 | Editar definicion de niveles | El sistema puede editar la definición de niveles del 1-5. | OK |
| E2E-004 | Crear, editar o eliminar criterios | El sistema puede crear un criterio por competencia y nivel, editar o eliminar. | Recargar en uno. OK |
| E2E-005 | Editar niveles de aceptación | El sistema permite guardar las definiciones de nivel | OK |
| E2E-006 | Gestión de fases y ciclos | El sistema puede crear un ciclo si no hay activos, y avanzar en fase. | Error al avanzar fase. OK |
| E2E-007 | Lista de evaluaciones RRHH | El sistema puede buscar personal. Ver evaluaciones, y evaluar. | OK |
| E2E-008 | Jerarquía de departamentos | El sistema puede mostrar todos los departamentos, ver empleados a evaluar y métricas | OK |

---

## 2. Director General
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-013 | Ver toda la organización y resultados obtenidos | El director general visualiza la estructura completa de la empresa con acceso a todos los niveles jerárquicos y resultados consolidados de evaluaciones. | Ver personal de gerencias o jefes de manera horizontal. OK |
| E2E-014 | Visibilidad de metas | El director general puede ver las metas de todos los empleados de la organización, incluyendo avance y estado de cada meta. | OK |
| E2E-015 | Visibilidad de competencias | El director general puede ver las competencias evaluadas de todos los empleados, con las calificaciones asignadas por RRHH y la brecha con la autoevaluación. Se visualiza tabla y gráfico de radar. | OK |
| E2E-016 | Visibilidad de nine box | El director general puede visualizar la matriz 9×9 de toda la organización, con el posicionamiento de cada empleado según su desempeño y potencial obtenido. | OK |

---

## 3. Director
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-058 | Ver matriz completa de su tramo | Ver todos los jefes y colaboradores de su jerarquía. | OK |
| E2E-091 | Ver su tramo organizacional | Ve las personas bajo su mando. | OK |
---

## 4. Jefe
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-050 | Ver tabla con evaluados de su equipo directo | Accede a la tabla de empleados, con resultados y avances obtenidos | Ver avance de asignación en inicio de año |
| E2E-051 | Ver detalle de un evaluado | Ver una tarjeta con el evaluado seleccionado. | --- |
| E2E-052 | Agregar comentario a un evaluado | Comentar sobre el evaluado. | OK. Evitar confusión de modales |
| E2E-054 | Ver brecha entre autoevaluación y evaluación RH | Ve la diferencia de puntuaciones con gráfico de radar y tabla | --- |
| E2E-057 | Ver matriz completa de su tramo | Ve los colaboradores bajo su mando en matriz 3x3. | --- |

---

## 5. Colaborador
| Id | Funcionalidad | Descripción | Comentarios |
|----|---------------|-------------|-------------|
| E2E-010 | Categoría de metas con peso | Crear categoría para sus metas con pesos, y descripción. | OK |
| E2E-011 | Meta dentro de una categoría con peso | Crear metas en la categoría con todos los campos. | Error al vincular KPIS |
| E2E-019 | Validar que las ponderaciones sumen 100% | Sistema valida que los pesos sumen 100%. | OK |
| E2E-019 | Kpis en metas | Sistema permite crear editar o eliminar kpis, y asignarlos a metas | OK |


Cambios 
> Agregar notificación por correo en comentarios. --

> Cambiar de departamentos a un usuario. OK

> Agregar botón de propuesta de meta. OK

> Agregar gestión y visualización de pilares de metas

> Agregar botón de descargar global. OK

