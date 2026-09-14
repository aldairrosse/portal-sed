import type { EvaluationProfile } from '$lib/types/evaluation';

// ponytail: simple array includes — no map/registry needed for 3 roles.
// Central manager check: jefe, gerente y coordinador validan al mismo nivel.
export const MANAGER_ROLES = ['jefe', 'gerente', 'coordinador'] as const;

export type ManagerRole = (typeof MANAGER_ROLES)[number];

function normalizeRole(roleOrProfile?: string | null): string {
	return (roleOrProfile ?? '').trim().toLowerCase();
}

export function isManager(roleOrProfile?: string | null): boolean {
	return (MANAGER_ROLES as readonly string[]).includes(
		normalizeRole(roleOrProfile),
	);
}

// Alias case-insensitive (trim + lower) para nombres de perfil libres.
export function isManagerProfile(profileName?: string | null): boolean {
	return isManager(profileName);
}

/** @deprecated Usa isManager() o isManagerProfile(). */
export function isJefe(roleOrProfile?: string | null): boolean {
	return normalizeRole(roleOrProfile) === 'jefe';
}

export function isManagerEvaluationProfile(
	profile?: EvaluationProfile | string | null,
): boolean {
	return isManager(profile);
}
