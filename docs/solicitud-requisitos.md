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
| Decisiones del proyecto | Guías de arquitectura | Seguridad, roles, estructura general |
| Sistema de comentarios | Funcionalidad interna | Comentarios por requisito (guardados localmente) |

---

## Sección 1: Pantalla Principal

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 1 | Todos los usuarios | Un sistema con menú lateral para navegar entre módulos | Pantalla con menú que se adapta al teléfono o computadora | Tener una estructura base navegable | En computadora el menú siempre visible, en teléfono se oculta y se abre con botón de menú |
| 2 | Todos los usuarios | Que el menú solo muestre lo que puedo ver según mi rol | Menú que cambia según el perfil del usuario | Controlar qué módulos ve cada quien | Cada opción del menú se muestra solo si el perfil del usuario tiene acceso |
| 3 | Todos los usuarios | Un tema claro y oscuro | La pantalla cambia entre modo claro y modo oscuro | Adaptar la vista según preferencia del usuario | Colores, bordes y tipografía consistentes. Los logos se adaptan al tema activo |
| 4 | Todos los usuarios | Pantallas para cuando algo sale mal o está cargando | Mensajes y animaciones para error, datos vacíos o carga | Que el usuario entienda qué pasa en cada situación | Muestra mensajes claros de error, pantalla de "sin datos" y figuras de carga |

---

## Sección 2: Metas — Asignación de Inicio de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 5 | Colaborador | Agrupar metas en categorías personalizadas | Pantalla con tarjetas para crear, editar y eliminar categorías | Organizar metas relacionadas | Se pueden crear, ver, editar y eliminar categorías. Son independientes de las competencias |
| 6 | Colaborador | Registrar metas con nombre, descripción, unidad, valor esperado y peso | Formulario para agregar metas dentro de cada categoría | Definir indicadores concretos por categoría | La unidad puede ser porcentaje (%) o moneda ($). El peso de todas las metas de una categoría suma 100% |
| 7 | Colaborador / Jefe | Que se valide que los porcentajes siempre sumen 100 | Indicador visual que muestra el progreso y no deja guardar si no cuadra | Asegurar que las ponderaciones sean correctas | Cada categoría suma 100%. Las metas dentro de cada categoría suman 100%. No se puede guardar si hay error |
| 8 | Jefe | Ver las metas de mis colaboradores pero sin poder modificarlas | Pantalla de solo lectura para jefes con botón para solicitar cambios | Supervisión sin riesgo de borrado accidental | El jefe no puede borrar ni agregar metas. Solo solicitar cambios. El dueño decide si acepta |
| 9 | Colaborador / Jefe | Un catálogo de indicadores predefinidos para reusar | Pantalla con lista de indicadores que se pueden asignar a las metas | Tener indicadores estándar reutilizables | Los indicadores pueden ser numéricos, porcentaje o moneda. Se vinculan a una o más metas |
| 10 | Jefe | Navegar entre mis colaboradores y gestionar sus metas | Selector de persona para cambiar entre colaboradores | Gestión centralizada de las metas del equipo | El selector marca "(Yo)" para el usuario actual. Usa el organigrama de la empresa |
| 11 | Colaborador / Jefe | Descargar la asignación anual de metas en archivo | Botón "Descargar" que genera un archivo con todas las metas del empleado | Tener un respaldo impreso o digital de las metas fijadas al inicio del año | El archivo incluye: categoría, peso de categoría (%), meta, descripción, unidad, peso de la meta (%), valor objetivo, KPIs asociados. Se descarga con un clic |

---

## Sección 3: Metas — Avance de Medio Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 12 | Colaborador | actualizar mis metas existentes en la fase de medio año | Pantalla donde se ajustan los campos permitidos de las metas | Registrar cambios de mitad de año | No se pueden eliminar metas. Solo se ajustan los valores permitidos |
| 13 | Colaborador | registrar el avance de cada meta | Formulario para ingresar avance con barra visual de progreso | Ver el progreso real contra el esperado | Indicador visual con colores. Avance por meta y por categoría |

---

