import type { EvaluationProfile, CyclePhase } from '$lib/types/evaluation';
import { MANAGER_ROLES } from '$lib/stores/roleStore.svelte';

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
		profiles: [
			'colaborador',
			'jefe',
			'vendedor',
			'gerente',
			'coordinador',
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'director-general',
			'rh',
		],
	},
	{
		label: 'Metas',
		href: '/objetivos/asignacion',
		icon: 'Target',
		profiles: [
			'colaborador',
			'jefe',
			'vendedor',
			'gerente',
			'coordinador',
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'rh',
		],
	},
	{
		label: 'Objetivos globales',
		href: '/objetivos/globales',
		icon: 'Target',
		profiles: ['rh'],
	},
	{
		label: 'Metas compartidas',
		href: '/objetivos/compartidas',
		icon: 'Users',
		profiles: [...MANAGER_ROLES, 'gerente-tienda', 'divisional', 'regional', 'director'],
	},
	{
		label: 'Mi evaluación',
		href: '/mi-evaluacion',
		icon: 'ClipboardCheck',
		profiles: [
			'colaborador',
			'jefe',
			'vendedor',
			'gerente',
			'coordinador',
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'rh',
		],
	},
	{
		label: 'Mis evaluados',
		href: '/mis-evaluados',
		icon: 'Users',
		profiles: [
			...MANAGER_ROLES,
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'rh',
		],
	},
	{
		label: 'Matriz 9-Box',
		href: '/evaluacion/9x9',
		icon: 'Grid3x3',
		profiles: [...MANAGER_ROLES, 'director', 'director-general', 'rh'],
	},
	{
		label: 'Jerarquía',
		href: '/evaluacion/9x9/jerarquia',
		icon: 'Network',
		profiles: ['director', 'director-general'],
	},
	{
		label: 'Pilares',
		href: '/rh/pilares',
		icon: 'Award',
		profiles: ['rh'],
	},
	{
		label: 'Criterios escala',
		href: '/rh/criterios-escala',
		icon: 'Grid3x3',
		profiles: ['rh'],
	},
	{
		label: 'Niveles aceptación',
		href: '/rh/niveles-aceptacion',
		icon: 'FileText',
		profiles: ['rh'],
	},
	{
		label: 'Ciclos',
		href: '/rh/ciclos',
		icon: 'Calendar',
		profiles: ['rh'],
	},
	{
		label: 'Evaluaciones',
		href: '/rh/evaluaciones',
		icon: 'ClipboardList',
		profiles: ['rh'],
	},
	{
		label: 'Jerarquía',
		href: '/rh/jerarquia',
		icon: 'Network',
		profiles: ['rh'],
	},
];

export function getVisibleMenuItems(
	profile: EvaluationProfile,
	phase: CyclePhase,
): MenuItem[] {
	return MENU_ITEMS.filter((item) => {
		if (!item.profiles.includes(profile)) return false;
		if (item.phases && !item.phases.includes(phase)) return false;
		return true;
	});
}

// Central route guard (CLIENT-ONLY UX hint — not a security boundary).
// Real enforcement lives in the backend (api rbac.go); this only hides
// pages and blocks direct-URL navigation in the UI.
// Single source of truth is MENU_ITEMS, matched by longest href-prefix so
// subroutes inherit (e.g. /evaluacion/9x9/competencias/123 → /evaluacion/9x9,
// /objetivos/asignacion/biblioteca → /objetivos/asignacion).
// Paths with no rule stay open (e.g. /perfil, /ajustes) to avoid blocking
// pages without a menu entry.
// Minimal explicit rules for guarded paths without a menu entry:
const ROUTE_RULES: Pick<MenuItem, 'href' | 'profiles'>[] = [
	{
		href: '/evaluacion/9x9/competencias',
		// Explicit longest-prefix override: lets every profile reach the page
		// via direct URL (layout guard passes) — the page itself enforces
		// ownership (self) vs manager/rh, so non-owners still see 403.
		profiles: [
			'colaborador',
			'jefe',
			'vendedor',
			'gerente',
			'coordinador',
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'director-general',
			'rh',
		],
	},
	{
		href: '/objetivos/avance',
		profiles: [
			'colaborador',
			'jefe',
			'vendedor',
			'gerente',
			'coordinador',
			'gerente-tienda',
			'divisional',
			'regional',
			'director',
			'rh',
		],
	},
];
export function canAccess(
	profile: EvaluationProfile | string | null | undefined,
	href: string,
): boolean {
	if (!profile) return true;
	let path = href.split('?')[0].split('#')[0];
	if (path.length > 1 && path.endsWith('/')) path = path.slice(0, -1);
	let best: Pick<MenuItem, 'href' | 'profiles'> | null = null;
	for (const rule of [...ROUTE_RULES, ...MENU_ITEMS]) {
		if (rule.href === '/') {
			if (path === '/') best = rule;
			continue;
		}
		if (path === rule.href || path.startsWith(rule.href + '/')) {
			if (!best || rule.href.length > best.href.length) best = rule;
		}
	}
	if (!best) return true;
	return (best.profiles as readonly string[]).includes(profile);
}
