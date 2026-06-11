import { client } from './client';
import type { EvaluationProfile } from '$lib/types/evaluation';

export interface AuthUser {
	employeeId: string;
	email: string;
	name: string;
	profileId: EvaluationProfile;
	organizationId: string;
}

const FIXTURE_USER: AuthUser = {
	employeeId: '00000000-0000-0000-0000-000000000001',
	email: 'dev@sed.local',
	name: 'Usuario Desarrollo',
	profileId: 'colaborador',
	organizationId: '00000000-0000-0000-0000-000000000001'
};

let user = $state<AuthUser | null>(null);
let loading = $state(true);
let error = $state<string | null>(null);

export async function ensureSession(): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		user = { ...FIXTURE_USER };
		loading = false;
		return;
	}

	try {
		const { data, error: apiError } = await (client as any).GET('/auth/me');
		if (apiError) {
			throw new Error(typeof apiError === 'string' ? apiError : 'Error de autenticación');
		}
		const raw = data as {
			employee: { id: string; email?: string; first_name?: string; last_name?: string };
			role: string;
			profile?: { id?: string; name?: string };
			organization_id?: string;
		};
		user = {
			employeeId: raw.employee.id ?? '',
			email: raw.employee.email ?? '',
			name: [raw.employee.first_name, raw.employee.last_name].filter(Boolean).join(' ') || 'Usuario',
			profileId: raw.role as EvaluationProfile,
			organizationId: raw.organization_id ?? raw.employee.id ?? ''
		};
	} catch (e) {
		error = e instanceof Error ? e.message : 'Error al cargar sesión';
		user = null;
	} finally {
		loading = false;
	}
}

export function getSession(): { user: AuthUser | null; loading: boolean; error: string | null } {
	return { user, loading, error };
}

export async function devLogin(email: string): Promise<void> {
	loading = true;
	error = null;

	if (import.meta.env.DEV && !import.meta.env.VITE_USE_API) {
		const userMap: Record<string, AuthUser> = {
			'dev-rh@empresa.com': {
				employeeId: '00000000-0000-0000-0000-000000000001',
				email: 'dev-rh@empresa.com',
				name: 'Frankil Perez',
				profileId: 'rh',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'dev-jefe@empresa.com': {
				employeeId: '00000000-0000-0000-0000-000000000002',
				email: 'dev-jefe@empresa.com',
				name: 'Juan Carlos',
				profileId: 'jefe',
				organizationId: '00000000-0000-0000-0000-000000000001'
			},
			'dev-colaborador@empresa.com': {
				employeeId: '00000000-0000-0000-0000-000000000003',
				email: 'dev-colaborador@empresa.com',
				name: 'Maria Lopez',
				profileId: 'colaborador',
				organizationId: '00000000-0000-0000-0000-000000000001'
			}
		};

		const match = userMap[email];
		if (!match) {
			error = 'Usuario demo no válido';
			loading = false;
			throw new Error('Usuario demo no válido');
		}

		user = { ...match };
		loading = false;
		return;
	}

	const { error: apiError } = await (client as any).POST('/auth/dev-login', {
		body: { email }
	});
	if (apiError) {
		error = typeof apiError === 'string' ? apiError : 'Error al iniciar sesión';
		loading = false;
		throw new Error(error);
	}
	await ensureSession();
}

export async function logout(): Promise<void> {
	try {
		if (!(import.meta.env.DEV && !import.meta.env.VITE_USE_API)) {
			await (client as any).POST('/auth/logout');
		}
	} catch {
		// Even if backend fails, clear local state
	}
	user = null;
	window.location.href = '/login';
}
