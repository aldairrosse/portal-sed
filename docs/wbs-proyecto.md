# SED — WBS del Proyecto

> Prototipo completado. 30 tareas, 28 completadas.

---

## Resumen Ejecutivo

| Sección | Tareas | Estimación | Horas | Estado | Dependencias principales |
|---------|--------|------------|-------|--------|--------------------------|
| 1. Shell de la aplicación | 8 | 40h | 19h | ✅ Completado | Ninguna (punto de partida) |
| 2. Metas — Asignación de inicio de año | 1 | 48h | 4.5h | ✅ Completado | Sección 1 |
| 3. Metas — Avance de medio año | 2 | 16h | 6h | ✅ Completado | Sección 2 |
| 4. Mi evaluación — Autoevaluación fin de año | 3 | 32h | 7h | ✅ Completado | Secciones 2, 3, 5 |
| 5. Administración RH — Competencias | 1 | 36h | 2.5h | ✅ Completado | Sección 1 |
| 6. Evaluación RH — Fin de año | 2 | 16h | 5h | ✅ Completado | Secciones 4, 5 |
| 7. Matriz 9×9 — Evaluación de jefes | 3 | 28h | 7.5h | ✅ Completado | Secciones 1, 4 |
| 8. Mis evaluados | 1 | 8h | 3h | ✅ Completado | Secciones 1, 2, 3 |
| 9. Perfil de usuario | 1 | 6h | 3h | ✅ Completado | Sección 1 |
| 10. Backend — Infraestructura | 8 | 56h | 22.5h | ✅ Completado | Ninguna (paralelo) |
| 11. Pendientes finales | 3 | — | 34h | 🔲 Pendiente | Ver detalle abajo |
| **TOTAL** | **33** | — | **112.5h** | **28/33 completado** | |

---

## 1. Shell de la Aplicación

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 1.1 | Configuración inicial del entorno | — | 6h | Incluye framework de interfaz, sistema de estilos base y estructura de carpetas. |
| 1.2 | Vistas y rutas dinámicas según el menú | 1.1 | 3.5h | Cada opción del menú lleva a una pantalla. Las rutas se generan automáticamente según el perfil. |
| 1.3 | Esqueleto de secciones y menú | 1.1 | 2.5h | Pantalla principal con menú lateral, encabezado y área de contenido. |
| 1.4 | Agregar autenticación interna y externa | 1.2 | 2.5h | Login simulado para pruebas. Preparado para conectar con SSO empresarial. |
| 1.5 | Centralizar estilos y colores | 1.1 | 1.5h | Paleta de colores, tipografía y bordes definidos. Tema claro y oscuro. |
| 1.6 | Definición de pantallas principales | 1.1 | 1h | Bocetos iniciales de las 10 secciones del sistema. |
| 1.7 | Instalación de dependencias SDD | 1.1 | 1h | Herramientas de validación y control de calidad del código. |
| 1.8 | Definición de secciones y orden | 1.2 | 1h | Orden del menú y agrupación de pantallas por módulo. |

---

## 2. Metas — Asignación de Inicio de Año

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 2.1 | Sección de inicio de año | 1.2 | 4.5h | Pantalla completa: crear categorías, agregar metas, validar ponderación 100%, vista de solo lectura para jefes. |

---

## 3. Metas — Avance de Medio Año

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 3.1 | Sección de gráficos avance y resultados | 2.1 | 3.5h | Barras de progreso y semáforos por meta y por categoría. |
| 3.2 | Sección de medio año | 2.1 | 2.5h | Edición de metas existentes. No permite eliminar, solo ajustar valores. |

---

## 4. Mi Evaluación — Autoevaluación de Fin de Año

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 4.1 | Sección de fin de año | 2.1, 3.1 | 3h | Autoevaluación de competencias (escala 1-5) y cierre de metas con valor final. |
| 4.2 | Realizar ajustes en pantallas evaluación | 4.1 | 2.5h | Correcciones de formato, validaciones y experiencia de usuario. |
| 4.3 | Agregar botón y calificación inversa | 4.1 | 1.5h | Flujo donde el colaborador elige metas con objetivos a la baja. |

---

## 5. Administración RH — Competencias

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 5.1 | Sección admin de RH | 1.2 | 2.5h | Administración de pilares y competencias. Catálogo único para toda la empresa. |

---

