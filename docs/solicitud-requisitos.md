# SED — Solicitud de Requisitos

> Fecha de corte: solo lo ya desarrollado (sin puntos nuevos)

---

## Fuentes y Relaciones

| Fuente | Tipo | Cubre |
|--------|------|-------|
| Lista maestra de requisitos | Datos de referencia | Los 33 requerimientos con entregables |
| Diseño de pantalla principal | Documento de diseño | Sección 1: Pantalla principal |
| Diseño de asignación de metas | Documento de diseño | Sección 2: Asignación de metas |
| Diseño de avance de medio año | Documento de diseño | Sección 3: Avance de medio año |
| Diseño de autoevaluación | Documento de diseño | Sección 4: Autoevaluación de fin de año |
| Diseño de administración RH | Documento de diseño | Sección 5: Administración RH |
| Diseño de matriz 9×9 | Documento de diseño | Sección 7: Matriz 9×9 |
| Diseño de jerarquía organizacional | Documento de diseño | Organigrama |
| Diseño de ponderación de metas | Documento de diseño | Peso y distribución de metas |
| Diseño de marco de competencias | Documento de diseño | Catálogo de competencias |
| Diseño de ciclo anual | Documento de diseño | Ciclo anual de evaluación |
| Diseño de estilo visual | Documento de diseño | Colores, tipografía y tema claro/oscuro |
| Decisiones del proyecto | Guías de decisiones del proyecto | Seguridad, roles, estructura general |
| Sistema de comentarios | Funcionalidad interna | Comentarios por requisito (guardados localmente) |

---

## Sección 1: Pantalla Principal

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 1 | Todos | Un sistema con menú lateral para navegar entre módulos | Pantalla con menú que se adapta al teléfono o computadora | Tener una estructura base navegable | En computadora el menú siempre visible, en teléfono se oculta y se abre con botón de menú |
| 2 | Todos | Que el menú solo muestre lo que puedo ver según mi rol | Menú que cambia según el perfil del usuario | Controlar qué módulos ve cada quien | Cada opción del menú se muestra solo si el perfil del usuario tiene acceso |
| 3 | Todos | Un tema claro y oscuro | La pantalla cambia entre modo claro y modo oscuro | Adaptar la vista según preferencia del usuario | Colores, bordes y tipografía consistentes. Los logos se adaptan al tema activo |
| 4 | Todos | Pantallas para cuando algo sale mal o está cargando | Mensajes y animaciones para error, datos vacíos o carga | Que el usuario entienda qué pasa en cada situación | Muestra mensajes claros de error, pantalla de "sin datos" y figuras de carga |

---

