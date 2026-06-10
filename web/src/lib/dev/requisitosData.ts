export interface Entregable {
	item: string;
	archivos: string[];
}

export interface Requisito {
	id: number;
	seccion: string;
	requerimiento: string;
	entregables: Entregable[];
	notas: string;
}

export interface SeccionData {
	seccion: string;
	descripcion: string;
	requisitos: Requisito[];
}

const DATA: SeccionData[] = [
	{
		seccion: '1. Inicio — Shell de la aplicación',
		descripcion:
			'Estructura base del portal: menú lateral, encabezado responsive, sistema de temas claro/oscuro, y barra de herramientas para desarrolladores.',
		requisitos: [
			{
				id: 1,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento:
					'Layout principal con menú lateral (sidebar) que se adapta a dispositivos móviles y escritorio.',
				entregables: [
					{
						item: 'AppShell con layout tipo "drawer": en escritorio el menú siempre visible, en móvil se oculta y se abre con un botón tipo hamburguesa.',
						archivos: [
							'web/src/lib/components/AppShell.svelte',
							'web/src/lib/components/Sidebar.svelte'
						]
					}
				],
				notas: 'Responsive: lg:drawer-open. El menú cambia según el perfil de usuario seleccionado en la barra dev.'
			},
			{
				id: 2,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento:
					'Menú de navegación que muestra solo las opciones permitidas para cada perfil y fase del ciclo.',
				entregables: [
					{
						item: 'Archivo de configuración menuConfig.ts con todas las rutas, íconos, perfiles y fases permitidas por cada una.',
						archivos: ['web/src/lib/nav/menuConfig.ts']
					}
				],
				notas: 'Soporta 9 perfiles: colaborador, jefe, vendedor, gerente-tienda, divisional, regional, director, director-general, rh.'
			},
			{
				id: 3,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento:
					'Barra de herramientas para desarrolladores (dev toolbar) que permite cambiar de perfil y fase del ciclo sin necesidad de iniciar sesión.',
				entregables: [
					{
						item: 'DevToolbar flotante en la esquina inferior derecha, visible solo en entorno de desarrollo. Se oculta/muestra con Ctrl+Shift+D.',
						archivos: [
							'web/src/lib/components/DevToolbar.svelte',
							'web/src/lib/stores/devContext.svelte.ts'
						]
					}
				],
				notas: 'Al cambiar el perfil, toda la app se comporta como si ese usuario hubiera iniciado sesión. Datos mock, sin autenticación real.'
			},
			{
				id: 4,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento: 'Sistema de temas claro y oscuro con DaisyUI.',
				entregables: [
					{
						item: 'Variables CSS (design tokens) para colores base, radios de borde y tipografía. Soporte nativo de DaisyUI para cambiar de tema.',
						archivos: ['web/src/app.css']
					}
				],
				notas: 'Los logos en la barra lateral también se adaptan al tema (logo_black.png para tema claro, logo_white.png para oscuro).'
			},
			{
				id: 5,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento:
					'Páginas de estado para cuando algo sale mal: error, vacío (sin datos), y carga (skeleton).',
				entregables: [
					{
						item: 'Componentes reutilizables: ErrorState, EmptyState, ForbiddenState, PageSkeleton.',
						archivos: [
							'web/src/lib/components/ui/ErrorState.svelte',
							'web/src/lib/components/ui/EmptyState.svelte',
							'web/src/lib/components/ui/ForbiddenState.svelte',
							'web/src/lib/components/ui/PageSkeleton.svelte'
						]
					}
				],
				notas: 'Página de error global en /+error.svelte.'
			},
			{
				id: 6,
				seccion: '1. Inicio — Shell de la aplicación',
				requerimiento:
					'Perfiles de usuario simulados (fixtures) con nombres, correos y datos de prueba.',
				entregables: [
					{
						item: 'Archivo profileUsers.ts con 9 perfiles completos (colaborador, jefe, vendedor, etc.) y nombres ficticios.',
						archivos: ['web/src/lib/dev/profileUsers.ts']
					}
				],
				notas: 'El perfil "colaborador" usa "María López García". Los datos se consumen desde archivos JSON en src/lib/fixtures/.'
			}
		]
	},
	{
		seccion: '2. Metas — Asignación de inicio de año',
		descripcion:
			'Pantalla para que el empleado defina sus metas del año, agrupadas en categorías personalizadas, con indicadores KPI y validación de ponderación. Los jefes pueden ver y solicitar cambios, pero no editar directamente.',
		requisitos: [
			{
				id: 7,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento:
					'El empleado puede crear categorías de metas personalizadas (ej. "Ventas", "Servicio", "Proyectos").',
				entregables: [
					{
						item: 'CategoryCard para crear/editar/eliminar categorías. Cada categoría agrupa metas relacionadas.',
						archivos: [
							'web/src/lib/components/goals/CategoryCard.svelte',
							'web/src/lib/fixtures/goals/goal-categories.json'
						]
					}
				],
				notas: 'Las categorías de metas son independientes de los pilares de competencias. Cada usuario define las suyas.'
			},
			{
				id: 8,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento:
					'El empleado puede crear metas con nombre, descripción, unidad (porcentaje o moneda), valor esperado y peso dentro de la categoría.',
				entregables: [
					{
						item: 'GoalRow para cada meta con edición inline. WeightIndicator para el peso porcentual.',
						archivos: [
							'web/src/lib/components/goals/GoalRow.svelte',
							'web/src/lib/components/goals/KpiBadge.svelte',
							'web/src/lib/fixtures/goals/goals.json'
						]
					}
				],
				notas: 'Todas las metas tienen unidad (%, $ o numérico) y un peso que suma 100% dentro de su categoría.'
			},
			{
				id: 9,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento:
					'Doble validación de ponderación 100%: las categorías suman 100% y las metas dentro de cada categoría suman 100%.',
				entregables: [
					{
						item: 'Validación matemática en goalValidation.ts. Indicadores visuales de progreso en ProgressIndicator.',
						archivos: [
							'web/src/lib/components/goals/goalValidation.ts',
							'web/src/lib/components/goals/ProgressIndicator.svelte'
						]
					}
				],
				notas: 'No se puede guardar si no suma 100% en ambos niveles. Es una regla de negocio obligatoria.'
			},
			{
				id: 10,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento:
					'Los jefes pueden ver las metas de sus colaboradores en modo solo lectura y solicitar cambios (ajustar KPIs y ponderaciones).',
				entregables: [
					{
						item: 'ReadOnlyBanner que indica modo visualización. RequestChangeModal para enviar solicitud de ajuste.',
						archivos: [
							'web/src/lib/components/goals/ReadOnlyBanner.svelte',
							'web/src/lib/components/goals/RequestChangeModal.svelte'
						]
					}
				],
				notas: 'El jefe NO puede borrar ni agregar metas, solo solicitar cambios. El dueño de la meta decide si acepta.'
			},
			{
				id: 11,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento: 'Biblioteca de KPIs: catálogo de indicadores predefinidos que se pueden asignar a las metas.',
				entregables: [
					{
						item: 'Ruta /objetivos/asignacion/biblioteca con listado de KPIs. KpiFormModal para crear/editar KPIs.',
						archivos: [
							'web/src/routes/objetivos/asignacion/biblioteca/+page.svelte',
							'web/src/lib/components/goals/KpiFormModal.svelte',
							'web/src/lib/fixtures/goals/kpis.json'
						]
					}
				],
				notas: 'KPIs pueden ser numéricos, porcentaje o moneda. Un KPI puede vincularse a una o más metas.'
			},
			{
				id: 12,
				seccion: '2. Metas — Asignación de inicio de año',
				requerimiento:
					'Ver y solicitar cambios de metas a colaboradores desde el selector de persona. Las metas del usuario actual se muestran con el sufijo (Yo).',
				entregables: [
					{
						item: 'AssigneePicker con selector de empleado que marca al usuario actual con "(yo)". ReadOnlyBanner + RequestChangeModal para vista y solicitud de cambios a colaboradores.',
						archivos: [
							'web/src/lib/components/goals/AssigneePicker.svelte',
							'web/src/lib/components/goals/ReadOnlyBanner.svelte',
							'web/src/lib/components/goals/RequestChangeModal.svelte'
						]
					}
				],
				notas: 'El jefe selecciona un colaborador en el picker, ve sus metas en solo lectura y puede solicitar cambios. Sus propias metas aparecen con "(Yo)". Usa org-tree.json para la jerarquía.'
			}
		]
	},
	{
		seccion: '3. Metas — Avance de medio año',
		descripcion:
			'Pantalla para registrar avances de metas a mitad del ciclo. Se pueden editar metas existentes pero no eliminarlas.',
		requisitos: [
			{
				id: 13,
				seccion: '3. Metas — Avance de medio año',
				requerimiento: 'Edición de metas existentes en la fase de medio año: actualizar valor esperado y registrar avance real.',
				entregables: [
					{
						item: 'Ruta /objetivos/avance con vista de metas y formulario de avance. Las metas mantienen su estructura de inicio de año.',
						archivos: ['web/src/routes/objetivos/avance/+page.svelte']
					}
				],
				notas: 'En medio año NO se pueden eliminar metas, solo ajustar campos permitidos y registrar avances.'
			},
			{
				id: 14,
				seccion: '3. Metas — Avance de medio año',
				requerimiento: 'Registro de avance porcentual o por valor para cada meta.',
				entregables: [
					{
						item: 'Input de avance dentro de cada GoalRow. Indicador visual de progreso por meta y por categoría.',
						archivos: ['web/src/lib/components/goals/ProgressIndicator.svelte']
					}
				],
				notas: 'El avance se muestra como semáforo o barra de progreso.'
			}
		]
	},
	{
		seccion: '4. Mi evaluación — Autoevaluación de fin de año',
		descripcion:
			'Pantalla donde el empleado califica sus propias competencias en escala 1-5 y cierra sus metas. Incluye gráfico radar para visualizar resultados.',
		requisitos: [
			{
				id: 15,
				seccion: '4. Mi evaluación — Autoevaluación de fin de año',
				requerimiento:
					'Autoevaluación del empleado: calificar competencias en escala 1-5 con criterios por perfil.',
				entregables: [
					{
						item: 'Ruta /mi-evaluacion con vista completa. CompetencyRatingCard para calificar cada competencia. ScaleRatingSelector para la escala.',
						archivos: [
							'web/src/routes/mi-evaluacion/+page.svelte',
							'web/src/lib/components/evaluation/CompetencyRatingCard.svelte',
							'web/src/lib/components/evaluation/ScaleRatingSelector.svelte',
							'web/src/lib/fixtures/evaluations/self-evaluations.json'
						]
					}
				],
				notas: 'La escala es 1-5 con criterios específicos que cambian según el perfil de evaluación.'
			},
			{
				id: 16,
				seccion: '4. Mi evaluación — Autoevaluación de fin de año',
				requerimiento: 'Visualización en gráfico radar de las competencias evaluadas.',
				entregables: [
					{
						item: 'RadarChart con Chart.js, conmutación entre vista tabla y radar mediante tabs. Se adapta al tema claro/oscuro.',
						archivos: [
							'web/src/lib/components/evaluation/RadarChart.svelte',
							'web/src/lib/components/evaluation/CompetencyNetworkView.svelte'
						]
					}
				],
				notas: 'Usa Chart.js con radar controller. Los colores se adaptan automáticamente al tema activo.'
			},
			{
				id: 17,
				seccion: '4. Mi evaluación — Autoevaluación de fin de año',
				requerimiento: 'Cierre de metas: registrar cumplimiento final de cada meta.',
				entregables: [
					{
						item: 'GoalClosureCard para cerrar cada meta con su valor final y comentarios.',
						archivos: [
							'web/src/lib/components/evaluation/GoalClosureCard.svelte',
							'web/src/lib/fixtures/evaluations/goal-closures.json'
						]
					}
				],
				notas: 'El cierre de metas ocurre en la fase de fin de año, junto con la autoevaluación de competencias.'
			},
			{
				id: 18,
				seccion: '4. Mi evaluación — Autoevaluación de fin de año',
				requerimiento: 'Exportar tabla de evaluación a archivo CSV.',
				entregables: [
					{
						item: 'Utilidad de exportación export.ts con botón de descarga CSV en EmployeeEvaluationTable.',
						archivos: ['web/src/lib/utils/export.ts']
					}
				],
				notas: 'Exporta los datos visibles en la tabla sin formato Niños/pesos.'
			}
		]
	},
	{
		seccion: '5. Administración RH — Competencias',
		descripcion:
			'Panel para que Recursos Humanos administre el catálogo único de pilares, competencias, criterios de escala y niveles de aceptación por perfil.',
		requisitos: [
			{
				id: 19,
				seccion: '5. Administración RH — Competencias',
				requerimiento: 'CRUD completo de pilares (categorías de competencias).',
				entregables: [
					{
						item: 'Ruta /rh/pilares con listado, formulario de creación/edición (PillarFormModal) y confirmación de eliminación (ConfirmDeleteModal).',
						archivos: [
							'web/src/routes/rh/pilares/+page.svelte',
							'web/src/routes/rh/pilares/[id]/competencias/+page.svelte',
							'web/src/lib/components/competency/PillarFormModal.svelte',
							'web/src/lib/components/competency/ConfirmDeleteModal.svelte',
							'web/src/lib/fixtures/competency/pillars.json'
						]
					}
				],
				notas: 'TODOS los perfiles usan el mismo catálogo de pilares. Es único para toda la empresa.'
			},
			{
				id: 20,
				seccion: '5. Administración RH — Competencias',
				requerimiento: 'CRUD de competencias dentro de cada pilar.',
				entregables: [
					{
						item: 'Vista de competencias por pilar con formulario de creación/edición (CompetencyFormModal).',
						archivos: [
							'web/src/lib/components/competency/CompetencyFormModal.svelte',
							'web/src/lib/fixtures/competency/competencies.json'
						]
					}
				],
				notas: 'Cada competencia pertenece a un solo pilar.'
			},
			{
				id: 21,
				seccion: '5. Administración RH — Competencias',
				requerimiento:
					'Matriz de criterios de escala (1-5) por competencia, que varían según el perfil de evaluación.',
				entregables: [
					{
						item: 'ScaleCriteriaMatrix con edición de criterios por nivel. ScaleCriterionModal para detalle. Rutas /rh/criterios-escala.',
						archivos: [
							'web/src/lib/components/competency/ScaleCriteriaMatrix.svelte',
							'web/src/lib/components/competency/ScaleCriterionModal.svelte',
							'web/src/routes/rh/criterios-escala/+page.svelte',
							'web/src/lib/fixtures/competency/scale-criteria.json'
						]
					}
				],
				notas: 'La escala 1-5 es fija, pero los criterios descriptivos cambian por perfil (ej. lo que significa "3" para un vendedor vs un director).'
			},
			{
				id: 22,
				seccion: '5. Administración RH — Competencias',
				requerimiento:
					'Niveles de aceptación: puntaje mínimo requerido en cada competencia según el perfil.',
				entregables: [
					{
						item: 'AcceptanceLevelEditor para configurar niveles. AcceptanceLevelSummaryModal para resumen. Rutas /rh/niveles-aceptacion.',
						archivos: [
							'web/src/lib/components/competency/AcceptanceLevelEditor.svelte',
							'web/src/lib/components/competency/AcceptanceLevelSummaryModal.svelte',
							'web/src/lib/components/competency/AcceptanceLevelEditor.svelte',
							'web/src/routes/rh/niveles-aceptacion/+page.svelte',
							'web/src/lib/fixtures/competency/acceptance-levels.json',
							'web/src/lib/fixtures/competency/competency-acceptance-levels.json'
						]
					}
				],
				notas: 'Cada perfil de evaluación tiene sus propios niveles de aceptación.'
			}
		]
	},
	{
		seccion: '6. Evaluación RH — Fin de año',
		descripcion:
			'Panel para que Recursos Humanos realice la evaluación formal de cada empleado al cierre del ciclo anual.',
		requisitos: [
			{
				id: 23,
				seccion: '6. Evaluación RH — Fin de año',
				requerimiento:
					'Listado de todos los empleados con su estado de evaluación (pendiente, en progreso, completada).',
				entregables: [
					{
						item: 'Ruta /rh/evaluaciones con tabla de empleados. EvaluationStatusBadge para el estado. EmployeeEvaluationTable con datos.',
						archivos: [
							'web/src/routes/rh/evaluaciones/+page.svelte',
							'web/src/lib/components/evaluation/EvaluationStatusBadge.svelte',
							'web/src/lib/components/evaluation/EmployeeEvaluationTable.svelte',
							'web/src/lib/components/evaluation/EmployeeEvaluationDetail.svelte',
							'web/src/lib/fixtures/evaluations/rh-evaluations.json'
						]
					}
				],
				notas: 'RH puede ver el detalle de cada evaluación y comparar resultados.'
			},
			{
				id: 24,
				seccion: '6. Evaluación RH — Fin de año',
				requerimiento: 'Comparación entre autoevaluación del empleado y evaluación de RH.',
				entregables: [
					{
						item: 'ComparisonTable para mostrar lado a lado la calificación del empleado vs la de RH.',
						archivos: ['web/src/lib/components/evaluation/ComparisonTable.svelte']
					}
				],
				notas: 'Útil para detectar brechas entre cómo se ve el empleado y cómo lo ve RH.'
			}
		]
	},
	{
		seccion: '7. Matriz 9×9 — Evaluación de jefes',
		descripcion:
			'Herramienta para que los jefes evalúen a sus colaboradores en dos ejes: desempeño y potencial, ubicándolos en una cuadrícula 9×9.',
		requisitos: [
			{
				id: 25,
				seccion: '7. Matriz 9×9 — Evaluación de jefes',
				requerimiento:
					'Cuadrícula 9×9 donde el jefe ubica a cada colaborador según su desempeño y potencial.',
				entregables: [
					{
						item: 'NineBoxMatrix con grid interactivo. NineBoxSliders para ajustar puntuaciones. NineBoxEntryCard con detalle por persona.',
						archivos: [
							'web/src/lib/components/nine-box/NineBoxMatrix.svelte',
							'web/src/lib/components/nine-box/NineBoxSliders.svelte',
							'web/src/lib/components/nine-box/NineBoxEntryCard.svelte',
							'web/src/lib/fixtures/nine-box/matrix-entries.json',
							'web/src/lib/fixtures/nine-box/quadrant-definitions.json',
							'web/src/lib/fixtures/nine-box/scale-definitions.json'
						]
					}
				],
				notas: 'NO reemplaza la evaluación RH de competencias. El 9×9 es SOLO para potencial/desempeño.'
			},
			{
				id: 26,
				seccion: '7. Matriz 9×9 — Evaluación de jefes',
				requerimiento: 'Árbol de jerarquía organizacional para navegar entre colaboradores.',
				entregables: [
					{
						item: 'OrgHierarchyTree con nodos expandibles. TreeNode para cada persona. Rutas /evaluacion/9x9/jerarquia.',
						archivos: [
							'web/src/lib/components/org-hierarchy/OrgHierarchyTree.svelte',
							'web/src/lib/components/org-hierarchy/TreeNode.svelte',
							'web/src/routes/evaluacion/9x9/jerarquia/+page.svelte',
							'web/src/lib/fixtures/org-hierarchy/org-tree.json'
						]
					}
				],
				notas: 'Jerarquía dual: corporativa (colaborador→jefe→director→director general) y retail (vendedor→gerente tienda→divisional→regional).'
			},
			{
				id: 27,
				seccion: '7. Matriz 9×9 — Evaluación de jefes',
				requerimiento:
					'Vista de competencias del colaborador desde la perspectiva del jefe.',
				entregables: [
					{
						item: 'Ruta /evaluacion/9x9/competencias con detalle por empleado.',
						archivos: [
							'web/src/routes/evaluacion/9x9/competencias/+page.svelte',
							'web/src/routes/evaluacion/9x9/competencias/[employeeId]/+page.svelte'
						]
					}
				],
				notas: 'El jefe puede ver las competencias de su colaborador pero no calificarlas (eso lo hace RH).'
			}
		]
	},
	{
		seccion: '8. Mis evaluados',
		descripcion:
			'Lista de colaboradores a cargo del usuario actual, con acceso rápido a sus metas y evaluaciones.',
		requisitos: [
			{
				id: 28,
				seccion: '8. Mis evaluados',
				requerimiento:
					'Lista de colaboradores directos que el jefe/gerente/director tiene a su cargo.',
				entregables: [
					{
						item: 'Ruta /mis-evaluados con tabla de colaboradores y enlaces a sus metas.',
						archivos: ['web/src/routes/mis-evaluados/+page.svelte']
					}
				],
				notas: 'El alcance de "mis evaluados" depende del perfil: un jefe ve a sus colaboradores directos; un director ve a todos los de su área.'
			}
		]
	},
	{
		seccion: '9. Perfil de usuario',
		descripcion:
			'Página de perfil con información del usuario activo, historial de actividad y cierre de sesión simulado.',
		requisitos: [
			{
				id: 29,
				seccion: '9. Perfil de usuario',
				requerimiento:
					'Página de perfil con datos del usuario, cierre de sesión y timeline de actividad reciente.',
				entregables: [
					{
						item: 'Ruta /perfil con nombre, correo, inicial del usuario, botón de cerrar sesión y lista de actividad reciente.',
						archivos: [
							'web/src/routes/perfil/+page.svelte',
							'web/src/lib/fixtures/activity/activity-logs.json'
						]
					}
				],
				notas: 'El cierre de sesión es simulado (no hay auth real). La actividad se filtra por perfil.'
			}
		]
	},
	{
		seccion: '10. Backend — Infraestructura (en progreso)',
		descripcion:
			'Servidor backend en Go con Docker, base de datos PostgreSQL y ORM Ent. En progreso, conectando las APIs para reemplazar los datos mock del frontend.',
		requisitos: [
			{
				id: 30,
				seccion: '10. Backend — Infraestructura (en progreso)',
				requerimiento: 'Servidor HTTP en Go con router Chi listo para servir APIs REST.',
				entregables: [
					{
						item: 'Servidor con graceful shutdown, health check, CORS, y compresión. Conexión a PostgreSQL via Ent.',
						archivos: ['api/cmd/server/main.go', 'api/go.mod', 'api/go.sum']
					}
				],
				notas: 'Usa Chi como router, Ent como ORM. Las migraciones se ejecutan automáticamente al iniciar.'
			},
			{
				id: 31,
				seccion: '10. Backend — Infraestructura (en progreso)',
				requerimiento: 'Contenedores Docker para desarrollo y pruebas.',
				entregables: [
					{
						item: 'Dockerfile multi-stage para producción. docker-compose.yml con PostgreSQL + servidor. docker-compose.test.yml con base de datos de prueba.',
						archivos: [
							'api/Dockerfile',
							'docker-compose.yml',
							'docker-compose.test.yml'
						]
					}
				],
				notas: 'Usuario no-root en Docker. Variables de entorno para configuración.'
			},
			{
				id: 32,
				seccion: '10. Backend — Infraestructura (en progreso)',
				requerimiento: 'Seeders con datos iniciales que coinciden con los fixtures del frontend.',
				entregables: [
					{
						item: 'Seeders para poblar la base de datos con datos de prueba consistentes con los JSON del frontend.',
						archivos: ['api/internal/seed/']
					}
				],
				notas: 'Los seeders replican los datos de web/src/lib/fixtures/ para que frontend y backend usen la misma información.'
			},
			{
				id: 33,
				seccion: '10. Backend — Infraestructura (en progreso)',
				requerimiento: 'Pruebas de integración que verifican que todas las rutas responden sin error 500.',
				entregables: [
					{
						item: 'Suite de pruebas con 70+ pruebas que conectan a PostgreSQL real, ejecutan migraciones y verifican rutas.',
						archivos: ['api/integration/']
					}
				],
				notas: 'Requiere Docker para la base de datos de prueba. Incluye verificación de endpoints CRUD principales.'
			}
		]
	}
];

export default DATA;