## 6. Evaluación RH — Fin de Año

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 6.1 | Listado de empleados con estado de evaluación | 4.1, 5.1 | 3h | Tabla con filtro por nombre, perfil y estado. Incluye barra de búsqueda. |
| 6.2 | Comparación autoevaluación vs evaluación RH | 6.1 | 2h | Vista lado a lado para detectar diferencias entre la autoevaluación y la evaluación formal. |

---

## 7. Matriz 9×9 — Evaluación de Jefes

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 7.1 | Sección evaluación de jefes/usuario | 1.2 | 3h | Cuadrícula interactiva para ubicar colaboradores por desempeño y potencial. |
| 7.2 | Sección jerarquía de áreas | 1.2 | 2.5h | Árbol organizacional con dos vistas: corporativa y retail. |
| 7.3 | Ajustar gráfico NineBox | 7.1 | 2h | Grilla 9×9 visualización de puntuaciones. |

---

## 8. Mis Evaluados

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 8.1 | Sección y opciones de reportes | 1.2, 2.1, 3.1 | 3h | Lista de colaboradores a cargo con acceso directo a sus metas y evaluaciones. |

---

## 9. Perfil de Usuario

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 9.1 | Página de perfil con datos y actividad | 1.1 | 3h | Muestra nombre, correo, historial de actividad reciente y botón de cerrar sesión. |

---

## 10. Backend — Infraestructura

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 10.1 | Configurar APIs y persistencia de datos | — | 5h | Servidor que conecta el sistema con la base de datos. Guarda y consulta información. |
| 10.2 | Reemplazar datos mockup por api | 10.1 | 4h | Sustituir datos de prueba por datos reales del servidor progresivamente. |
| 10.3 | Configurar y ajustar pruebas de api y datos | 10.1 | 4h | Pruebas automáticas que verifican que los procesos responden correctamente. |
| 10.4 | Pruebas de flujo completo | 10.3 | 2.5h | Recorrido completo de principio a fin: login, crear meta, evaluar, cerrar ciclo. |
| 10.5 | Módulo de base de datos | 10.1 | 2h | Estructura de tablas, relaciones y reglas de información. |
| 10.6 | Pruebas generales de cambios SED | 10.3 | 2h | Verificación de que los cambios no rompan funcionalidad existente. |
| 10.7 | Revisar comentarios y feedback | — | 2h | Ajustes derivados de revisiones de calidad del código. |
| 10.8 | Desplegar prototipo dev en UAT | 10.1 | 1.5h | Publicar el sistema en ambiente de pruebas para validación. |

---

## 11. Pendientes Finales

| ID | Tarea | Dependencias | Horas | Comentarios |
|----|-------|--------------|-------|-------------|
| 11.1 | Pruebas de usuario (UAT) | 10.8 | 20h | 4-6 sesiones con usuarios reales (RH, jefes, colaboradores). Validar flujos completos: metas, avance, autoevaluación, evaluación RH, matriz 9×9. Incluye preparación y reporte de incidencias. |
| 11.2 | Ajustes post-pruebas | 11.1 | 8h | Corrección de incidencias priorizadas: bloqueadores (flujo roto), mayores (funcionalidad incorrecta), menores (mejoras visuales). |
| 11.3 | Despliegue a productivo y estabilización | 11.2 | 6h | Migración a producción: credenciales, certificados, dominio. Verificación post-despliegue. Monitoreo de estabilidad las primeras 48h. Plan de retroceso documentado. |

---

## Diagrama de Dependencias

```
Sección 1 (Shell)
  ├─→ Sección 2 (Metas inicio año)
  │     ├─→ Sección 3 (Avance medio año)
  │     │     └─→ Sección 8 (Mis evaluados)
  │     └─→ Sección 4 (Autoevaluación)
  │           └─→ Sección 6 (Evaluación RH)
  ├─→ Sección 5 (Admin RH)
  │     ├─→ Sección 4
  │     └─→ Sección 6
  ├─→ Sección 7 (Matriz 9×9)
  └─→ Sección 9 (Perfil)

Sección 10 (Backend) → PARALELO a todo el frontend

Sección 11 (Pendientes) → POST-PROTOTIPO
  11.1 → 11.2 → 11.3
```

---

## Archivos de Referencia

| Archivo | Rol en WBS |
|---------|------------|
| Lista maestra de requisitos | Definición de los 33 requerimientos originales |
| Changes del proyecto | Cambios aplicados durante el prototipo |
| Guías del proyecto | Decisiones transversales |
| Datos de prueba | Fixtures por módulo |
| Seeders del backend | Datos iniciales para la base de datos |