## Sección 2: Metas — Asignación de Inicio de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 5 | Colaborador | Agrupar metas en categorías personalizadas | Pantalla con tarjetas para crear, editar y eliminar categorías | Organizar metas relacionadas | La lista muestra: Nombre de la categoría, Descripción, Peso y Acciones (editar, eliminar). Son independientes de las competencias |
| 6 | Colaborador | Registrar metas con nombre, descripción, unidad, valor objetivo, peso y comportamiento | Formulario para agregar metas dentro de cada categoría. Cada meta puede ser hacia arriba (más es mejor) o hacia abajo (menos es mejor) | Definir indicadores y metas concretos por categoría | La unidad puede ser porcentaje (%), moneda ($) o numérico (#). El peso de todas las metas de una categoría suma 100%. Se visualiza una tabla con: Meta, Valor objetivo, Peso, KPI asociado y Acciones (editar, eliminar) |
| 7 | Colaborador / Jefe | Que se valide que los porcentajes siempre sumen 100 | Indicador visual que muestra el progreso y no deja guardar si no suma 100% | Asegurar que las ponderaciones sean correctas | Cada categoría suma 100% global. Las metas dentro de cada categoría suman 100%. No se puede guardar si hay error |
| 8 | Jefe | Ver las metas de mis colaboradores pero sin poder modificarlas | Pantalla de solo lectura para jefes con botón para solicitar cambios | Supervisión sin riesgo de borrado accidental | El jefe no puede borrar ni agregar metas. Al pedir cambios se abre una caja de comentarios tipo chat entre el jefe directo y el empleado, que puede ser sobre una meta individual o sobre todas las metas del empleado. El jefe escribe su observación, el empleado la ve, responde y al final decide si acepta o rechaza el cambio solicitado |
| 9 | Colaborador / Jefe | Un catálogo de indicadores predefinidos para reusar | Pantalla con lista de indicadores. Cada indicador puede ser hacia arriba (más es mejor) o hacia abajo (menos es mejor) | Tener indicadores estándar medibles y reutilizables | Los indicadores pueden ser numéricos, porcentaje o moneda. Se vinculan a una o más metas. Se visualiza Nombre, Descripción, Unidad, Objetivo |
| 10 | Jefe | Navegar entre mis colaboradores y gestionar sus metas | Selector de persona para cambiar entre colaboradores | Gestión centralizada de las metas del equipo | El selector marca "(Yo)" para el usuario actual. Usa el organigrama de la empresa para empleados a cargo. |
| 11 | Colaborador / Jefe | Descargar la asignación anual de metas en archivo | Botón "Descargar" que genera un archivo con todas las metas del empleado | Tener un respaldo impreso o digital de las metas fijadas al inicio del año | El archivo incluye: categoría, peso de categoría (%), meta, descripción, unidad, peso de la meta (%), valor objetivo, KPIs asociados. Se descarga con un clic |

---

## Sección 3: Metas — Avance de Medio Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 12 | Colaborador / Jefe | Comentar avance de metas existentes en la fase de medio año | Botón de comentarios con pantalla tipo chat entre el jefe directo y el empleado | Registrar retroalimentación para mejorar a mitad de año | No se pueden eliminar metas. Permite ver todos los comentarios. |
| 13 | Colaborador | Registrar el avance de cada meta | Formulario para ingresar avance con barra visual de progreso | Ver el progreso real contra el esperado | Indicador visual con colores. Avance por meta y por categoría |

---

## Sección 4: Mi Evaluación — Autoevaluación de Fin de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 14 | Colaborador | Autoevaluar mis competencias con escala del 1 al 5 | Pantalla con tarjetas de competencias y selector de puntaje | Que cada empleado evalúe sus propias competencias | Cada tarjeta muestra: nombre de la competencia, puntaje seleccionado (1-5), y espacio de comentarios |
| 15 | Colaborador | Ver un gráfico de radar con mis resultados de competencias | Visualización en forma de gráfico o tabla intercambiable | Ver fortalezas y debilidades de forma visual | La tabla muestra: Competencia, Puntaje, Mínimo esperado, Diferencia. Se puede cambiar entre gráfico y tabla. Los colores se adaptan al tema claro/oscuro |
| 16 | Colaborador | Cerrar mis metas al final del ciclo | Pantalla para registrar el cumplimiento final de cada meta | Formalizar el resultado de cada meta al cierre del año | Para cada meta se muestra: nombre, objetivo, KPIs, comentarios y avance final. Al cerrar el ciclo se bloquea la edición |
| 17 | Jefe / RH | Descargar la evaluación de empleados en archivo | Botón "Descargar" que genera un archivo con los datos de la tabla | Generar reportes de evaluación para imprimir o enviar por correo | El archivo incluye: nombre del empleado, perfil, avance general (%), estado de evaluación. Se descarga con un clic |

---

## Sección 5: Administración RH — Competencias

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 18 | RH | Administrar pilares (categorías de competencias) | Pantalla para crear, ver, editar y eliminar pilares | Mantener el catálogo de pilares de competencias de la empresa | La lista muestra: Nombre del pilar, Descripción, Acciones (editar, eliminar con confirmación) |
| 19 | RH | Administrar competencias dentro de cada pilar | Pantalla para crear, ver, editar y eliminar competencias | Definir las competencias específicas por pilar | La lista muestra: Nombre de la competencia, Pilar al que pertenece, Descripción, Acciones (editar, eliminar) |
| 20 | RH | Definir qué significa cada nivel de calificación por competencia | Pantalla para establecer criterios del 1 al 5 por competencia y nivel | Estandarizar la evaluación con descripciones claras de cada nivel | Escala fija 1 a 5. Los criterios cambian según el nivel esperado |
| 21 | RH | Definir el puntaje mínimo requerido por perfil y competencia | Pantalla para establecer niveles de aceptación por perfil | Saber si un empleado cumple el mínimo esperado en cada competencia | Por cada perfil se muestra: Competencia, Puntaje mínimo aceptado. Se muestra cuadro comparativo entre perfiles por competencia |

---

## Sección 6: Evaluación RH — Fin de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 22 | RH | Ver una lista de empleados con estado de su evaluación | Pantalla con tabla de empleados y su estado actual | Que RH dé seguimiento a las evaluaciones pendientes | La tabla muestra: Nombre del empleado, Perfil, Estado de evaluación (pendiente, en progreso, completada), Acciones (evaluar) |
| 23 | RH | Comparar la autoevaluación del empleado y realizar evaluación de RH | Pantalla que muestra resumen de calificaciones, metas y competencias de empleados. | Detectar diferencias entre cómo se ve el empleado y cómo lo ve RH | Por cada competencia se muestra: Autoevaluación del empleado, Evaluación de RH, Diferencia entre ambas. Por cada meta se permite agregar comentarios. Por cada competencia se permite evaluar en escala 1-5 y agregar comentarios.  |

---

## Sección 7: Matriz 9×9 — Evaluación de Jefes

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 24 | Jefe | Ubicar a mis colaboradores en una cuadrícula de desempeño contra potencial | Cuadrícula interactiva para visualizar a cada empleado | Visualizar el talento del equipo en un solo vistazo | La cuadrícula se puede usar interactivamente. No reemplaza la evaluación de competencias. |
| 25 | Jefe / RH | Navegar el organigrama de la empresa | Árbol visual para explorar la jerarquía organizacional | Navegar entre colaboradores según su posición en la empresa | Vistas disponibles: vista de Directores (muestra el equipo a cargo) y vista de RH (muestra toda la empresa con el promedio de evaluación por área). Al seleccionar una persona del árbol se muestra: nombre, cargo, perfil, promedio de evaluación y personas a cargo |
| 26 | Jefe | Consultar las competencias de mis colaboradores | Pantalla donde el jefe ve las competencias de sus subordinados | Que el jefe conozca las evaluaciones de su equipo | El jefe puede ver gráfico de radar y tabla. No puede calificar (esa tarea es de RH) |

---

## Sección 8: Mis Evaluados

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 27 | Jefe / Director | Ver la lista de mis colaboradores directos | Pantalla con lista de personas a cargo | Acceso rápido a metas y evaluaciones del equipo | La lista muestra: Nombre, Cargo, Perfil, Estado de evaluación, Acción (evaluar). El jefe ve sus empleados directos, un director ve toda su área. Puede agregar comentarios por meta. |

---

## Sección 9: Perfil de Usuario

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 28 | Todos | Ver mi página de perfil con datos personales y actividad reciente | Pantalla con nombre, correo, historial de actividad | Ver la identidad del usuario y su actividad en el sistema | El historial muestra: Fecha, Acción realizada, Detalle de la actividad. Además se ve el nombre, correo, botón para cerrar sesión y la actividad por perfil |
| 29 | Todos | Ingresar con usuario y contraseña al sistema | Pantalla de inicio de sesión con formulario de credenciales | Ver las secciones correspondientes a un usuario | Ingreso con usuario empresarial, y contraseña propia. Cargar secciones correspondientes al usuario |

---

## Archivos Clave

| Documento | Rol |
|-----------|-----|
| Lista maestra de requisitos | Fuente principal de los 33 requerimientos |
| Sistema de comentarios por requisito | Notas y comentarios en cada requerimiento |
| Pantalla de consulta de requisitos | Interfaz para revisar esta lista de requerimientos |
| Mapa general del proyecto | Orden y dependencias entre los documentos de diseño |
| Documentos de diseño por módulo | Diseños detallados de cada sección del sistema |
| Documentos de diseño archivados | Diseños de funcionalidades ya terminadas |
| Guías del proyecto | Decisiones sobre seguridad, roles y estructura general |

