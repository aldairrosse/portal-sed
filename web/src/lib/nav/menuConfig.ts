import type { EvaluationProfile, CyclePhase } from '$lib/types/evaluation';

export interface MenuItem {
	label: string;
	href: string;
	icon: string;
	profiles: EvaluationProfile[];
	phases?: CyclePhase[];
}

export const MENU_ITEMS: MenuItem[] = [
	{
		label: 'Inicio',
		href: '/',
		icon: 'Home',
		profiles: ['colaborador', 'jefe', 'vendedor', 'gerente-tienda', 'divisional', 'regional', 'director', 'rh']
	},
	{
		label: 'Metas',
		href: '/objetivos/asignacion',
		icon: 'Target',
		profiles: ['colaborador', 'jefe', 'vendedor', 'gerente-tienda', 'divisional', 'regional', 'director', 'rh']
	},
	{
		label: 'Mi evaluación',
		href: '/mi-evaluacion',
		icon: 'ClipboardCheck',
		profiles: ['colaborador', 'jefe', 'vendedor', 'gerente-tienda', 'divisional', 'regional', 'director']
	},
	{
		label: 'Mis evaluados',
		href: '/mis-evaluados',
		icon: 'Users',
		profiles: ['jefe', 'gerente-tienda', 'divisional', 'regional', 'director']
	},
	{
		label: 'Matriz 9×9',
		href: '/evaluacion/9x9',
		icon: 'Grid3x3',
		profiles: ['jefe', 'director', 'director-general']
	},
	{
		label: 'Jerarquía',
		href: '/evaluacion/9x9/jerarquia',
		icon: 'Network',
		profiles: ['director', 'director-general']
	},
	{
		label: 'Competencias',
		href: '/evaluacion/9x9/competencias',
		icon: 'Star',
		profiles: ['jefe', 'director', 'director-general']
	},
	{
		label: 'Pilares',
		href: '/rh/pilares',
		icon: 'Award',
		profiles: ['rh']
	},
	{
		label: 'Criterios escala',
		href: '/rh/criterios-escala',
		icon: 'Grid3x3',
		profiles: ['rh']
	},
	{
		label: 'Niveles aceptación',
		href: '/rh/niveles-aceptacion',
		icon: 'FileText',
		profiles: ['rh']
	},
	{
		label: 'Evaluaciones RH',
		href: '/rh/evaluaciones',
		icon: 'ClipboardList',
		profiles: ['rh']
	}
];

export function getVisibleMenuItems(profile: EvaluationProfile, phase: CyclePhase): MenuItem[] {
	return MENU_ITEMS.filter((item) => {
		if (!item.profiles.includes(profile)) return false;
		if (item.phases && !item.phases.includes(phase)) return false;
		return true;
	});
}