## Sección 4: Mi Evaluación — Autoevaluación de Fin de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 14 | Colaborador | autoevaluar mis competencias con escala del 1 al 5 | Pantalla con tarjetas de competencias y selector de puntaje | Que cada empleado evalúe sus propias competencias | Escala 1 a 5 con descripciones según el perfil del empleado |
| 15 | Colaborador | ver un gráfico de radar con mis resultados de competencias | Visualización en forma de gráfico o tabla intercambiable | Ver fortalezas y debilidades de forma visual | Se puede cambiar entre gráfico y tabla. Los colores se adaptan al tema claro/oscuro |
| 16 | Colaborador | cerrar mis metas al final del ciclo | Pantalla para registrar el cumplimiento final de cada meta | Formalizar el resultado de cada meta al cierre del año | Se registra valor final y comentarios por meta |
| 17 | Jefe / RH | descargar la evaluación de empleados en archivo | Botón "Descargar" que genera un archivo con los datos de la tabla | Generar reportes de evaluación para imprimir o enviar por correo | El archivo incluye: nombre del empleado, perfil, avance general (%), estado de evaluación. Se descarga con un clic |

---

## Sección 5: Administración RH — Competencias

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 18 | RH | administrar pilares (categorías de competencias) | Pantalla para crear, ver, editar y eliminar pilares | Mantener el catálogo de pilares de competencias de la empresa | Listado con opciones de crear, editar y eliminar con confirmación |
| 19 | RH | administrar competencias dentro de cada pilar | Pantalla para crear, ver, editar y eliminar competencias | Definir las competencias específicas por pilar | Cada competencia pertenece a un solo pilar. Se puede crear, editar y eliminar |
| 20 | RH | definir qué significa cada nivel de calificación por competencia | Pantalla para establecer criterios del 1 al 5 por competencia y perfil | Estandarizar la evaluación con descripciones claras de cada nivel | Escala fija 1 a 5. Los criterios cambian según el perfil del puesto |
| 21 | RH | definir el puntaje mínimo requerido por competencia | Pantalla para establecer niveles de aceptación por perfil | Saber si un empleado cumple el mínimo esperado en cada competencia | Cada perfil tiene sus propios niveles mínimos |

---

## Sección 6: Evaluación RH — Fin de Año

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 22 | RH | ver una lista de empleados con estado de su evaluación | Pantalla con tabla de empleados y su estado actual | Que RH dé seguimiento a las evaluaciones pendientes | Estados: pendiente, en progreso, completada |
| 23 | RH | comparar la autoevaluación del empleado con la evaluación de RH | Pantalla que muestra lado a lado ambas calificaciones | Detectar diferencias entre cómo se ve el empleado y cómo lo ve RH | Vista lado a lado de las calificaciones de ambas partes |

---

## Sección 7: Matriz 9×9 — Evaluación de Jefes

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 24 | Jefe | ubicar a mis colaboradores en una cuadrícula de desempeño versus potencial | Cuadrícula interactiva para posicionar a cada empleado | Visualizar el talento del equipo en un solo vistazo | La cuadrícula se puede usar interactivamente. No reemplaza la evaluación de competencias |
| 25 | Jefe / RH | navegar el organigrama de la empresa | Árbol visual para explorar la jerarquía organizacional | Navegar entre colaboradores según su posición en la empresa | Dos vistas disponibles: corporativa y retail |
| 26 | Jefe | consultar las competencias de mis colaboradores | Pantalla donde el jefe ve las competencias de sus subordinados | Que el jefe conozca las evaluaciones de su equipo | El jefe puede ver pero no calificar (esa tarea es de RH) |

---

## Sección 8: Mis Evaluados

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 27 | Jefe / Director | ver la lista de mis colaboradores directos | Pantalla con lista de personas a cargo | Acceso rápido a metas y evaluaciones del equipo | El jefe ve sus reportes directos, un director ve toda su área |

---

## Sección 9: Perfil de Usuario

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 28 | Todos los usuarios | ver mi página de perfil con datos personales y actividad reciente | Pantalla con nombre, correo, inicial del usuario e historial de actividad | Ver la identidad del usuario y su actividad en el sistema | Muestra datos del usuario, botón para cerrar sesión y actividad filtrada por perfil |

---

## Sección 10: Infraestructura del Sistema

| # | Quién | Historia | Cómo | Para qué | Criterios de aceptación |
|---|-------|----------|------|-----------|-------------------------|
| 29 | Administrador | un servidor que conecte el sistema con la base de datos | Servicio interno que procesa y guarda la información | Que los datos se guarden y consulten correctamente | El sistema no se cae inesperadamente, responde a verificación de salud y se comunica con la base de datos |
| 30 | Desarrollador | un entorno estandarizado para desarrollo | Configuración para que todos los desarrolladores usen el mismo ambiente | Evitar problemas entre diferentes computadoras | Se puede iniciar todo con un solo comando. La base de datos se configura automáticamente |

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

